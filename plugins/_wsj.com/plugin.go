package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/wsj.com/admin"
	"github.com/lucmichalski/finance-contrib/wsj.com/crawler"
	"github.com/lucmichalski/finance-contrib/wsj.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingWsj{},
}

var Resources = []interface{}{
	&models.SettingWsj{},
}

type wsjPlugin string

func (o wsjPlugin) Name() string      { return string(o) }
func (o wsjPlugin) Section() string   { return `wsj.com` }
func (o wsjPlugin) Usage() string     { return `` }
func (o wsjPlugin) ShortDesc() string { return `wsj.com crawler"` }
func (o wsjPlugin) LongDesc() string  { return o.ShortDesc() }

func (o wsjPlugin) Migrate() []interface{} {
	return Tables
}

func (o wsjPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o wsjPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o wsjPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.wsj.com", "wsj.com"},
		URLs: []string{
			"https://www.wsj.com/sitemaps/web/wsj/en/sitemap_wsj_en_index.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 6,
		IsSitemapIndex:  true,
	}
	return cfg
}

type wsjCommands struct{}

func (t *wsjCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------
wsj.com
`)

	return nil
}

func (t *wsjCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"wsj": wsjPlugin("wsj"), //OP
	}
}

var Plugins wsjCommands
