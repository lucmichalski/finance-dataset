package crawler

import (
	"fmt"
	"strings"
	"context"

	"github.com/k0kubun/pp"
	// "github.com/nozzle/throttler"
	log "github.com/sirupsen/logrus"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/models"
	"github.com/lucmichalski/finance-dataset/pkg/wordpress"
)

func Extract(cfg *config.Config) error {

	// create wp-api client
	client, _ := wordpress.NewClient(cfg.URLs[0], nil)
	ctx := context.Background()

	for {

		// lits of posts
		opts := &wordpress.PostListOptions{}
		opts.Offset = 0
		opts.PerPage = 100

		posts, _, err := client.Posts.List(ctx, opts)
		checkErr(err)

		if len(posts) == 0 {
			break
		}

		for _, post := range posts {

			// check in the databse if exists
			var pageExists models.Page
			if !cfg.DryMode {
				if !cfg.DB.Where("link = ?", post.Link).First(&pageExists).RecordNotFound() {
					fmt.Printf("skipping url=%s as already exists\n", post.Link)
					continue
				}
			}

			page := &models.Page{}
			page.Link = post.Link
			page.Source = "fcpablog.com"
			page.Class = "post"

			if cfg.IsDebug {
				pp.Println("Title: ", post.Title.Rendered)
			}
			page.Title = post.Title.Rendered

			if cfg.IsDebug {
				pp.Println("Rendered: ", post.Content.Rendered)
			}
			page.Content = post.Content.Rendered

			if cfg.IsDebug {
				pp.Println("Date: ", post.Date.Time)
			}
			page.PublishedAt = post.Date.Time

			a, _, err := client.Users.Get(ctx, post.Author, nil)
			checkErr(err)
			if cfg.IsDebug {
				pp.Println("Author:", a.Name)
			}
			page.Authors = a.Name

			var cats []string
			for _, category := range post.Categories {
				c, _, err := client.Categories.Get(ctx, category, nil)
				checkErr(err)
				if cfg.IsDebug {
					pp.Println("Category:", c.Name)
				}
				if c.Name != "" {
					cats = append(cats, c.Name)
				}
			}
			page.Categories = strings.Join(cats, ",")

			var tags []string
			for _, tag := range post.Tags {
				t, _, err := client.Tags.Get(ctx, tag, nil)
				checkErr(err)
				if cfg.IsDebug {
					pp.Println("Tag:", t.Name)
				}
				if t.Name != "" {
					tags = append(tags, t.Name)
				}
			}
			page.Tags = strings.Join(tags, ",")

			if page.Link == "" && page.Content == "" && page.PublishedAt.String() == "" {
				continue
			}

			if cfg.IsDebug {
				pp.Println("page:", page)
			}

			if !cfg.DryMode {
				if err := cfg.DB.Create(&page).Error; err != nil {
					log.Warnf("create page (%v) failure, got err %v", page, err)
					continue
				}
			}
		}
		opts.Offset++

	}

	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
