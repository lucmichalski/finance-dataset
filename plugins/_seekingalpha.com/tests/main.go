package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/qor/media"
	"github.com/qor/validations"

	"github.com/lucmichalski/finance-dataset/pkg/config"
	"github.com/lucmichalski/finance-dataset/pkg/models"

	"github.com/lucmichalski/finance-contrib/seekingalpha.com/crawler"
)

func main() {

	DB, err := gorm.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4,utf8&parseTime=True", os.Getenv("FD_MYSQL_USER"), os.Getenv("FD_MYSQL_PASSWORD"), os.Getenv("FD_MYSQL_HOST"), os.Getenv("FD_MYSQL_PORT"), os.Getenv("FD_MYSQL_DATABASE")))
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// callback for images and validation
	validations.RegisterCallbacks(DB)
	media.RegisterCallbacks(DB)

	// migrate tables
	DB.AutoMigrate(&models.Page{})

	cfg := &config.Config{
		AllowedDomains: []string{"www.seekingalpha.com", "seekingalpha.com"},
		URLs: []string{
			// "https://seekingalpha.com/article/4350197-only-lonely-can-play",
			// "https://seekingalpha.com/article/2-mexico-holds-rates-for-4th-month-as-inflation-slows",
			"https://seekingalpha.com/news/3529014-fda-set-to-ban-all-e-cig-pod-flavors",
		},
		DB:              DB,
		CacheDir:        "../../../shared/data",
		QueueMaxSize:    1000000,
		ConsumerThreads: 35,
		DryMode:         true,
		IsDebug:         true,
	}

	err = crawler.Extract(cfg)
	if err != nil {
		log.Fatal(err)
	}

}
