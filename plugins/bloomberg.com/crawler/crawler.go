package crawler

import (
	"fmt"
	"strings"
	"time"

	captcha "github.com/gocolly/twocaptcha"
	"github.com/k0kubun/pp"
	"github.com/nozzle/throttler"
	log "github.com/sirupsen/logrus"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/models"
	"github.com/lucmichalski/finance-dataset/pkg/selenium"
	"github.com/lucmichalski/finance-dataset/pkg/selenium/chrome"
	slog "github.com/lucmichalski/finance-dataset/pkg/selenium/log"
	"github.com/lucmichalski/finance-dataset/pkg/sitemap"
)

var (
	successMsg     = "div[class='recaptcha-success']"
	apiKey2captcha = "4c4cb693aef7c0dbd7af6622e78ee5eb"         // Your 2captcha.com API key
	recaptchaV2Key = "6Lcj-R8TAAAAABs3FrRPuQhLMbp5QrHsHufzLf7b" // v2 Site Key (data-sitekey) inspected from target website
)

func Extract(cfg *config.Config) error {

	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--headless",
			"--no-sandbox",
			"--start-maximized",
			"--window-size=1024,768",
			"--disable-crash-reporter",
			"--hide-scrollbars",
			"--disable-gpu",
			"--disable-setuid-sandbox",
			"--disable-infobars",
			"--window-position=0,0",
			"--ignore-certifcate-errors",
			"--ignore-certifcate-errors-spki-list",
			"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/604.4.7 (KHTML, like Gecko) Version/11.0.2 Safari/604.4.7",
			// "--proxy-server=http://tor-haproxy:8119",
			// "--host-resolver-rules=\"MAP * 0.0.0.0 , EXCLUDE localhost\"",
		},
	}
	caps.AddChrome(chromeCaps)

	caps.SetLogLevel(slog.Server, slog.Off)
	caps.SetLogLevel(slog.Browser, slog.Off)
	caps.SetLogLevel(slog.Client, slog.Off)
	caps.SetLogLevel(slog.Driver, slog.Off)
	caps.SetLogLevel(slog.Performance, slog.Off)
	caps.SetLogLevel(slog.Profiler, slog.Off)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 4444))
	if err != nil {
		return err
	}
	defer wd.Quit()

	var links []string
	if cfg.IsSitemapIndex {
		for _, i := range cfg.URLs {
			log.Infoln("extractSitemapIndex...", i)
			sitemaps, err := sitemap.ExtractSitemapIndex(i)
			if err != nil {
				log.Fatal("ExtractSitemapIndex:", err)
				return err
			}
			for _, s := range sitemaps {
				log.Infoln("processing ", s)
				if strings.HasSuffix(s, ".gz") {
					log.Infoln("extract sitemap gz compressed...")
					locs, err := sitemap.ExtractSitemapGZ(s)
					if err != nil {
						log.Fatal("ExtractSitemapGZ: ", err, "sitemap: ", s)
						return err
					}
					for _, loc := range locs {
						if strings.HasPrefix(loc, "news/articles") {
							links = append(links, loc)
						}
					}
				} else {
					locs, err := sitemap.ExtractSitemap(s)
					if err != nil {
						log.Warn("ExtractSitemap", err)
						// return err
						continue
					}
					for _, loc := range locs {
						if strings.HasPrefix(loc, "news/articles") {
							links = append(links, loc)
						}
					}
				}
			}
		}
	} else {
		links = append(links, cfg.URLs...)
	}

	pp.Println("found:", len(links))

	t := throttler.New(1, len(links))

	for _, link := range links {
		log.Println("processing link:", link)
		go func(link string) error {
			defer t.Done(nil)
			err := scrapeSelenium(link, cfg, wd)
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

// 6Lcj-R8TAAAAABs3FrRPuQhLMbp5QrHsHufzLf7b
func scrapeSelenium(url string, cfg *config.Config, wd selenium.WebDriver) error {

	// check in the databse if exists
	var pageExists models.Page
	if !cfg.DB.Where("link = ?", url).First(&pageExists).RecordNotFound() {
		fmt.Printf("skipping link=%s as already exists\n", url)
		return nil
	}

	err := wd.Get(url)
	if err != nil {
		return err
	}

	src, err := wd.PageSource()
	if err != nil {
		return err
	}
	// fmt.Println("source", src)

	if strings.Contains(src, recaptchaV2Key) {
		log.Warnln("does contain captacha challenge")
		wd = v2Solver(url, wd)
	}

	// create vehicle
	page := &models.Page{}
	page.Link = url
	page.Class = "news"
	page.Source = "bloomberg.com"

	// write email
	titleCnt, err := wd.FindElement(selenium.ByCSSSelector, "h1[class=\"lede-text-v2__hed\"]")
	if err != nil {
		return err
	}

	title, err := titleCnt.Text()
	if err != nil {
		return err
	}
	pp.Println("title:", title)

	authorsCnt, err := wd.FindElements(selenium.ByCSSSelector, "div.author-v2 a")
	if err != nil {
		return err
	}

	for _, authorCnt := range authorsCnt {
		author, err := authorCnt.Text()
		if err != nil {
			return err
		}
		pp.Println("author:", author)
	}

	timeCnt, err := wd.FindElement(selenium.ByCSSSelector, "time[itemprop=\"datePublished\"]")
	if err != nil {
		return err
	}

	publishedAt, err := timeCnt.Text()
	if err != nil {
		return err
	}
	pp.Println("publishedAt:", publishedAt)

	bodyCnt, err := wd.FindElement(selenium.ByCSSSelector, "div.body-copy-v2.fence-body")
	if err != nil {
		return err
	}

	body, err := bodyCnt.Text()
	if err != nil {
		return err
	}
	pp.Println("body:", body)

	pp.Println(page)
	// save page
	if !cfg.DryMode {
		if err := cfg.DB.Create(&page).Error; err != nil {
			log.Fatalf("create page (%v) failure, got err %v", page, err)
			return err
		}
	}

	return nil
}

func v2Solver(recaptchaURL string, wd selenium.WebDriver) selenium.WebDriver {
	c := captcha.New(apiKey2captcha)
	solved, err := c.SolveRecaptchaV2(recaptchaURL, recaptchaV2Key)
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Println("[✓](v2) Solved via 2captcha.com") // String

		// Show hidden Textarea
		_, err = wd.ExecuteScript(fmt.Sprintf("document.getElementById('g-recaptcha-response').style='"+"width: 250px; height: 40px; border: 1px solid rgb(193, 193, 193); margin: 10px 25px; padding: 0px; resize: none;"+"';"), nil)
		if err != nil {
			panic(fmt.Sprintf("[✕](v2) Textarea style not changed: %s", err)) // ReCaptcha Key wasn't submitted.
		} else {
			textArea, err := wd.FindElement(selenium.ByID, "g-recaptcha-response")
			if err != nil {
				panic(err)
			}
			if err := textArea.Clear(); err != nil {
				log.Println("\n\tTextarea not cleared.\n")

				panic(err)
			} else {
				// Send Solved Key
				_, err = wd.ExecuteScript(fmt.Sprintf("document.getElementById('g-recaptcha-response').innerHTML='"+solved+"';"), nil)
				if err != nil {
					panic(fmt.Sprintf("[✕](v2) Reponse Key Submission Error: %s", err)) // ReCaptcha Key wasn't submitted back to website.
				} else {
					log.Println("[✓](v2) ReCaptcha Response Key submitted back to site's captcha")
				}

				time.Sleep(3 * time.Second) // Wait

				// Submit form
				_, err = wd.ExecuteScript(fmt.Sprintf("document.getElementById('recaptcha-demo-form').submit();"), nil)
				if err != nil {
					log.Println(fmt.Sprintf("[✕](v2) Submit button not clicked: %s", err)) // ReCaptcha Form wasn't submitted.

					time.Sleep(3 * time.Minute) // Wait
				} else {
					log.Println("[✓](v2) Submit button clicked.")

					time.Sleep(3 * time.Second) // Wait

					_, err := wd.FindElement(selenium.ByCSSSelector, successMsg)
					if err != nil {
						log.Println(fmt.Sprintf("[✕](v2) Success message not dislayed: %s", err))
					} else {
						log.Println("[✓](v2) ReCaptcha successfully solved!")
					}

					time.Sleep(2 * time.Minute) // Wait

					// End of script
				}
			}
		}
	}
	return wd
}
