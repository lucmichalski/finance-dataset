package config

import (
	"database/sql"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/jinzhu/gorm"
)

type Config struct {
	IsDebug         bool
	IsSitemapIndex  bool
	AllowedDomains  []string
	CacheDir        string
	ConsumerThreads int
	QueueMaxSize    int
	URLs            []string
	DryMode         bool
	IsClean         bool
	ProxyAddress    string
	CatalogURL      string
	Index           string
	DumpDir         string     `default:"./shared/dump"`
	DB              *gorm.DB   `gorm:"-"`
	IDX             *sql.DB    `gorm:"-"`
	KV              *badger.DB `gorm:"-"`
	Writer          *os.File   `gorm:"-"`
}
