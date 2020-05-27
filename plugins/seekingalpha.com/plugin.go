package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/seekingalpha.com/admin"
	"github.com/lucmichalski/finance-contrib/seekingalpha.com/crawler"
	"github.com/lucmichalski/finance-contrib/seekingalpha.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingSeekingAlpha{},
}

var Resources = []interface{}{
	&models.SettingSeekingAlpha{},
}

type seekingAlphaPlugin string

func (o seekingAlphaPlugin) Name() string      { return string(o) }
func (o seekingAlphaPlugin) Section() string   { return `seekingalpha.com` }
func (o seekingAlphaPlugin) Usage() string     { return `` }
func (o seekingAlphaPlugin) ShortDesc() string { return `seekingalpha.com crawler"` }
func (o seekingAlphaPlugin) LongDesc() string  { return o.ShortDesc() }

func (o seekingAlphaPlugin) Migrate() []interface{} {
	return Tables
}

func (o seekingAlphaPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o seekingAlphaPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o seekingAlphaPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.seekingalpha.com", "seekingalpha.com"},
		URLs: []string{
			"https://seekingalpha.com/instablog/index.xml",
			"https://seekingalpha.com/news/index.xml",
			"https://seekingalpha.com/article/index.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 6,
		IsSitemapIndex:  true,
	}
	return cfg
}

type seekingAlphaCommands struct{}

func (t *seekingAlphaCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------
seekingalpha.com
`)

	return nil
}

func (t *seekingAlphaCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"seekingAlpha": seekingAlphaPlugin("seekingAlpha"), //OP
	}
}

var Plugins seekingAlphaCommands
