package crawler

import (
	"fmt"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/corpix/uarand"
	// "github.com/gocolly/colly/v2"
	// "github.com/gocolly/colly/v2/proxy"
	// "github.com/gocolly/colly/v2/queue"
	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"

        "github.com/lucmichalski/finance-dataset/pkg/colly"
        "github.com/lucmichalski/finance-dataset/pkg/colly/proxy"
        "github.com/lucmichalski/finance-dataset/pkg/colly/queue"

	"github.com/lucmichalski/finance-dataset/pkg/articletext"
	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/models"
	"github.com/lucmichalski/finance-dataset/pkg/sitemap"
)

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
	)

	// Rotate two socks5 proxies
	rp, err := proxy.RoundRobinProxySwitcher("http://localhost:8119")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	c.DisableCookies()

	// create a request queue with 1 consumer thread until we solve the multi-threadin of the darknet model
	q, _ := queue.New(
		cfg.ConsumerThreads,
		&queue.InMemoryQueueStorage{
			MaxSize: cfg.QueueMaxSize,
		},
	)

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//sitemap/loc", func(e *colly.XMLElement) {
		if strings.Contains(e.Text, "/articles/") || strings.Contains(e.Text, "/news/") {
			q.AddURL(e.Text)
		}
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
                if strings.Contains(e.Text, "/articles/") || strings.Contains(e.Text, "/news/") {
                        q.AddURL(e.Text)
                }
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("error:", err, r.Request.URL, r.StatusCode)
		if r.StatusCode == 429 {
			q.AddURL(r.Request.URL.String())
		}
	})

	c.OnHTML(`html`, func(e *colly.HTMLElement) {

		// check if news page
		//if !strings.Contains(e.Request.Ctx.Get("url"), "/articles/") || !strings.Contains(e.Request.Ctx.Get("url"), "/news/") {
		//	return
		//}

		// check in the databse if exists
		var pageExists models.Page
		if !cfg.DryMode {
			if !cfg.DB.Where("link = ?", e.Request.Ctx.Get("url")).First(&pageExists).RecordNotFound() {
				fmt.Printf("skipping url=%s as already exists\n", e.Request.Ctx.Get("url"))
				return
			}
		}

		page := &models.Page{}
		page.Link = e.Request.Ctx.Get("url")
		page.Source = "barrons.com"
		page.Class = "news"

		var categories []string
		e.ForEach(`ul[itemtype="http://schema.org/BreadcrumbList"] a`, func(_ int, el *colly.HTMLElement) {
			if el.Text != "" {
				categories = append(categories, strings.TrimSpace(el.Text))
			}
		})
		page.Categories = strings.Join(categories, ",")
		page.Title = strings.TrimSpace(e.ChildText(`h1[itemprop="headline"]`))
		page.Authors = strings.TrimSpace(e.ChildText(`div.author span.name`))

		publishedAtStr := e.ChildText("time.timestamp")
		publishedAtStr = strings.Replace(publishedAtStr, "Updated ", "", -1)
		publishedAtStr = strings.Replace(publishedAtStr, "Original ", "", -1)
		publishedAtStr = strings.Replace(publishedAtStr, " ET", "", -1)
		publishedAtParts := strings.Split(publishedAtStr, "/")
		var publishedAt string
		if len(publishedAtParts) > 1 {
			publishedAt = strings.TrimSpace(publishedAtParts[0])
			if cfg.IsDebug {
				fmt.Println("publishedAt:", publishedAt)
			}
			// convert date to time
			publishedAtTime, err := dateparse.ParseAny(publishedAt)
			if err != nil {
				log.Fatal(err)
			}
			page.PublishedAt = publishedAtTime
		} else {
			page.PublishedAt = time.Now()
		}

		// articletext.
		// page.Content = strings.TrimSpace(e.ChildText(`div[itemprop="articleBody"]`))
		content, err := articletext.GetArticleTextFromHtmlNode(e.Node)
		if err != nil {
			log.Fatal(err)
		}

		page.Content = content

		// e.ForEach(`script[type="application/ld+json"]`, func(_ int, el *colly.HTMLElement) {
		// })

		// page.PageProperties = append(page.PageProperties, models.PageProperty{Name: "InteriorColor", Value: val})
		//if cfg.IsDebug {
			pp.Println(page)
		//}

		if page.Link == "" && page.Content == "" && page.PublishedAt.String() == "" && page.Authors == "" {
			return
		}

		if !cfg.DryMode {
			if err := cfg.DB.Create(&page).Error; err != nil {
				log.Fatalf("create page (%v) failure, got err %v", page, err)
				return
			}
		}

	})

	c.OnResponse(func(r *colly.Response) {
		if cfg.IsDebug {
			fmt.Println("OnResponse from", r.Ctx.Get("url"))
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		//if cfg.IsDebug {
		fmt.Println("Visiting", r.URL.String())
		//}
		r.Ctx.Put("url", r.URL.String())
	})

	if cfg.IsSitemapIndex {
		log.Infoln("extractSitemapIndex...")
		for _, i := range cfg.URLs {
			sitemaps, err := sitemap.ExtractSitemapIndex(i)
			if err != nil {
				log.Fatal("ExtractSitemapIndex:", err)
				return err
			}
			// shall we shuffle ?
			// utils.Shuffle(sitemaps)
			for _, s := range sitemaps {
				log.Infoln("processing ", s)
				if strings.HasSuffix(s, ".gz") {
					log.Infoln("extract sitemap gz compressed...")
					locs, err := sitemap.ExtractSitemapGZ(s)
					if err != nil {
						log.Fatal("ExtractSitemapGZ: ", err, "sitemap: ", s)
						return err
					}
					// utils.Shuffle(locs)
					for _, loc := range locs {
						if strings.Contains(loc, "/articles/") {
							q.AddURL(loc)
						}
					}
				} else {
					locs, err := sitemap.ExtractSitemap(s)
					if err != nil {
						log.Fatal("ExtractSitemap", err)
						return err
					}
					// utils.Shuffle(locs)
					for _, loc := range locs {
						if strings.Contains(loc, "/articles/") {
							q.AddURL(loc)
						}
					}
				}
			}
		}
	} else {
		for _, u := range cfg.URLs {
			q.AddURL(u)
		}
	}

	// Consume URLs
	q.Run(c)

	return nil
}
