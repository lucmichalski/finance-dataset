package crawler

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/araddon/dateparse"
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
		// colly.AllowURLRevisit(),
		// colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
	)

	// Rotate two socks5 proxies
	rp, err := proxy.RoundRobinProxySwitcher("socks5://localhost:1080")
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
		// fmt.Println("Body", string(r.Body))
		// os.Exit(1)
		// q.AddURL(r.Request.URL.String())
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
		page.Source = "reuters.com"
		page.Class = "article"

		// categories
		var categories []string
		e.ForEach(`div.ArticleHeader_channel a`, func(_ int, el *colly.HTMLElement) {
			if el.Text != "" {
				categories = append(categories, el.Text)
			}
		})
		page.Categories = strings.Join(categories, ",")

		// title
		page.Title = strings.TrimSpace(e.ChildText("h1.ArticleHeader_headline"))

		// authors
		var authors []string
		e.ForEach(`div.BylineBar_byline a`, func(_ int, el *colly.HTMLElement) {
			if el.Text != "" {
				authors = append(authors, el.Text)
			}
		})
		page.Authors = strings.Join(authors, ",")

		// date
		publishedAtStr := e.ChildText("div.ArticleHeader_date")
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

		// article content
		content, err := articletext.GetArticleTextFromHtmlNode(e.Node)
		if err != nil {
			log.Fatal(err)
		}
		page.Content = content

		if cfg.IsDebug {
			pp.Println("page:", page)
		}

		if page.Title == "" && page.Link == "" && page.Content == "" && page.PublishedAt.String() == "" && len(authors) == 0 {
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
		fmt.Println("Visiting", r.URL.String())
		var pageExists models.Page
		if !cfg.DryMode {
			if !cfg.DB.Where("link = ?", r.URL.String()).First(&pageExists).RecordNotFound() {
				fmt.Printf("skipping url=%s as already exists\n", r.URL.String())
				return
			}
		}
		r.Ctx.Put("url", r.URL.String())
	})

	utils.EnsureDir("./shared/queue/")
	if _, err := os.Stat("shared/queue/reuters.com_sitemap.csv"); !os.IsNotExist(err) {
		file, err := os.Open("shared/queue/reuters.com_sitemap.csv")
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
		csvSitemap, err := ccsv.NewCsvWriter("shared/queue/reuters.com_sitemap.csv", ',')
		if err != nil {
			panic("Could not open `reuters.com_sitemap.csv` for writing")
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

			start := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
			end := time.Now()
			for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
				startDate := d.AddDate(0, 0, -1)
				u := fmt.Sprintf("https://www.reuters.com/sitemap_%s-%s.xml", startDate.Format("20060102"), d.Format("20060102"))
				fmt.Println("url:", u)
				q.AddURL(u)
			}
		}
	}

	// Consume URLs
	q.Run(c)

	return nil
}
