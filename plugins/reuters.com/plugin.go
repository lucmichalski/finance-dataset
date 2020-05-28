package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/reuters.com/admin"
	"github.com/lucmichalski/finance-contrib/reuters.com/crawler"
	"github.com/lucmichalski/finance-contrib/reuters.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingReuters{},
}

var Resources = []interface{}{
	&models.SettingReuters{},
}

type reutersPlugin string

func (o reutersPlugin) Name() string      { return string(o) }
func (o reutersPlugin) Section() string   { return `reuters.com` }
func (o reutersPlugin) Usage() string     { return `` }
func (o reutersPlugin) ShortDesc() string { return `reuters.com crawler"` }
func (o reutersPlugin) LongDesc() string  { return o.ShortDesc() }

func (o reutersPlugin) Migrate() []interface{} {
	return Tables
}

func (o reutersPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o reutersPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o reutersPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.reuters.com", "reuters.com"},
		URLs: []string{
			"https://www.reuters.com/sitemap_20200426-20200427.xml",
			//"https://www.reuters.com/sitemap_index.xml",
			//"https://www.reuters.com/sitemap_news_index.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  false,
	}
	return cfg
}

type reutersCommands struct{}

func (t *reutersCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------
reuters.com
`)

	return nil
}

func (t *reutersCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"reuters": reutersPlugin("reuters"), //OP
	}
}

var Plugins reutersCommands
