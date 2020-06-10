package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nozzle/throttler"
)

func main() {
	t := throttler.New(50, 100000000)

	start := time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Now() // start.AddDate(, 0, 0)
	for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
		startDate := d.AddDate(0, 0, -1)
		u := fmt.Sprintf("https://www.reuters.com/sitemap_%s-%s.xml", startDate.Format("20060102"), d.Format("20060102"))
		go func(url string) error {
			defer t.Done(nil)
			cmd := exec.Command("wget", "-nc", "--quiet", url)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			fmt.Println("#> Executing: ", strings.Join(cmd.Args, " "))
			err := cmd.Run()
			if err != nil {
				return err
			}
			return nil
		}(u)

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
}
