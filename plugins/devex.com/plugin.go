package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/devex.com/admin"
	"github.com/lucmichalski/finance-contrib/devex.com/crawler"
	"github.com/lucmichalski/finance-contrib/devex.com/models"
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

type devexPlugin string

func (o devexPlugin) Name() string      { return string(o) }
func (o devexPlugin) Section() string   { return `devex.com` }
func (o devexPlugin) Usage() string     { return `` }
func (o devexPlugin) ShortDesc() string { return `devex.com crawler"` }
func (o devexPlugin) LongDesc() string  { return o.ShortDesc() }

func (o devexPlugin) Migrate() []interface{} {
	return Tables
}

func (o devexPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o devexPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o devexPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.devex.com", "devex.com"},
		URLs: []string{
			"http://devexsitemap.s3.amazonaws.com/sitemaps/companies-sitemap.xml.gz",
			"http://devexsitemap.s3.amazonaws.com/sitemaps/procurement-sitemap.xml.gz",
			"http://devexsitemap.s3.amazonaws.com/sitemaps/news-sitemap.xml.gz",
			"http://devexsitemap.s3.amazonaws.com/sitemaps/google-news-sitemap.xml.gz",
			"http://devexsitemap.s3.amazonaws.com/sitemaps/people-sitemap.xml.gz",
			"http://devexsitemap.s3.amazonaws.com/sitemaps/jobs-sitemap.xml.gz",
			"http://devexsitemap.s3.amazonaws.com/sitemaps/routes-sitemap.xml.gz",
			"http://devexsitemap.s3.amazonaws.com/sitemaps/countries-sitemap.xml.gz",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  true,
	}
	return cfg
}

type devexCommands struct{}

func (t *devexCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------
'########::'########:'##::::'##:'########:'##::::'##:::::::'######:::'#######::'##::::'##:
 ##.... ##: ##.....:: ##:::: ##: ##.....::. ##::'##:::::::'##... ##:'##.... ##: ###::'###:
 ##:::: ##: ##::::::: ##:::: ##: ##::::::::. ##'##:::::::: ##:::..:: ##:::: ##: ####'####:
 ##:::: ##: ######::: ##:::: ##: ######:::::. ###::::::::: ##::::::: ##:::: ##: ## ### ##:
 ##:::: ##: ##...::::. ##:: ##:: ##...:::::: ## ##:::::::: ##::::::: ##:::: ##: ##. #: ##:
 ##:::: ##: ##::::::::. ## ##::: ##:::::::: ##:. ##::'###: ##::: ##: ##:::: ##: ##:.:: ##:
 ########:: ########:::. ###:::: ########: ##:::. ##: ###:. ######::. #######:: ##:::: ##:
........:::........:::::...:::::........::..:::::..::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *devexCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"devex": devexPlugin("devex"), //OP
	}
}

var Plugins devexCommands
