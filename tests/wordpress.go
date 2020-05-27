package main

import (
	"context"
	"log"
	"os"

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

	pflag.StringVarP(&endpoint, "endpoint", "", os.Getenv("WORDPRESS_API_ENDPOINT"), "wordpress api endpoint (eg. https://x0rzkov.com/wp-json).")
	pflag.BoolVarP(&help, "help", "h", false, "help info")
	pflag.Parse()
	if help {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// v=spf1 include:amazonses.com ~all

	// create wp-api client
	client, _ := wordpress.NewClient(endpoint, nil)

	ctx := context.Background()

	// lits of posts
	opts := &wordpress.PostListOptions{}
	opts.Offset = 0
	opts.PerPage = 100

	posts, _, err := client.Posts.List(ctx, opts)
	checkErr(err)
	for _, post := range posts {
		pp.Println("Rendered: ", post.Content.Rendered)
		pp.Println("Date: ", post.Date.Time)
		a, _, err := client.Users.Get(ctx, post.Author, nil)
		pp.Println("Author:", a.Name)
		checkErr(err)
		for _, category := range post.Categories {
			c, _, err := client.Categories.Get(ctx, category, nil)
			pp.Println("Category:", c.Name)
			checkErr(err)
		}
		for _, tag := range post.Tags {
			c, _, err := client.Tags.Get(ctx, tag, nil)
			pp.Println("Tag:", c.Name)
			checkErr(err)
		}
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
