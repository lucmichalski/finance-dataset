package crawler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/k0kubun/pp"
	"github.com/nozzle/throttler"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	ccsv "github.com/lucmichalski/finance-dataset/pkg/csv"
	"github.com/lucmichalski/finance-dataset/pkg/models"
	"github.com/lucmichalski/finance-dataset/pkg/sitemap"
	"github.com/lucmichalski/finance-dataset/pkg/utils"

	pmodels "github.com/lucmichalski/finance-contrib/lesechos.fr/models"
)

// change it asap, pass it through the config
const (
	torProxyAddress   = "socks5://51.210.37.251:5566"
	torPrivoxyAddress = "socks5://51.210.37.251:8119"
)

func Extract(cfg *config.Config) error {

	rand.Seed(time.Now().UnixNano())

	var links []string
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
			var pageExists models.Page
			if !cfg.DryMode {
				if !cfg.DB.Where("link = ?", loc[0]).First(&pageExists).RecordNotFound() {
					fmt.Printf("skipping url=%s as already exists\n", loc[0])
					continue
				}
			}
			links = append(links, loc[0])
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
							links = append(links, loc)
							csvSitemap.Write([]string{loc, s})
							csvSitemap.Flush()
						}
					} else {
						locs, err := sitemap.ExtractSitemap(s)
						if err != nil {
							log.Fatal("ExtractSitemap", err)
							return err
						}
						utils.Shuffle(locs)
						for _, loc := range locs {
							links = append(links, loc)
							csvSitemap.Write([]string{loc, s})
							csvSitemap.Flush()
						}
					}
				}
			}
		} else {
			for _, u := range cfg.URLs {
				links = append(links, u)
				csvSitemap.Write([]string{u, ""})
				csvSitemap.Flush()
			}
		}
	}

	pp.Println("found:", len(links))

	t := throttler.New(cfg.ConsumerThreads, len(links))

	for _, link := range links {
		log.Println("processing link:", link)
		go func(link string) error {
			defer t.Done(nil)
			// https://www.lesechos.fr/industrie-services/automobile/lautomobile-tricolore-attend-febrilement-son-plan-de-relance-1205614
			err := getArticle(link, cfg)
			if err != nil {
				log.Warnln(err)
			}
			return err
		}(link)
		t.Throttle()
	}

	// throttler errors iteration
	if t.Err() != nil {
		// Loop through the errors to see the details
		for i, err := range t.Errs() {
			log.Printf("error #%d: %s", i, err)
		}
		log.Fatal(t.Err())
	}

	return nil
}

func getArticle(link string, cfg *config.Config) error {

	// check in the databse if exists
	var pageExists models.Page
	if !cfg.DryMode {
		if !cfg.DB.Where("link = ?", link).First(&pageExists).RecordNotFound() {
			fmt.Printf("skipping url=%s as already exists\n", link)
			return nil
		}
	}

	linkParts := strings.Split(link, "-")
	linkID := linkParts[len(linkParts)-1]
	rawUrl := fmt.Sprintf("https://api.lesechos.fr/api/v1/articles/%s", linkID)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	tbProxyURL, err := url.Parse(torProxyAddress)
	if err != nil {
		return err
	}

	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	if err != nil {
		return err
	}
	tbTransport := &http.Transport{
		Dial: tbDialer.Dial,
	}
	client.Transport = tbTransport

	request, err := http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer response.Body.Close()

	// unmarshall response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var result pmodels.Article
	json.NewDecoder(strings.NewReader(string(body))).Decode(&result)
	// pp.Println(result)

	page := &models.Page{}
	page.Link = link
	page.Source = "lesechos.fr"
	page.Class = "article"

	var authors []string
	if len(result.Stripes) == 0 {
		return errors.New("No stripes")
	}

	if len(result.Stripes[0].MainContent) == 0 {
		return errors.New("No MainContent")
	}

	for _, author := range result.Stripes[0].MainContent[0].Data.Authors {
		authors = append(authors, author.Signature)
	}

	page.Authors = strings.Join(authors, ",")

	page.Title = result.Stripes[0].MainContent[0].Data.Title
	page.Content = result.Stripes[0].MainContent[0].Data.Description
	publishedAtTime, err := dateparse.ParseAny(result.Stripes[0].MainContent[0].Data.PublicationDate)
	if err != nil {
		log.Fatal(err)
	}
	page.PublishedAt = publishedAtTime

	var categories []string
	categories = append(categories, result.Stripes[0].MainContent[0].Data.Section.Label)
	categories = append(categories, result.Stripes[0].MainContent[0].Data.Subsection.Label)
	page.Categories = strings.Join(categories, ",")

	var tags []string
	for _, cat := range result.Stripes[0].MainContent[0].Data.Tags.Categorization {
		tags = append(tags, cat.Names)
	}
	//for _, geo := range result.Stripes[0].MainContent[0].Data.Tags.Geography {
	//	tags = append(tags, geo)
	//}
	for _, org := range result.Stripes[0].MainContent[0].Data.Tags.Organizations {
		tags = append(tags, org)
	}
	for _, people := range result.Stripes[0].MainContent[0].Data.Tags.People {
		tags = append(tags, people)
	}

	page.Tags = strings.Join(tags, ",")

	if page.Title == "" {
		return errors.New("Article title is missing, skipping this entry...")
	}

	if page.Link == "" && page.Content == "" && page.PublishedAt.String() == "" {
		return errors.New("No enough attributes to be registered")
	}

	if cfg.IsDebug {
		pp.Println("page:", page)
	}

	if !cfg.DryMode {
		if err := cfg.DB.Create(&page).Error; err != nil {
			log.Warnf("create page (%v) failure, got err %v", page, err)
			return err
		}
	}

	return nil
}
