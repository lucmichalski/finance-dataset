package crawler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/nozzle/throttler"
	log "github.com/sirupsen/logrus"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/models"
	"github.com/lucmichalski/finance-dataset/pkg/wordpress"
)

func Extract(cfg *config.Config) error {

	// create wp-api client
	client, _ := wordpress.NewClient(cfg.URLs[0], nil)
	ctx := context.Background()

	// init throttle
	t := throttler.New(3, 1000000)

	opts := &wordpress.PostListOptions{}
	// page := -1
	for {

		// lits of posts
		// opts.Offset = 0
		opts.ListOptions.PerPage = 100
		// opts.ListOptions.Page = 0

		pp.Println("opts:", opts)

		posts, _, err := client.Posts.List(ctx, opts)
		checkErr(err)
		opts.ListOptions.Page++
		// opts.Offset = opts.Offset + 1

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

			go func(post *wordpress.Post) error {
				defer t.Done(nil)

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
					return errors.New("not enough attributs for registration into the db")
				}

				if cfg.IsDebug {
					pp.Println("page:", page)
				}

				if !cfg.DryMode {
					if err := cfg.DB.Create(&page).Error; err != nil {
						log.Warnf("create page (%v) failure, got err %v", page, err)
						return err
					}
				}
				return nil
			}(post)
			t.Throttle()
		}
	}

	// throttler errors iteration
	if t.Err() != nil {
		// Loop through the errors to see the details
		for i, err := range t.Errs() {
			log.Printf("error #%d: %s", i, err)
		}
		log.Fatal(t.Err())
	}

	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
