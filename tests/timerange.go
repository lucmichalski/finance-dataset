package main

import (
	"fmt"
	"time"

	"github.com/tejasmanohar/timerange-go"
)

func main() {
	start := time.Date(2016, 8, 28, 9, 0, 0, 0, time.UTC)
	end := time.Now() // time.Date(2016, 8, 28, 11, 0, 0, 0, time.UTC)
	iter := timerange.New(start, end, time.Days)
	for iter.Next() {
		t := iter.Current()
		fmt.Println(t)
	}
}
