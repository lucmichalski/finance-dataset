package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/business.financialpost.com/admin"
	"github.com/lucmichalski/finance-contrib/business.financialpost.com/crawler"
	"github.com/lucmichalski/finance-contrib/business.financialpost.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingBusinessFinancialPost{},
}

var Resources = []interface{}{
	&models.SettingBusinessFinancialPost{},
}

type businessFinancialPostPlugin string

func (o businessFinancialPostPlugin) Name() string      { return string(o) }
func (o businessFinancialPostPlugin) Section() string   { return `business.financialpost.com` }
func (o businessFinancialPostPlugin) Usage() string     { return `` }
func (o businessFinancialPostPlugin) ShortDesc() string { return `business.financialpost.com crawler"` }
func (o businessFinancialPostPlugin) LongDesc() string  { return o.ShortDesc() }

func (o businessFinancialPostPlugin) Migrate() []interface{} {
	return Tables
}

func (o businessFinancialPostPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o businessFinancialPostPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o businessFinancialPostPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"business.financialpost.com"},
		URLs: []string{
			//"https://business.financialpost.com/news-sitemap.xml",
			"https://business.financialpost.com/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  true,
	}
	return cfg
}

type businessFinancialPostCommands struct{}

func (t *businessFinancialPostCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
------------------------------------------------------------------------------------------------------------------------------------------------------------------
'########:'####:'##::: ##::::'###::::'##::: ##::'######::'####::::'###::::'##:::::::'########:::'#######:::'######::'########:::::::'######:::'#######::'##::::'##:
 ##.....::. ##:: ###:: ##:::'## ##::: ###:: ##:'##... ##:. ##::::'## ##::: ##::::::: ##.... ##:'##.... ##:'##... ##:... ##..:::::::'##... ##:'##.... ##: ###::'###:
 ##:::::::: ##:: ####: ##::'##:. ##:: ####: ##: ##:::..::: ##:::'##:. ##:: ##::::::: ##:::: ##: ##:::: ##: ##:::..::::: ##::::::::: ##:::..:: ##:::: ##: ####'####:
 ######:::: ##:: ## ## ##:'##:::. ##: ## ## ##: ##:::::::: ##::'##:::. ##: ##::::::: ########:: ##:::: ##:. ######::::: ##::::::::: ##::::::: ##:::: ##: ## ### ##:
 ##...::::: ##:: ##. ####: #########: ##. ####: ##:::::::: ##:: #########: ##::::::: ##.....::: ##:::: ##::..... ##:::: ##::::::::: ##::::::: ##:::: ##: ##. #: ##:
 ##:::::::: ##:: ##:. ###: ##.... ##: ##:. ###: ##::: ##:: ##:: ##.... ##: ##::::::: ##:::::::: ##:::: ##:'##::: ##:::: ##::::'###: ##::: ##: ##:::: ##: ##:.:: ##:
 ##:::::::'####: ##::. ##: ##:::: ##: ##::. ##:. ######::'####: ##:::: ##: ########: ##::::::::. #######::. ######::::: ##:::: ###:. ######::. #######:: ##:::: ##:
..::::::::....::..::::..::..:::::..::..::::..:::......:::....::..:::::..::........::..::::::::::.......::::......::::::..:::::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *businessFinancialPostCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"businessFinancialPost": businessFinancialPostPlugin("businessFinancialPost"), //OP
	}
}

var Plugins businessFinancialPostCommands
