package crawler

import (
	"fmt"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/corpix/uarand"
	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"

	"github.com/lucmichalski/finance-dataset/pkg/articletext"
	"github.com/lucmichalski/finance-dataset/pkg/colly"
	"github.com/lucmichalski/finance-dataset/pkg/colly/proxy"
	"github.com/lucmichalski/finance-dataset/pkg/colly/queue"
	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/models"
	"github.com/lucmichalski/finance-dataset/pkg/sitemap"
	// "github.com/lucmichalski/finance-dataset/pkg/utils"
)

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		// colly.AllowURLRevisit(),
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
	)

	// Rotate two socks5 proxies
	rp, err := proxy.RoundRobinProxySwitcher("http://localhost:1080")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	// create a request queue with 1 consumer thread until we solve the multi-threadin of the darknet model
	q, _ := queue.New(
		cfg.ConsumerThreads,
		&queue.InMemoryQueueStorage{
			MaxSize: cfg.QueueMaxSize,
		},
	)

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//sitemap/loc", func(e *colly.XMLElement) {
		q.AddURL(e.Text)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		q.AddURL(e.Text)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("error:", err, r.Request.URL, r.StatusCode)
		fmt.Println("Body: ", string(r.Body))
		//q.AddURL(r.Request.URL.String())
	})

	c.OnHTML(`html`, func(e *colly.HTMLElement) {

		// check if news page
		if !strings.Contains(e.Request.Ctx.Get("url"), "/news/") {
			return
		}

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
		page.Source = "scmp.com"
		page.Class = "news"

		// categories
		var categories []string
		e.ForEach(`a.tag__link`, func(_ int, el *colly.HTMLElement) {
			if el.Text != "" {
				categories = append(categories, strings.TrimSpace(el.Text))
			}
		})
		page.Categories = strings.Join(categories, ",")

		// title
		page.Title = strings.TrimSpace(e.ChildText("h1.info__headline headline"))

		// author
		page.Authors = e.ChildText("span.main-info__name")

		// date
		e.ForEach(`time`, func(_ int, el *colly.HTMLElement) {
			if el.Attr("datetime") != "" {
				publishedAtTime, err := dateparse.ParseAny(el.Attr("datetime"))
				if err != nil {
					log.Fatal(err)
				}
				page.PublishedAt = publishedAtTime
			}
		})

		var tags []string
		e.ForEach(`div.topic__title span`, func(_ int, el *colly.HTMLElement) {
			if el.Text != "" {
				tags = append(tags, el.Text)
			}
		})
		page.Tags = strings.Join(tags, ",")

		// article content
		content, err := articletext.GetArticleTextFromHtmlNode(e.Node)
		if err != nil {
			log.Fatal(err)
		}
		page.Content = content

		if cfg.IsDebug {
			pp.Println("page:", page)
		}

		// page.PageProperties = append(page.PageProperties, models.PageProperty{Name: "InteriorColor", Value: val})

		if page.Title == "" && page.Link == "" && page.Content == "" && page.PublishedAt.String() == "" {
			return
		}

		if cfg.IsDebug {
			pp.Println("page:", page)
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
						if strings.Contains(loc, "/news/") {
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
						if strings.Contains(loc, "/news/") {
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
