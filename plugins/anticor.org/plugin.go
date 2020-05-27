package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/anticor.org/admin"
	"github.com/lucmichalski/finance-contrib/anticor.org/crawler"
	"github.com/lucmichalski/finance-contrib/anticor.org/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingAnticorOrg{},
}

var Resources = []interface{}{
	&models.SettingAnticorOrg{},
}

type anticorOrgPlugin string

func (o anticorOrgPlugin) Name() string      { return string(o) }
func (o anticorOrgPlugin) Section() string   { return `anticor.org` }
func (o anticorOrgPlugin) Usage() string     { return `` }
func (o anticorOrgPlugin) ShortDesc() string { return `anticor.org crawler"` }
func (o anticorOrgPlugin) LongDesc() string  { return o.ShortDesc() }

func (o anticorOrgPlugin) Migrate() []interface{} {
	return Tables
}

func (o anticorOrgPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o anticorOrgPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o anticorOrgPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.anticor.org", "anticor.org"},
		URLs: []string{
			"https://www.anticor.org/sitemap_index.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 6,
		IsSitemapIndex:  true,
	}
	return cfg
}

type anticorOrgCommands struct{}

func (t *anticorOrgCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------
anticor.org
`)

	return nil
}

func (t *anticorOrgCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"anticorOrg": anticorOrgPlugin("anticorOrg"), //OP
	}
}

var Plugins anticorOrgCommands
