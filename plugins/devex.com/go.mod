module github.com/lucmichalski/finance-contrib/devex.com

replace github.com/lucmichalski/finance-dataset => ../..

go 1.14

require (
	github.com/araddon/dateparse v0.0.0-20200409225146-d820a6159ab1
	github.com/beevik/etree v1.1.0 // indirect
	github.com/corpix/uarand v0.1.1
	github.com/gocolly/colly/v2 v2.0.1
	github.com/jinzhu/gorm v1.9.12
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/lucmichalski/finance-dataset v0.0.0-00010101000000-000000000000
	github.com/qor/admin v0.0.0-20200315024928-877b98a68a6f
	github.com/sirupsen/logrus v1.6.0
	github.com/tsak/concurrent-csv-writer v0.0.0-20200206204244-84054e222625 // indirect
)
