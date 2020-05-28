package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/k0kubun/pp"
	"github.com/karrick/godirwalk"
	log "github.com/sirupsen/logrus"
)

var (
	isDryMode = true
)

const (
	datasetAbsPath = `/home/ubuntu/lucmichalski/finance-dataset/shared/dataset/20061020_20131126_bloomberg_news`
)

func main() {

	walkImages(datasetAbsPath)
}

func walkImages(dirnames ...string) (err error) {
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					// process file
					pp.Println("osPathname:", osPathname)

				}
				return nil
			},
			Unsorted: true,
		})
	}
	return
}

func openTextFile(osPathname string) error {
	file, err := os.Open(osPathname)

	if err != nil {
		// log.Fatalf("failed opening file: %s", err)
		return err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	file.Close()

	/*
		------------------------------------------------------------------------------------------------------

		--
		SHIH HER TECH February Sales Rise 11.08% (Table) : 3551 TT

		-- B y   J a n e t   O n g
		--
		2011-03-09T03:16:07Z

		-- http://www.bloomberg.com/news/2011-03-09/shih-her-tech-february-sales-rise-11-08-table-3551-tt.html

		------------------------------------------------------------------------------------------------------

		--
		UNILITE CORPORAT February Sales Fall 8.14% (Table) : 3517 TT

		-- B y   J a n e t   O n g
		--
		2011-03-09T02:38:09Z

		-- http://www.bloomberg.com/news/2011-03-09/unilite-corporat-february-sales-fall-8-14-table-3517-tt.html

		------------------------------------------------------------------------------------------------------
	*/

	// txtlines[0]
	// txtlines[1]
	// txtlines[2]
	// txtlines[3]

	for _, eachline := range txtlines {
		// check if starts
		fmt.Println(eachline)
	}
	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
