package admin

import (
	"html/template"

	"github.com/qor/admin"
	// "github.com/k0kubun/pp"
)

func initFuncMap(Admin *admin.Admin) {
	Admin.RegisterFuncMap("render_latest_pages", renderLatestPages)
}

func renderLatestPages(context *admin.Context) template.HTML {
	var pageContext = context.NewResourceContext("Page")
	pageContext.Searcher.Pagination.PerPage = 25
	if pages, err := pageContext.FindMany(); err == nil {
		return pageContext.Render("index/table", pages)
	}
	return template.HTML("")
}
