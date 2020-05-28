package main

import (
	"bufio"
	"fmt"
	"os"
	// "strings"

	// "github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"

	"github.com/lucmichalski/finance-dataset/pkg/pluck"
)

func main() {

	fp := "/home/ubuntu/lucmichalski/finance-dataset/shared/dataset/20061020_20131126_bloomberg_news/2011-03-10/plotech-co-ltd-february-sales-fall-12-51-table-6141-"

	file, err := os.Open(fp)

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	file.Close()

	for idx, eachline := range txtlines {
		fmt.Println(idx, "=", eachline)
	}

	p, err := pluck.New()
	checkErr(err)

	p.Add(pluck.Config{
		Activators:  []string{"--"},
		Deactivator: "--",
		Limit:       -1,
		Sanitize:    false,
		// Finisher:    c.GlobalString("finisher"),
		// Permanent:   0,
	})

	err = p.PluckFile(fp)
	checkErr(err)
	result := p.ResultJSON(true)

	fmt.Println(result)

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
