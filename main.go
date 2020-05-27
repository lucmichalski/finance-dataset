package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/k0kubun/pp"
	"github.com/oschwald/geoip2-golang"
	"github.com/qor/admin"
	"github.com/qor/assetfs"
	"github.com/qor/media"
	"github.com/qor/media/media_library"
	"github.com/qor/qor/utils"
	"github.com/qor/validations"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	realip "github.com/thanhhh/gin-gonic-realip"

	padmin "github.com/lucmichalski/finance-dataset/pkg/admin"
	"github.com/lucmichalski/finance-dataset/pkg/middlewares"
	"github.com/lucmichalski/finance-dataset/pkg/models"
	"github.com/lucmichalski/finance-dataset/pkg/plugins"
)

var (
	isHelp       bool
	isVerbose    bool
	isAdmin      bool
	isCrawl      bool
	isDataset    bool
	isTruncate   bool
	isClean      bool
	isCatalog    bool
	isDryMode    bool
	isNoCache    bool
	isTor        bool
	isExtract    bool
	parallelJobs int
	torAddress   string
	geoIpFile    string
	pluginDir    string
	cacheDir     string
	usePlugins   []string
	geo          *geoip2.Reader
	queueMaxSize = 100000000
	cachePath    = "./data/cache"
)

func main() {

	listPlugins, err := filepath.Glob("./release/*.so")
	if err != nil {
		panic(err)
	}
	var defaultPlugins []string
	for _, p := range listPlugins {
		p = strings.Replace(p, ".so", "", -1)
		p = strings.Replace(p, "finance-dataset-", "", -1)
		p = strings.Replace(p, "release/", "", -1)
		defaultPlugins = append(defaultPlugins, p)
	}

	pflag.BoolVarP(&isDryMode, "dry-mode", "", false, "do not insert data into database tables.")
	pflag.BoolVarP(&isCatalog, "catalog", "", false, "import datasets/catalogs.")
	pflag.StringVarP(&pluginDir, "plugin-dir", "", "./release", "plugins directory.")
	pflag.StringVarP(&geoIpFile, "geoip-db", "", "./shared/geoip2/GeoLite2-City.mmdb", "geoip filepath.")
	pflag.StringSliceVarP(&usePlugins, "plugins", "", defaultPlugins, "plugins to load.")
	pflag.IntVarP(&parallelJobs, "parallel-jobs", "j", 12, "parallel jobs.")
	pflag.BoolVarP(&isCrawl, "crawl", "c", false, "launch the crawler.")
	pflag.BoolVarP(&isDataset, "dataset", "d", false, "launch the crawler.")
	pflag.BoolVarP(&isClean, "clean", "", false, "auto-clean temporary files.")
	pflag.BoolVarP(&isAdmin, "admin", "", false, "launch the admin interface.")
	pflag.BoolVarP(&isTruncate, "truncate", "t", false, "truncate table content.")
	pflag.BoolVarP(&isExtract, "extract", "e", false, "extract data from urls.")
	pflag.BoolVarP(&isTor, "tor", "", false, "Proxy any GET requests with tor.")
	pflag.StringVarP(&torAddress, "tor-address", "", "sock5://localhost:5566", "Proxy addess with tor")
	pflag.StringVarP(&torAddress, "tor-privoxy", "", "http://localhost:8119", "Proxy address with tor-privoxy.")
	pflag.StringVarP(&cacheDir, "cache-dir", "", "./shared/data", "cache directory.")
	pflag.BoolVarP(&isNoCache, "no-cache", "", false, "disable crawler cache.")
	pflag.BoolVarP(&isVerbose, "verbose", "v", false, "verbose mode.")
	pflag.BoolVarP(&isHelp, "help", "h", false, "help info.")
	pflag.Parse()
	if isHelp {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// Instanciate geoip2 database
	geo = must(geoip2.Open(geoIpFile)).(*geoip2.Reader)

	// Instanciate the mysql client
	DB, err := gorm.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4,utf8&parseTime=True", os.Getenv("FD_MYSQL_USER"), os.Getenv("FD_MYSQL_PASSWORD"), os.Getenv("FD_MYSQL_HOST"), os.Getenv("FD_MYSQL_PORT"), os.Getenv("FD_MYSQL_DATABASE")))
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// callback for images and validation
	validations.RegisterCallbacks(DB)
	media.RegisterCallbacks(DB)

	// truncate table
	if isTruncate {
		if err := DB.DropTableIfExists(&models.Page{}).Error; err != nil {
			panic(err)
		}
	}

	// migrate tables
	DB.AutoMigrate(&models.Page{})
        DB.AutoMigrate(&models.PageAttribute{})
	DB.AutoMigrate(&media_library.MediaLibrary{})

	// load plugins
	ptPlugins := plugins.New()

	// The plugins (the *.so files) must be in a 'release' sub-directory
	allPlugins, err := filepath.Glob(pluginDir + "/*.so")
	if err != nil {
		panic(err)
	}

	var loadPlugins []string
	if len(usePlugins) > 0 {
		for _, p := range allPlugins {
			for _, u := range usePlugins {
				fmt.Println("usePlugin", u, "currentPlugin", p)
				if strings.HasPrefix(p, "release/finance-dataset-"+u+".so") {
					loadPlugins = append(loadPlugins, p)
				}
			}
		}
	} else {
		loadPlugins = allPlugins
	}

	// register commands from plugins
	for _, filename := range loadPlugins {
		p, err := plugin.Open(filename)
		if err != nil {
			panic(err)
		}
		// lookup for symbols
		cmdSymbol, err := p.Lookup(plugins.CmdSymbolName)
		if err != nil {
			fmt.Printf("plugin %s does not export symbol \"%s\"\n",
				filename, plugins.CmdSymbolName)
			continue
		}
		// check if symbol is implemented in Plugins interface
		commands, ok := cmdSymbol.(plugins.Plugins)
		if !ok {
			fmt.Printf("Symbol %s (from %s) does not implement Plugins interface\n",
				plugins.CmdSymbolName, filename)
			continue
		}
		// initialize plugin
		if err := commands.Init(ptPlugins.Ctx); err != nil {
			fmt.Printf("%s initialization failed: %v\n", filename, err)
			continue
		}
		// register commands from plugin
		for name, cmd := range commands.Registry() {
			ptPlugins.Commands[name] = cmd
		}
	}

	// migrate table from plugins
	for _, cmd := range ptPlugins.Commands {
		for _, table := range cmd.Migrate() {
			DB.AutoMigrate(table)
		}
	}

	if isExtract {
		fmt.Print("extracting...\n")
		for _, cmd := range ptPlugins.Commands {
			fmt.Printf(" from %s", cmd.Name())
			c := cmd.Config()
			if !isNoCache {
				c.CacheDir = cacheDir
			}
			c.IsDebug = true
			c.IsClean = isClean
			c.ConsumerThreads = parallelJobs
			pp.Println(c)
			c.DB = DB
			err := cmd.Crawl(c)
			if err != nil {
				log.Fatal(err)
			}
		}
		os.Exit(1)
	}

	if isAdmin {

		// Initialize AssetFS
		AssetFS := assetfs.AssetFS().NameSpace("admin")

		// Register custom paths to manually saved views
		AssetFS.RegisterPath(filepath.Join(utils.AppRoot, "./templates/qor/admin/views"))
		AssetFS.RegisterPath(filepath.Join(utils.AppRoot, "./templates/qor/media/views"))

		// Initialize Admin
		Admin := admin.New(&admin.AdminConfig{
			SiteName: "Finance Dataset",
			DB:       DB,
			AssetFS:  AssetFS,
		})

		padmin.SetupDashboard(DB, Admin)

		Admin.AddMenu(&admin.Menu{Name: "Crawl Management", Priority: 1})

		// Add media library
		Admin.AddResource(&media_library.MediaLibrary{}, &admin.Config{Menu: []string{"Crawl Management"}, Priority: -1})

		pages := Admin.AddResource(&models.Page{}, &admin.Config{Menu: []string{"Crawl Management"}})
		pages.IndexAttrs("ID", "Authors", "Link", "Title")

                pages.Meta(&admin.Meta{
                        Name: "Content",
                        Type: "rich_editor",
                })

		pages.Filter(&admin.Filter{
			Name: "Authors",
			Type: "string",
		})

		pages.Filter(&admin.Filter{
			Name: "PublishedAt",
		})

		pages.Filter(&admin.Filter{
			Name: "CreatedAt",
		})

		// initalize an HTTP request multiplexer
		mux := http.NewServeMux()

		// Mount admin interface to mux
		Admin.MountTo("/admin", mux)

		router := gin.Default()

		// router.Use(realip.RealIP())
		// globally use middlewares
		router.Use(
			realip.RealIP(),
			middlewares.RecoveryWithWriter(os.Stderr),
			middlewares.Logger(geo),
			middlewares.CORS(),
			gin.ErrorLogger(),
		)

		// add basic auth
		admin := router.Group("/admin", gin.BasicAuth(gin.Accounts{"finance": "moneytalk"}))
		{
			admin.Any("/*resources", gin.WrapH(mux))
		}

		router.Static("/system", "./public/system")
		router.Static("/public", "./public")

		fmt.Println("Listening on: 9009")
		s := &http.Server{
			Addr:           ":9009",
			Handler:        router,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		s.ListenAndServe()

	}

}

// fail fast on initialization
func must(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}

	return i
}
