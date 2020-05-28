package main

import (
	//"fmt"

	"github.com/k0kubun/pp"
	"github.com/aodin/date"
)

func main() {
	dr := date.NewRange(date.Today(), date.Today().AddDays(7))
	pp.Println(dr)
}
