package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abadojack/whatlanggo"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
)

type Page struct {
	gorm.Model
	Link               string `gorm:"index:link"`
	Title              string
	Content            string `gorm:"type:longtext; CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci" sql:"type:longtext"`
	Categories         string
	Tags               string
	Authors            string
	LanguageConfidence float64
	Language           string    `gorm:"index:language"`
	PublishedAt        time.Time `gorm:"index:published_at"`
	Source             string    `gorm:"index:source"`
	Class              string    `gorm:"index:class"`
	PageAttributes     []PageAttribute
	// PageProperties     PageProperties `sql:"type:text"`
}

func (p Page) Validate(db *gorm.DB) {
	if strings.TrimSpace(p.Title) == "" {
		db.AddError(validations.NewError(p, "Name", "Name can not be empty"))
	}
}

func (p *Page) BeforeCreate() (err error) {
	// add to whatlango
	if p.Content != "" {
		info := whatlanggo.Detect(p.Content)
		p.Language = info.Lang.String()
		p.LanguageConfidence = info.Confidence
		fmt.Println("======> Language:", p.Language, " Script:", whatlanggo.Scripts[info.Script], " Confidence: ", p.LanguageConfidence)
	}
	return
}

func (p *Page) AfterCreate() (err error) {
	// add to manticore
	// add to bleve
	return
}

type PageProperties []PageProperty

type PageProperty struct {
	Name  string
	Value string
}

func (pageProperties *PageProperties) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, pageProperties)
	case string:
		if v != "" {
			return pageProperties.Scan([]byte(v))
		}
	default:
		return errors.New("not supported")
	}
	return nil
}

func (pageProperties PageProperties) Value() (driver.Value, error) {
	if len(pageProperties) == 0 {
		return nil, nil
	}
	return json.Marshal(pageProperties)
}

type PageAttribute struct {
	gorm.Model
	PageID uint
	Name   string
	Value  string
}

func (p PageAttribute) Validate(db *gorm.DB) {
	if strings.TrimSpace(p.Name) == "" {
		db.AddError(validations.NewError(p, "Name", "Name can not be empty"))
	}
	if strings.TrimSpace(p.Value) == "" {
		db.AddError(validations.NewError(p, "Value", "Value can not be empty"))
	}
}
