package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/fcpablog.com/admin"
	"github.com/lucmichalski/finance-contrib/fcpablog.com/crawler"
	"github.com/lucmichalski/finance-contrib/fcpablog.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingDevex{},
}

var Resources = []interface{}{
	&models.SettingDevex{},
}

type fcpaBlogPlugin string

func (o fcpaBlogPlugin) Name() string      { return string(o) }
func (o fcpaBlogPlugin) Section() string   { return `fcpablog.com` }
func (o fcpaBlogPlugin) Usage() string     { return `` }
func (o fcpaBlogPlugin) ShortDesc() string { return `fcpablog.com crawler"` }
func (o fcpaBlogPlugin) LongDesc() string  { return o.ShortDesc() }

func (o fcpaBlogPlugin) Migrate() []interface{} {
	return Tables
}

func (o fcpaBlogPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o fcpaBlogPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o fcpaBlogPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.fcpablog.com", "fcpablog.com"},
		URLs: []string{
			"https://fcpablog.com/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  true,
	}
	return cfg
}

type fcpaBlogCommands struct{}

func (t *fcpaBlogCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------
fcpaBlog
`)

	return nil
}

func (t *fcpaBlogCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"fcpaBlog": fcpaBlogPlugin("fcpaBlog"), //OP
	}
}

var Plugins fcpaBlogCommands
