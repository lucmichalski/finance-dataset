// This example utilizes github.com/aerokube/selenoid (Selenium in Docker Containers) to run at a larger scale + VNC monitoring

package main

import (
	"fmt"
	"log"
	"time"

	captcha "github.com/gocolly/twocaptcha"

	"github.com/fatih/color"
	"github.com/tebeka/selenium"
)

var (
	red    = color.New(color.FgRed, color.Bold)
	yellow = color.New(color.FgYellow, color.Bold)

	successMsg = "div[class='recaptcha-success']"

	apiKey2captcha = "4c4cb693aef7c0dbd7af6622e78ee5eb"     // Your 2captcha.com API key
	recaptchaV2Key = "e700c1f0-9d81-11ea-9b7d-b51e416fd3db" // v2 Site Key (data-sitekey) inspected from target website
)

// <div style="width: 302px; height: 462px;"><iframe src="/recaptcha/api/fallback?k=6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-" frameborder="0" scrolling="no"></iframe><div><textarea id="g-recaptcha-response" name="g-recaptcha-response" class="g-recaptcha-response"></textarea></div></div><br>

const (
	host = "http://127.0.0.1" // Selenoid Host IP
	port = 4444               // 4444 (Selenoid Port)

	recaptchaURL = "https://www.similarweb.com/website/carvana.com" // Target website URL
)

func v2Solver(wd selenium.WebDriver) {
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
				_, _ = red.Println("\n\tTextarea not cleared.\n")

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
					_, _ = red.Println(fmt.Sprintf("[✕](v2) Submit button not clicked: %s", err)) // ReCaptcha Form wasn't submitted.

					time.Sleep(3 * time.Minute) // Wait
				} else {
					log.Println("[✓](v2) Submit button clicked.")

					time.Sleep(3 * time.Second) // Wait

					_, err := wd.FindElement(selenium.ByCSSSelector, successMsg)
					if err != nil {
						_, _ = red.Println(fmt.Sprintf("[✕](v2) Success message not dislayed: %s", err))
					} else {
						log.Println("[✓](v2) ReCaptcha successfully solved!")
					}

					time.Sleep(2 * time.Minute) // Wait

					// End of script
				}
			}
		}
	}
}

func main() {
	// Connect to the WebDriver instance running locally. (Selenoid)
	caps := selenium.Capabilities{"browserName": "chrome", "enableVNC": true}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("%s:%d/wd/hub", host, port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic occured: ", r)

			wd.Quit()

			main()
		}
	}()

	// Navigate to page containing recaptcha V2
	if err := wd.Get(recaptchaURL); err != nil {
		panic(err)
	} else {
		_, _ = yellow.Println("\tPage reached.")

		if title, err := wd.Title(); err == nil {
			fmt.Printf("\nPage Title: \t%s\n\n", title)
		} else {
			_, _ = red.Printf("\n\tFailed to get page title: %s\n", err)
			return
		}
	}

	time.Sleep(3 * time.Second) // wait

	v2Solver(wd)
}
