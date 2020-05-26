package models

type Article struct {
	Breadcrumbs []ArticleBreadcrumb `json:"breadcrumbs"`
	Meta        ArticleMeta         `json:"meta"`
	Path        string              `json:"path"`
	Section     ArticleSection      `json:"section"`
	Stripes     []ArticleStripe     `json:"stripes"`
	Subsection  ArticleSubsection   `json:"subsection"`
}

type ArticleBreadcrumb struct {
	Label string `json:"label"`
	Path  string `json:"path"`
	Slug  string `json:"slug"`
}

type ArticleMeta struct {
	Description interface{} `json:"description"`
	Robots      []string    `json:"robots"`
	Title       interface{} `json:"title"`
}

type ArticleSection struct {
	Label string `json:"label"`
	Path  string `json:"path"`
	Slug  string `json:"slug"`
}

type ArticleStripe struct {
	AllowedPostTypes   []string                   `json:"allowedPostTypes"`
	ForbiddenPostTypes []string                   `json:"forbiddenPostTypes"`
	ID                 string                     `json:"id"`
	MainContent        []ArticleStripeMainContent `json:"mainContent"`
	OnlyNotAmp         interface{}                `json:"onlyNotAmp"`
	OnlyProspect       bool                       `json:"onlyProspect"`
	OnlySubscribed     bool                       `json:"onlySubscribed"`
	Sidebar            []interface{}              `json:"sidebar"`
	Stripe             string                     `json:"stripe"`
	Template           string                     `json:"template"`
}

type ArticleStripeMainContent struct {
	Data       ArticleStripeMainContentData   `json:"data"`
	ID         string                         `json:"id"`
	Items      []ArticleStripeMainContentItem `json:"items"`
	Nb         int                            `json:"nb"`
	ReturnType string                         `json:"returnType"`
	Type       string                         `json:"type"`
}

type ArticleStripeMainContentData struct {
	Access          string                                 `json:"access"`
	Authors         []ArticleStripeMainContentDataAuthor   `json:"authors"`
	Description     string                                 `json:"description"`
	ID              float64                                `json:"id"`
	Image           ArticleStripeMainContentDataImage      `json:"image"`
	Lead            string                                 `json:"lead"`
	Path            string                                 `json:"path"`
	PublicationDate string                                 `json:"publicationDate"`
	ReadingDuration int                                    `json:"readingDuration"`
	Section         ArticleStripeMainContentDataSection    `json:"section"`
	Signature       string                                 `json:"signature"`
	Slug            string                                 `json:"slug"`
	Subsection      ArticleStripeMainContentDataSubsection `json:"subsection"`
	Subtitle        interface{}                            `json:"subtitle"`
	Tags            ArticleStripeMainContentDataTags       `json:"tags"`
	Title           string                                 `json:"title"`
	Type            string                                 `json:"type"`
	UpdateDate      string                                 `json:"updateDate"`
}

type ArticleStripeMainContentDataAuthor struct {
	FirstName      string                                            `json:"firstName"`
	ID             string                                            `json:"id"`
	Job            string                                            `json:"job"`
	LastName       string                                            `json:"lastName"`
	Signature      string                                            `json:"signature"`
	Slug           string                                            `json:"slug"`
	SubscribeLinks []ArticleStripeMainContentDataAuthorSubscribeLink `json:"subscribeLinks"`
	Type           string                                            `json:"type"`
}

type ArticleStripeMainContentDataAuthorSubscribeLink struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

type ArticleStripeMainContentDataImage struct {
	Caption  string `json:"caption"`
	Credits  string `json:"credits"`
	Filename string `json:"filename"`
	ID       string `json:"id"`
	Title    string `json:"title"`
}

type ArticleStripeMainContentDataSection struct {
	Label string `json:"label"`
	Path  string `json:"path"`
	Slug  string `json:"slug"`
}

type ArticleStripeMainContentDataSubsection struct {
	Label string `json:"label"`
	Path  string `json:"path"`
	Slug  string `json:"slug"`
}

type ArticleStripeMainContentDataTags struct {
	Categorization []ArticleStripeMainContentDataTagsCategorization `json:"categorization"`
	Geography      []interface{}                                    `json:"geography"`
	Location       []string                                         `json:"location"`
	Organizations  []string                                         `json:"organizations"`
	People         []string                                         `json:"people"`
}

type ArticleStripeMainContentDataTagsCategorization struct {
	Medias string `json:"medias"`
	Names  string `json:"names"`
}

type ArticleStripeMainContentItem struct {
	Access           string                                 `json:"access"`
	Authors          []ArticleStripeMainContentItemAuthor   `json:"authors"`
	Category         interface{}                            `json:"category"`
	ID               float64                                `json:"id"`
	Image            ArticleStripeMainContentItemImage      `json:"image"`
	Label            interface{}                            `json:"label"`
	Lead             string                                 `json:"lead"`
	Path             string                                 `json:"path"`
	PublicationDate  string                                 `json:"publicationDate"`
	Section          ArticleStripeMainContentItemSection    `json:"section"`
	ShortDescription string                                 `json:"shortDescription"`
	Subsection       ArticleStripeMainContentItemSubsection `json:"subsection"`
	Synthesis        interface{}                            `json:"synthesis"`
	Title            string                                 `json:"title"`
	Type             string                                 `json:"type"`
	UpdateDate       string                                 `json:"updateDate"`
}

type ArticleStripeMainContentItemAuthor struct {
	FirstName      string                                            `json:"firstName"`
	ID             string                                            `json:"id"`
	Job            string                                            `json:"job"`
	LastName       string                                            `json:"lastName"`
	Signature      string                                            `json:"signature"`
	Slug           string                                            `json:"slug"`
	SubscribeLinks []ArticleStripeMainContentItemAuthorSubscribeLink `json:"subscribeLinks"`
	Type           string                                            `json:"type"`
}

type ArticleStripeMainContentItemAuthorSubscribeLink struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

type ArticleStripeMainContentItemImage struct {
	Caption  string `json:"caption"`
	Credits  string `json:"credits"`
	Filename string `json:"filename"`
	ID       string `json:"id"`
	Title    string `json:"title"`
}

type ArticleStripeMainContentItemSection struct {
	Label string `json:"label"`
	Path  string `json:"path"`
	Slug  string `json:"slug"`
}

type ArticleStripeMainContentItemSubsection struct {
	Label string `json:"label"`
	Path  string `json:"path"`
	Slug  string `json:"slug"`
}

type ArticleSubsection struct {
	Label string `json:"label"`
	Path  string `json:"path"`
	Slug  string `json:"slug"`
}
