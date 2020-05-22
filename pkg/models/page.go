package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/media_library"
	"github.com/qor/validations"
)

type Page struct {
	gorm.Model
	Domain             string `gorm:"index:domain"`
	URL                string `gorm:"index:url"`
	Title              string
	Content            string
	Language           string
	LanguageConfidence float64
	PublishedAt        time.Time
	PageProperties     PageProperties `sql:"type:text"`
}

func (p Page) Validate(db *gorm.DB) {
	if strings.TrimSpace(p.Title) == "" {
		db.AddError(validations.NewError(p, "Name", "Name can not be empty"))
	}
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
