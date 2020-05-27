package crawler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/araddon/dateparse"
	// "github.com/corpix/uarand"
	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/models"
	"github.com/lucmichalski/finance-dataset/pkg/sitemap"
	// "github.com/lucmichalski/finance-dataset/pkg/utils"
)

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		// colly.AllowURLRevisit(),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36"),
		colly.CacheDir(cfg.CacheDir),
	)

	// Rotate two socks5 proxies
	// rp, err := proxy.RoundRobinProxySwitcher("http://51.210.37.251:5566")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// c.SetProxyFunc(rp)

	// create a request queue with 1 consumer thread until we solve the multi-threadin of the darknet model
	q, _ := queue.New(
		cfg.ConsumerThreads,
		&queue.InMemoryQueueStorage{
			MaxSize: cfg.QueueMaxSize,
		},
	)

	c.DisableCookies()

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
		q.AddURL(r.Request.URL.String())
	})

	c.OnHTML(`html`, func(e *colly.HTMLElement) {

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
		page.Source = "fcpablog.com"
		page.Class = "post"

		// categories
		var categories []string
		e.ForEach(`li.category a`, func(_ int, el *colly.HTMLElement) {
			if el.Text != "" {
				categories = append(categories, el.Text)
			}
		})
		page.Categories = strings.Join(categories, ",")

		// title
		page.Title = e.ChildText("h1.elementor-heading-title")

		// author
		page.Authors = e.ChildText("span.elementor-post-info__item--type-author")

		// date
		publishedAtStr := e.ChildText("span.elementor-post-info__item--type-date")
		publishedAtTime, err := dateparse.ParseAny(publishedAtStr)
		if err != nil {
			log.Fatal(err)
		}
		page.PublishedAt = publishedAtTime

		var tags []string
		e.ForEach(`a.elementor-post-info__terms-list-item`, func(_ int, el *colly.HTMLElement) {
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

		// var carDataImage []string
		// e.ForEach(`div.gallery-controls__thumbnail-image`, func(_ int, el *colly.HTMLElement) {
		// 	carImage := el.Attr("data-image")
		// 	if cfg.IsDebug {
		// 		fmt.Println("carImage:", carImage)
		// 	}
		// 	carDataImage = append(carDataImage, carImage)
		// })

		if page.Link == "" && page.Content == "" && page.PublishedAt.String() == "" {
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
