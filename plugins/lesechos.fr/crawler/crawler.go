package crawler

import (
	"encoding/csv"
	"fmt"
	"os"
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
	ccsv "github.com/lucmichalski/finance-dataset/pkg/csv"
	"github.com/lucmichalski/finance-dataset/pkg/models"
	"github.com/lucmichalski/finance-dataset/pkg/sitemap"
	"github.com/lucmichalski/finance-dataset/pkg/utils"
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

		// check if news page
		//if !strings.Contains(e.Request.Ctx.Get("url"), "/news/") || !strings.Contains(e.Request.Ctx.Get("url"), "/news/") {
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
		page.Source = "lesechos.fr"
		page.Class = "news"

		// categories
		var categories []string
		e.ForEach(`ul.breadcrumb-list span`, func(_ int, el *colly.HTMLElement) {
			if el.Text != "" {
				categories = append(categories, el.Text)
			}
		})
		page.Categories = strings.Join(categories, ",")

		// title
		page.Title = e.ChildText("h1[itemprop=\"headline\"]")

		// author
		var authors []string
		e.ForEach(`a.author`, func(_ int, el *colly.HTMLElement) {
			if el.Text != "" {
				authors = append(authors, strings.TrimSpace(el.Text))
			}
		})
		page.Authors = strings.Join(authors, ",")

		// date
		e.ForEach(`time[itemprop="datePublished"]`, func(_ int, el *colly.HTMLElement) {
			publishedAtStr := el.Attr("datetime")
			publishedAtStr = strings.TrimSpace(publishedAtStr)
			if publishedAtStr != "" {
				publishedAtTime, err := dateparse.ParseAny(publishedAtStr)
				if err != nil {
					log.Fatal(err)
				}
				page.PublishedAt = publishedAtTime
			} else {
				page.PublishedAt = time.Now()
			}
		})

		// article content
		content, err := articletext.GetArticleTextFromHtmlNode(e.Node)
		if err != nil {
			log.Fatal(err)
		}
		page.Content = content

		// page.Content = e.ChildText("div[itemprop=\"articleBody\"]")

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

	utils.EnsureDir("./shared/queue/")
	if _, err := os.Stat("shared/queue/lesechos.fr_sitemap.csv"); !os.IsNotExist(err) {
		file, err := os.Open("shared/queue/lesechos.fr_sitemap.csv")
		if err != nil {
			return err
		}

		reader := csv.NewReader(file)
		reader.Comma = ','
		reader.LazyQuotes = true
		data, err := reader.ReadAll()
		if err != nil {
			return err
		}

		utils.Shuffle(data)
		for _, loc := range data {
			q.AddURL(loc[0])
		}
	} else {

		// save discovered links
		csvSitemap, err := ccsv.NewCsvWriter("shared/queue/lesechos.fr_sitemap.csv", ',')
		if err != nil {
			panic("Could not open `lesechos.fr_sitemap.csv` for writing")
		}

		// Flush pending writes and close file upon exit of Sitemap()
		defer csvSitemap.Close()

		if cfg.IsSitemapIndex {
			log.Infoln("extractSitemapIndex...")
			for _, i := range cfg.URLs {
				sitemaps, err := sitemap.ExtractSitemapIndex(i)
				if err != nil {
					log.Fatal("ExtractSitemapIndex:", err)
					return err
				}
				// shall we shuffle ?
				utils.Shuffle(sitemaps)
				for _, s := range sitemaps {
					log.Infoln("processing ", s)
					if strings.HasSuffix(s, ".gz") {
						log.Infoln("extract sitemap gz compressed...")
						locs, err := sitemap.ExtractSitemapGZ(s)
						if err != nil {
							log.Fatal("ExtractSitemapGZ: ", err, "sitemap: ", s)
							return err
						}
						utils.Shuffle(locs)
						for _, loc := range locs {
							if strings.Contains(loc, "/news/") || strings.Contains(loc, "/pmn/") {
								q.AddURL(loc)
								csvSitemap.Write([]string{loc, s})
								csvSitemap.Flush()
							}
						}
					} else {
						locs, err := sitemap.ExtractSitemap(s)
						if err != nil {
							log.Fatal("ExtractSitemap", err)
							return err
						}
						utils.Shuffle(locs)
						for _, loc := range locs {
							if strings.Contains(loc, "/news/") || strings.Contains(loc, "/pmn/") {
								q.AddURL(loc)
								csvSitemap.Write([]string{loc, s})
								csvSitemap.Flush()
							}
						}
					}
				}
			}
		} else {
			for _, u := range cfg.URLs {
				q.AddURL(u)
				csvSitemap.Write([]string{u, ""})
				csvSitemap.Flush()
			}
		}
	}

	// Consume URLs
	q.Run(c)

	return nil
}
