package pocketapi

// Note: the article text request is form-encoded, not JSON.
type ArticleTextResponse struct {
	GivenURL            string            `json:"given_url"`
	ItemID              string            `json:"item_id"`
	NormalURL           string            `json:"normal_url"`
	ResolvedNormalURL   string            `json:"resolved_normal_url"`
	DateResolved        string            `json:"date_resolved"`
	DomainID            string            `json:"domain_id"`
	OriginDomainID      string            `json:"origin_domain_id"`
	MimeType            string            `json:"mime_type"`
	ContentLength       string            `json:"content_length"`
	Encoding            string            `json:"encoding"`
	TimeFirstParsed     string            `json:"time_first_parsed"`
	HasOldDupes         string            `json:"has_old_dupes"`
	InnerdomainRedirect string            `json:"innerdomain_redirect"`
	TimeToRead          int               `json:"time_to_read"`
	HasImage            string            `json:"has_image"`
	HasVideo            string            `json:"has_video"`
	ResolvedID          string            `json:"resolved_id"`
	ResolvedURL         string            `json:"resolvedUrl"`
	Host                string            `json:"host"`
	Title               string            `json:"title"`
	DatePublished       string            `json:"datePublished"`
	TimePublished       int               `json:"timePublished"`
	ResponseCode        string            `json:"responseCode"`
	Excerpt             string            `json:"excerpt"`
	Authors             map[string]Author `json:"authors"`
	Images              map[string]Image  `json:"images"`
	Videos              string            `json:"videos"`
	WordCount           int               `json:"wordCount"`
	IsArticle           int               `json:"isArticle"`
	IsVideo             int               `json:"isVideo"`
	IsIndex             int               `json:"isIndex"`
	UsedFallback        int               `json:"usedFallback"`
	RequiresLogin       int               `json:"requiresLogin"`
	Lang                string            `json:"lang"`
	TopImageURL         string            `json:"topImageUrl"`
	DomainMetadata      *DomainMetadata   `json:"domainMetadata"`
	// The HTML text of the article.
	// Note that images are not given as <img> tags, but instead by an HTML comment <!--IMG_{n}-->
	// where {n} is the kep of the image in the `images` map.
	Article string `json:"article"`
}
