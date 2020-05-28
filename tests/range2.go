package main

import (
	"fmt"
	"time"
)

func main() {
	//start := time.Now()
	start := time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Now() // start.AddDate(, 0, 0)
	for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
		startDate := d.AddDate(0, 0, -1)
		u := fmt.Printf("https://www.reuters.com/sitemap_%s-%s.xml\n", startDate.Format("20060102"), d.Format("20060102"))

		cmd := exec.Command("wget", "-nc", "--silent", u)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		fmt.Println("#> Executing: ", strings.Join(cmd.Args, " "))
		err := cmd.Run()
		if err != nil {
			log.Println("err:", err)
			return err
		}

	}

}
