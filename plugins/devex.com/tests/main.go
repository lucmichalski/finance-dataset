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

	"github.com/lucmichalski/finance-contrib/devex.com/crawler"
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
		AllowedDomains: []string{"www.devex.com", "devex.com"},
		URLs: []string{
			"https://www.devex.com/news/exclusive-state-department-makes-bid-for-us-global-pandemic-response-powers-97315",
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
