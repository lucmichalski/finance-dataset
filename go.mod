module github.com/lucmichalski/finance-dataset

go 1.14

require (
	cloud.google.com/go v0.41.0
	github.com/BurntSushi/toml v0.3.1
	github.com/BurntSushi/xgbutil v0.0.0-20160919175755-f7c97cef3b4e
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/abadojack/whatlanggo v1.0.1
	github.com/antchfx/htmlquery v1.0.0
	github.com/antchfx/xmlquery v1.0.0
	github.com/aodin/date v0.0.0-20160219192542-c5f6146fc644
	github.com/araddon/dateparse v0.0.0-20200409225146-d820a6159ab1
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535 // indirect
	github.com/aws/aws-sdk-go v1.32.1 // indirect
	github.com/beevik/etree v1.1.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/corpix/uarand v0.1.1
	github.com/dghubble/oauth1 v0.6.0
	github.com/dgraph-io/badger v1.6.1
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200531234253-77e06fda0c94+incompatible
	github.com/fatih/color v1.9.0
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/gin-gonic/contrib v0.0.0-20191209060500-d6e26eeaa607
	github.com/gin-gonic/gin v1.6.3
	github.com/go-shiori/go-readability v0.0.0-20200413080041-05caea5f6592
	github.com/gobwas/glob v0.2.3
	github.com/gocolly/colly/v2 v2.0.1
	github.com/gocolly/twocaptcha v0.0.0-20180302192906-5ade8d29ba53
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.4.0
	github.com/google/go-github/v27 v27.0.4
	github.com/google/go-querystring v1.0.0
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/sessions v1.2.0 // indirect
	github.com/gosimple/slug v1.9.0
	github.com/h2non/filetype v1.1.0
	github.com/jawher/mow.cli v1.1.0
	github.com/jaytaylor/html2text v0.0.0-20200412013138-3577fbdbcff7
	github.com/jinzhu/configor v1.2.0 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/jinzhu/now v1.1.1
	github.com/joho/godotenv v1.3.0
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/karrick/godirwalk v1.15.6
	github.com/kennygrant/sanitize v1.2.4
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/microcosm-cc/bluemonday v1.0.2 // indirect
	github.com/neurosnap/sentences v1.0.6 // indirect
	github.com/nozzle/throttler v0.0.0-20180817012639-2ea982251481
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/oschwald/geoip2-golang v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/qor/admin v0.0.0-20200315024928-877b98a68a6f
	github.com/qor/assetfs v0.0.0-20170713023933-ff57fdc13a14
	github.com/qor/media v0.0.0-20191022071353-19cf289e17d4
	github.com/qor/middlewares v0.0.0-20170822143614-781378b69454 // indirect
	github.com/qor/oss v0.0.0-20191031055114-aef9ba66bf76 // indirect
	github.com/qor/qor v0.0.0-20200224122013-457d2e3f50e1
	github.com/qor/responder v0.0.0-20171031032654-b6def473574f // indirect
	github.com/qor/roles v0.0.0-20171127035124-d6375609fe3e // indirect
	github.com/qor/serializable_meta v0.0.0-20180510060738-5fd8542db417 // indirect
	github.com/qor/session v0.0.0-20170907035918-8206b0adab70 // indirect
	github.com/qor/validations v0.0.0-20171228122639-f364bca61b46
	github.com/qorpress/go-wordpress v0.0.0-20200302054333-81c736c8fa04
	github.com/robbiet480/go-wordpress v0.0.0-20180206201500-3b8369ffcef3
	github.com/saintfish/chardet v0.0.0-20120816061221-3af4cd4741ca
	github.com/schollz/pluck v1.1.3
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/pflag v1.0.5
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stretchr/testify v1.5.1
	github.com/tebeka/selenium v0.9.9
	github.com/tejasmanohar/timerange-go v1.0.0
	github.com/temoto/robotstxt v1.1.1
	github.com/thanhhh/gin-gonic-realip v0.0.0-20180527053022-1a91c06e8abf
	github.com/theplant/cldr v0.0.0-20190423050709-9f76f7ce4ee8 // indirect
	github.com/theplant/htmltestingutils v0.0.0-20190423050759-0e06de7b6967 // indirect
	github.com/theplant/testingutils v0.0.0-20190603093022-26d8b4d95c61 // indirect
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	github.com/tsak/concurrent-csv-writer v0.0.0-20200206204244-84054e222625
	github.com/yosssi/gohtml v0.0.0-20200519115854-476f5b4b8047 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.7.0
	google.golang.org/appengine v1.6.1
	gopkg.in/neurosnap/sentences.v1 v1.0.6
)
