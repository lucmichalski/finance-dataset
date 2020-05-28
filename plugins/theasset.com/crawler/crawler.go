package crawler

import (
	"fmt"
	"strings"
	"time"

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
		//colly.AllowURLRevisit(),
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
		colly.AllowedDomains(cfg.AllowedDomains...),
	)

	// Rotate two socks5 proxies
	rp, err := proxy.RoundRobinProxySwitcher("http://localhost:8119")
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
		if r.StatusCode == 429 {
			q.AddURL(r.Request.URL.String())
		}
	})

	c.OnHTML(`a[href]`, func(e *colly.HTMLElement) {
		fmt.Println("Found link: ", e.Attr("href"))
		// check if article page
		if strings.Contains(e.Attr("href"), "/article/") {
			q.AddURL(e.Attr("href"))
		}
	})

	c.OnHTML(`html`, func(e *colly.HTMLElement) {

		// check if news page
		if !strings.Contains(e.Request.Ctx.Get("url"), "/article/") {
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
		page.Source = "theasset.com"
		page.Class = "article"

		// categories
		var categories []string
		e.ForEach(`div.single-post__header a`, func(_ int, el *colly.HTMLElement) {
			if el.Text != "" {
				categories = append(categories, el.Text)
			}
		})
		page.Categories = strings.Join(categories, ",")

		// title
		page.Title = e.ChildText("div.single-post__title")

		// author
		page.Authors = "The Asset"

		// date
		publishedAtStr := e.ChildText("div.single-post__d")
		publishedAtParts := strings.Split(publishedAtStr, " | ")
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

		if page.Link == "" && page.Content == "" && page.PublishedAt.String() == "" {
			return
		}

		if cfg.IsDebug {
			pp.Println("page:", page)
		}

		if !cfg.DryMode {
			if err := cfg.DB.Create(&page).Error; err != nil {
				log.Warnf("create page (%v) failure, got err %v", page, err)
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
						if strings.Contains(loc, "/article/") {
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
						if strings.Contains(loc, "/article/") {
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
		for i := 0; i < 100; i++ {
			u := fmt.Sprintf("https://theasset.com/mwapi/article/loop/wealth-management/%d", i)
			q.AddURL(u)
		}

	}

	// Consume URLs
	q.Run(c)

	return nil
}
