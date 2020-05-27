package crawler

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	"net/http"
	"crypto/tls"

	"github.com/araddon/dateparse"
	"github.com/astaxie/flatmap"
	"github.com/corpix/uarand"
	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"

	"github.com/lucmichalski/finance-dataset/pkg/articletext"
	"github.com/lucmichalski/finance-dataset/pkg/colly"
	//"github.com/lucmichalski/finance-dataset/pkg/colly/proxy"
	"github.com/lucmichalski/finance-dataset/pkg/colly/queue"
	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/models"
	"github.com/lucmichalski/finance-dataset/pkg/sitemap"
)

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		// colly.AllowURLRevisit(),
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
	)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c.WithTransport(tr)

	// Rotate two socks5 proxies
	/*
	rp, err := proxy.RoundRobinProxySwitcher("http://localhost:8119")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)
	*/
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
		page.Source = "anticor.org"
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
		page.Title = e.ChildText("h1.entry-title")

		// author
		page.Authors = e.ChildText("a[rel=author]")

		// date
		e.ForEach(`time.updated`, func(_ int, el *colly.HTMLElement) {
			if el.Attr("datetime") != "" {
				publishedAtTime, err := dateparse.ParseAny(el.Attr("datetime"))
				if err != nil {
					log.Fatal(err)
				}
				page.PublishedAt = publishedAtTime
			} else {
				page.PublishedAt = time.Now()
			}
		})

		var pageInfo map[string]interface{}
		e.ForEach(`script[type="application/ld+json"].yoast-schema-graph`, func(_ int, el *colly.HTMLElement) {
			jsonLdStr := strings.TrimSpace(el.Text)
			if cfg.IsDebug {
				fmt.Println("jsonLdStr:", jsonLdStr)
			}
			jsonLdStr = "{\"jsonld\":" + jsonLdStr + "}"
			if err := json.Unmarshal([]byte(jsonLdStr), &pageInfo); err != nil {
				log.Warnln("unmarshal error, ", err)
				return
			}

			fm, err := flatmap.Flatten(pageInfo)
			if err != nil {
				log.Fatal(err)
			}
			var ks []string
			for k := range fm {
				ks = append(ks, k)
			}
			sort.Strings(ks)

			if cfg.IsDebug {
				for _, k := range ks {
					fmt.Println(k, ":", fm[k])
				}
			}

			if val, ok := fm["jsonld.@graph.3.@type"]; ok {
				page.Class = val
			} else {
				page.Class = "post"
			}

			if val, ok := fm["jsonld.@graph.3.articleSection"]; ok {
				page.Categories = val
			}

		})

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
