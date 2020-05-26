package crawler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	// "github.com/araddon/dateparse"
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

	// publishedAtTime, err := dateparse.ParseAny(publishedAtStr)

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
			linkParts := strings.Split(link, "-")
			linkID := linkParts[len(linkParts)-1]
			err := getArticle(linkID, cfg)
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

func getArticle(id string, cfg *config.Config) error {
	rawUrl := fmt.Sprintf("https://api.lesechos.fr/api/v1/articles/%s", id)

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

	// log.Println(string(body))
	fmt.Println("Body:", string(body))
	var result pmodels.Article
	json.NewDecoder(strings.NewReader(string(body))).Decode(&result)
	pp.Println(result)

	page := &models.Page{}
	pp.Println(page)
	return nil
}
