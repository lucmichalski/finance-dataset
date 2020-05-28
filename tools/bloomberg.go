package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/k0kubun/pp"
	"github.com/karrick/godirwalk"
	"github.com/qor/media"
	"github.com/qor/validations"
	log "github.com/sirupsen/logrus"

	"github.com/lucmichalski/finance-dataset/pkg/models"
)

var (
	isDryMode = true
)

const (
	datasetAbsPath = `../shared/dataset/20061020_20131126_bloomberg_news`
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

	walkImages(DB, datasetAbsPath)
}

func walkImages(DB *gorm.DB, dirnames ...string) (err error) {
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					// process file
					pp.Println("osPathname:", osPathname)
					err := openTextFile(DB, osPathname)
					checkErr(err)
				}
				return nil
			},
			Unsorted: true,
		})
	}
	return
}

func cleanTextFile(osPathname string) error {
	read, err := ioutil.ReadFile(osPathname)
	if err != nil {
		return err
	}
	//fmt.Println(string(read))
	// fmt.Println(osPathname)
	newContents := strings.Replace(string(read), "-- \n", "--", -1)
	// fmt.Println(newContents)
	err = ioutil.WriteFile(osPathname, []byte(newContents), 0)
	if err != nil {
		return err
	}
	return nil
}

func fixAuthorName(input string) string {
	input = strings.Replace(input, "A n d", ",", -1)
	input = strings.Replace(input, "B y  ", "", -1)
	authors := strings.Split(input, "  ")
	var auths []string
	for _, partName := range authors {
		partName = strings.Replace(partName, " ", "", -1)
		auths = append(auths, partName)
	}
	return strings.Join(auths, " ")
}

func openTextFile(DB *gorm.DB, osPathname string) error {
	file, err := os.Open(osPathname)

	if err != nil {
		// log.Fatalf("failed opening file: %s", err)
		return err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	file.Close()

	if len(txtlines) < 4 {
		return errors.New("abnormal input")
	}

	// title
	title := txtlines[0]
	title = strings.Replace(title, "-- ", "", -1)
	title = strings.TrimSpace(title)

	// author
	author := txtlines[1]
	author = strings.Replace(author, "-- ", "", -1)
	author = strings.TrimSpace(author)
	if strings.Contains(author, "B y  ") {
		author = fixAuthorName(author)
	}

	// fix author's name

	// date
	published := txtlines[2]
	published = strings.Replace(published, "-- ", "", -1)
	published = strings.TrimSpace(published)
	publishedAtTime, err := dateparse.ParseAny(published)
	if err != nil {
		return err
	}

	// link
	link := txtlines[3]
	link = strings.Replace(link, "-- ", "", -1)
	link = strings.Replace(link, "http:", "https:", -1)
	link = strings.Replace(link, ".html", "", -1)
	link = strings.Replace(link, "bloomberg.com/news/2", "bloomberg.com/news/articles/2", -1)
	link = strings.TrimSpace(link)

	// content
	content := txtlines[4:]

	/*
		fmt.Println("title: ", title)
		fmt.Println("author: ", author)
		fmt.Println("published: ", publishedAtTime)
		fmt.Println("link: ", link)
		fmt.Println("content: ", content)
		fmt.Println("============================================================================================================")
	*/

	// check if exists
	var pageExists models.Page
	if !DB.Where("link = ?", link).First(&pageExists).RecordNotFound() {
		fmt.Printf("skipping url=%s as already exists\n", link)
		return nil
	}

	// insert into database
	page := &models.Page{}
	page.Link = link
	page.Source = "bloomberg.com"
	page.Class = "article"
	page.Title = title
	page.Authors = author
	page.Content = strings.Join(content, "")
	page.PublishedAt = publishedAtTime

	if page.Title == "" && page.Link == "" && page.Content == "" && page.PublishedAt.String() == "" {
		return nil
	}

	pp.Println(page)

	if err := DB.Create(&page).Error; err != nil {
		log.Fatalf("create page (%v) failure, got err %v", page, err)
		return err
	}

	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Warn(err)
	}
}
