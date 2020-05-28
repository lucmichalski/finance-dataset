package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/finance-contrib/scmp.com/admin"
	"github.com/lucmichalski/finance-contrib/scmp.com/crawler"
	"github.com/lucmichalski/finance-contrib/scmp.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingScmp{},
}

var Resources = []interface{}{
	&models.SettingScmp{},
}

type scmpPlugin string

func (o scmpPlugin) Name() string      { return string(o) }
func (o scmpPlugin) Section() string   { return `scmp.com` }
func (o scmpPlugin) Usage() string     { return `` }
func (o scmpPlugin) ShortDesc() string { return `scmp.com crawler"` }
func (o scmpPlugin) LongDesc() string  { return o.ShortDesc() }

func (o scmpPlugin) Migrate() []interface{} {
	return Tables
}

func (o scmpPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o scmpPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o scmpPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.scmp.com", "scmp.com"},
		URLs: []string{
			"https://www.scmp.com/sitemap.xml",
			"https://www.scmp.com/sitemap_explained.xml",
			"https://www.scmp.com/sitemap_podcasts.xml",
			"https://www.scmp.com/sitemap_announcements.xml",
			"https://www.scmp.com/sitemap_infographics.xml",
			"https://www.scmp.com/sitemap_news.xml",
			"https://www.scmp.com/sitemap_economy.xml",
			"https://www.scmp.com/sitemap_business.xml",
			"https://www.scmp.com/sitemap_comment.xml",
			"https://www.scmp.com/sitemap_tech.xml",
			"https://www.scmp.com/sitemap_lifestyle.xml",
			"https://www.scmp.com/sitemap_culture.xml",
			"https://www.scmp.com/sitemap_sport.xml",
			"https://www.scmp.com/sitemap_property.xml",
			"https://www.scmp.com/sitemap_photos.xml",
			"https://www.scmp.com/sitemap_video.xml",
			"https://www.scmp.com/sitemap_destination_macau.xml",
			"https://www.scmp.com/sitemap_magazines.xml",
			"https://www.scmp.com/sitemap_this_week_in_asia.xml",
			"https://www.scmp.com/sitemap_directories.xml",
			"https://www.scmp.com/sitemap_weather.xml",
			"https://www.scmp.com/sitemap_about_us.xml",
			"https://www.scmp.com/sitemap_lists.xml",
			"https://www.scmp.com/sitemap_special_reports.xml",
			"https://www.scmp.com/sitemap_country_reports.xml",
			"https://www.scmp.com/sitemap_video_comments.xml",
			"https://www.scmp.com/sitemap_video_scmp_originals.xml",
			"https://www.scmp.com/sitemap_video_hong_kong.xml",
			"https://www.scmp.com/sitemap_video_china.xml",
			"https://www.scmp.com/sitemap_video_asia.xml",
			"https://www.scmp.com/sitemap_video_world.xml",
			"https://www.scmp.com/sitemap_video_business.xml",
			"https://www.scmp.com/sitemap_video_arts_culture.xml",
			"https://www.scmp.com/sitemap_video_technology.xml",
			"https://www.scmp.com/sitemap_video_lifestyle.xml",
			"https://www.scmp.com/sitemap_video_sport.xml",
			"https://www.scmp.com/sitemap_video_offbeat.xml",
			"https://www.scmp.com/sitemap_video_style.xml",
			"https://www.scmp.com/sitemap_video_post_mag.xml",
			"https://www.scmp.com/sitemap_video_presented.xml",
			"https://www.scmp.com/sitemap_article.xml",
			"https://www.scmp.com/sitemap_gallery.xml",
			"https://www.scmp.com/sitemap_poll.xml",
			"https://www.scmp.com/sitemap_promotion.xml",
			"https://www.scmp.com/sitemap_webform.xml",
			"https://www.scmp.com/sitemap_video_format.xml",
			"https://www.scmp.com/sitemap_sections.xml",
			"https://www.scmp.com/sitemap_topics.xml",
			"https://www.scmp.com/sitemap_authors.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  true,
	}
	return cfg
}

type scmpCommands struct{}

func (t *scmpCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------
scmp.com
`)

	return nil
}

func (t *scmpCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"scmp": scmpPlugin("scmp"), //OP
	}
}

var Plugins scmpCommands
