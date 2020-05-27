package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/k0kubun/pp"
	"github.com/spf13/pflag"

	"github.com/lucmichalski/finance-dataset/pkg/wordpress"
)

var (
	username string
	password string
	endpoint string
	help     bool
)

func main() {

	// read .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pflag.StringVarP(&endpoint, "endpoint", "", os.Getenv("WORDPRESS_API_ENDPOINT"), "wordpress api endpoint (eg. https://x0rzkov.com/wp-json).")
	pflag.BoolVarP(&help, "help", "h", false, "help info")
	pflag.Parse()
	if help {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// create wp-api client
	client, _ := wordpress.NewClient(endpoint, nil)

	ctx := context.Background()

	// lits of posts
	posts, _, err := client.Posts.List(ctx, nil)
	checkErr(err)
	pp.Println(err)

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
