package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/barrons.com/admin"
	"github.com/lucmichalski/finance-contrib/barrons.com/crawler"
	"github.com/lucmichalski/finance-contrib/barrons.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingBarrons{},
}

var Resources = []interface{}{
	&models.SettingBarrons{},
}

type barronsPlugin string

func (o barronsPlugin) Name() string      { return string(o) }
func (o barronsPlugin) Section() string   { return `barrons.com` }
func (o barronsPlugin) Usage() string     { return `` }
func (o barronsPlugin) ShortDesc() string { return `barrons.com crawler"` }
func (o barronsPlugin) LongDesc() string  { return o.ShortDesc() }

func (o barronsPlugin) Migrate() []interface{} {
	return Tables
}

func (o barronsPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o barronsPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o barronsPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.barrons.com", "barrons.com"},
		URLs: []string{
			"https://www.barrons.com/bol_news_sitemap.xml",
			"https://www.barrons.com/sitemaps/web/barrons/en/sitemap_barrons_en_index.xml",
			"https://www.barrons.com/sitemaps/web/barrons/afp_news/sitemap_barrons_afp_news_index.xml",
			// "https://www.barrons.com/sitemaps/web/barrons-video/en/sitemap_barrons-video_en_index.xml",
			"https://www.barrons.com/sitemap.xml",
			// "https://www.barrons.com/quote/stock/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  false,
	}
	return cfg
}

type barronsCommands struct{}

func (t *barronsCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
---------------------------------------------------------------------------------------------------------------
'########:::::'###::::'########::'########:::'#######::'##::: ##::'######::::::::'######:::'#######::'##::::'##:
 ##.... ##:::'## ##::: ##.... ##: ##.... ##:'##.... ##: ###:: ##:'##... ##::::::'##... ##:'##.... ##: ###::'###:
 ##:::: ##::'##:. ##:: ##:::: ##: ##:::: ##: ##:::: ##: ####: ##: ##:::..::::::: ##:::..:: ##:::: ##: ####'####:
 ########::'##:::. ##: ########:: ########:: ##:::: ##: ## ## ##:. ######::::::: ##::::::: ##:::: ##: ## ### ##:
 ##.... ##: #########: ##.. ##::: ##.. ##::: ##:::: ##: ##. ####::..... ##:::::: ##::::::: ##:::: ##: ##. #: ##:
 ##:::: ##: ##.... ##: ##::. ##:: ##::. ##:: ##:::: ##: ##:. ###:'##::: ##:'###: ##::: ##: ##:::: ##: ##:.:: ##:
 ########:: ##:::: ##: ##:::. ##: ##:::. ##:. #######:: ##::. ##:. ######:: ###:. ######::. #######:: ##:::: ##:
........:::..:::::..::..:::::..::..:::::..:::.......:::..::::..:::......:::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *barronsCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"barrons": barronsPlugin("barrons"), //OP
	}
}

var Plugins barronsCommands
