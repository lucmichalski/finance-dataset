package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	file, err := os.Open("/home/ubuntu/lucmichalski/finance-dataset/shared/dataset/20061020_20131126_bloomberg_news/2011-03-10/plotech-co-ltd-february-sales-fall-12-51-table-6141-")

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
}
