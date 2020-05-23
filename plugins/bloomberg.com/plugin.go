package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/bloomberg.com/admin"
	"github.com/lucmichalski/finance-contrib/bloomberg.com/crawler"
	"github.com/lucmichalski/finance-contrib/bloomberg.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingBloomberg{},
}

var Resources = []interface{}{
	&models.SettingBloomberg{},
}

type bloombergPlugin string

func (o bloombergPlugin) Name() string      { return string(o) }
func (o bloombergPlugin) Section() string   { return `bloomberg.com` }
func (o bloombergPlugin) Usage() string     { return `` }
func (o bloombergPlugin) ShortDesc() string { return `bloomberg.com crawler"` }
func (o bloombergPlugin) LongDesc() string  { return o.ShortDesc() }

func (o bloombergPlugin) Migrate() []interface{} {
	return Tables
}

func (o bloombergPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o bloombergPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o bloombergPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.bloomberg.com", "bloomberg.com"},
		URLs: []string{
			"https://www.bloomberg.com/secure/sitemap/s_sitemap_index.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  true,
	}
	return cfg
}

type bloombergCommands struct{}

func (t *bloombergCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-------------------------------------------------------------------------------------------------------------------------------------
'########::'##::::::::'#######:::'#######::'##::::'##:'########::'########:'########:::'######:::::::::'######:::'#######::'##::::'##:
 ##.... ##: ##:::::::'##.... ##:'##.... ##: ###::'###: ##.... ##: ##.....:: ##.... ##:'##... ##:::::::'##... ##:'##.... ##: ###::'###:
 ##:::: ##: ##::::::: ##:::: ##: ##:::: ##: ####'####: ##:::: ##: ##::::::: ##:::: ##: ##:::..:::::::: ##:::..:: ##:::: ##: ####'####:
 ########:: ##::::::: ##:::: ##: ##:::: ##: ## ### ##: ########:: ######::: ########:: ##::'####:::::: ##::::::: ##:::: ##: ## ### ##:
 ##.... ##: ##::::::: ##:::: ##: ##:::: ##: ##. #: ##: ##.... ##: ##...:::: ##.. ##::: ##::: ##::::::: ##::::::: ##:::: ##: ##. #: ##:
 ##:::: ##: ##::::::: ##:::: ##: ##:::: ##: ##:.:: ##: ##:::: ##: ##::::::: ##::. ##:: ##::: ##::'###: ##::: ##: ##:::: ##: ##:.:: ##:
 ########:: ########:. #######::. #######:: ##:::: ##: ########:: ########: ##:::. ##:. ######::: ###:. ######::. #######:: ##:::: ##:
........:::........:::.......::::.......:::..:::::..::........:::........::..:::::..:::......::::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *bloombergCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"bloomberg": bloombergPlugin("bloomberg"), //OP
	}
}

var Plugins bloombergCommands
