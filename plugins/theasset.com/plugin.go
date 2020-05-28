package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/theasset.com/admin"
	"github.com/lucmichalski/finance-contrib/theasset.com/crawler"
	"github.com/lucmichalski/finance-contrib/theasset.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingTheAsset{},
}

var Resources = []interface{}{
	&models.SettingTheAsset{},
}

type theAssetPlugin string

func (o theAssetPlugin) Name() string      { return string(o) }
func (o theAssetPlugin) Section() string   { return `theasset.com` }
func (o theAssetPlugin) Usage() string     { return `` }
func (o theAssetPlugin) ShortDesc() string { return `theasset.com crawler"` }
func (o theAssetPlugin) LongDesc() string  { return o.ShortDesc() }

func (o theAssetPlugin) Migrate() []interface{} {
	return Tables
}

func (o theAssetPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o theAssetPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o theAssetPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains:  []string{"www.theasset.com", "theasset.com"},
		URLs:            []string{
					"https://theasset.com/",
					"https://theasset.com/section/wealth-management",
					"https://theasset.com/section/asia-connect",
					"https://theasset.com/section/treasury-capital-markets",
					"https://theasset.com/section/europe",
					"https://theasset.com/section/esg-forum",
					"https://theasset.com/section/covid-19",
					"https://theasset.com/section/on-the-move",
					"https://theasset.com/today-in-china",
				 },
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  false,
	}
	return cfg
}

type theAssetCommands struct{}

func (t *theAssetCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------
TheAsset
`)

	return nil
}

func (t *theAssetCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"theAsset": theAssetPlugin("theAsset"), //OP
	}
}

var Plugins theAssetCommands
