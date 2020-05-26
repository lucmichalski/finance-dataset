package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/lesechos.fr/admin"
	"github.com/lucmichalski/finance-contrib/lesechos.fr/crawler"
	"github.com/lucmichalski/finance-contrib/lesechos.fr/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingLesEchos{},
}

var Resources = []interface{}{
	&models.SettingLesEchos{},
}

type lesEchosPlugin string

func (o lesEchosPlugin) Name() string      { return string(o) }
func (o lesEchosPlugin) Section() string   { return `lesechos.fr` }
func (o lesEchosPlugin) Usage() string     { return `` }
func (o lesEchosPlugin) ShortDesc() string { return `lesechos.fr crawler"` }
func (o lesEchosPlugin) LongDesc() string  { return o.ShortDesc() }

func (o lesEchosPlugin) Migrate() []interface{} {
	return Tables
}

func (o lesEchosPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o lesEchosPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o lesEchosPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"lesechos.fr", "www.lesechos.fr"},
		URLs: []string{
			"https://sitemap.lesechos.fr/sitemap_index.xml",
		},
		QueueMaxSize:    10000000,
		ConsumerThreads: 6,
		IsSitemapIndex:  true,
	}
	return cfg
}

type lesEchosCommands struct{}

func (t *lesEchosCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
------------------------------------------------------------------------------------------------------------------------------------------------------------------
LesEchos.fr
`)

	return nil
}

func (t *lesEchosCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"lesEchos": lesEchosPlugin("lesEchos"), //OP
	}
}

var Plugins lesEchosCommands
