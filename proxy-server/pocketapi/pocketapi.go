package pocketapi

type DomainMetadata struct {
	Name          string `json:"name,omitempty"`
	Logo          string `json:"logo,omitempty"`
	GreyscaleLogo string `json:"greyscale_logo,omitempty"`
}

type GetRequest struct {
	AccessToken string `json:"access_token"`
	ConsumerKey string `json:"consumer_key"`
	// "article", "video", "image"
	ContentType string `json:"contentType"`
	// "simple" or "complete"
	DetailType string `json:"detailType"`
	// "unread" = only return unread items
	// "archive" = only return archived items
	// "all" = return both unread and archived items (default)
	State string `json:"state"`
	// "0" for only unfavorite, "1" for only favorite, "" for both.
	Favorite string `json:"favorite"`
	// "oldest", "newest", "title", "site"
	Sort string `json:"sort"`
	// How many items to retrieve. Max 30.
	Count *int `json:"count"`
	// Items to skip, used for pagination.
	Offset *int `json:"offset"`
	// Unix timestamp.
	Since *int `json:"since"`
}

type Image struct {
	ItemID  string `json:"item_id"`
	ImageID string `json:"image_id"`
	Src     string `json:"src"`
	Width   string `json:"width,omitempty"`
	Height  string `json:"height,omitempty"`
	Credit  string `json:"credit,omitempty"`
	Caption string `json:"caption,omitempty"`
}

type GetResponseItem struct {
	ItemID   string `json:"item_id"`
	Favorite string `json:"favorite"`
	// "1" if the item is archived, "2" if the item should be deleted
	Status        string `json:"status"`
	TimeAdded     string `json:"time_added"`
	TimeUpdated   string `json:"time_updated"`
	TimeRead      string `json:"time_read"`
	TimeFavorited string `json:"time_favorited"`
	SortID        int    `json:"sort_id"`
	Tags          struct {
	} `json:"tags"`
	TopImageURL   string `json:"top_image_url"`
	ResolvedID    string `json:"resolved_id"`
	GivenURL      string `json:"given_url"`
	GivenTitle    string `json:"given_title"`
	ResolvedTitle string `json:"resolved_title"`
	ResolvedURL   string `json:"resolved_url"`
	Excerpt       string `json:"excerpt"`
	IsArticle     string `json:"is_article"`
	IsIndex       string `json:"is_index"`
	HasVideo      string `json:"has_video"`
	HasImage      string `json:"has_image"`
	WordCount     string `json:"word_count"`
	Lang          string `json:"lang"`
	// In minutes.
	TimeToRead             int              `json:"time_to_read"`
	ListenDurationEstimate int              `json:"listen_duration_estimate"`
	DomainMetadata         *DomainMetadata  `json:"domain_metadata"`
	Images                 map[string]Image `json:"images"`
	Image                  *Image           `json:"image"`
}

type GetResponse struct {
	MaxActions int    `json:"maxActions"`
	Cachetype  string `json:"cachetype"`
	Status     int    `json:"status"`
	Error      any    `json:"error"`
	Complete   int    `json:"complete"`
	// Unix timestamp.
	Since int                        `json:"since"`
	List  map[string]GetResponseItem `json:"list"`
	Total int                        `json:"total"`
}

// Note: the article text request is form-encoded, not JSON.
type ArticleTextResponse struct {
	GivenURL            string `json:"given_url"`
	ItemID              string `json:"item_id"`
	NormalURL           string `json:"normal_url"`
	ResolvedNormalURL   string `json:"resolved_normal_url"`
	DateResolved        string `json:"date_resolved"`
	DomainID            string `json:"domain_id"`
	OriginDomainID      string `json:"origin_domain_id"`
	MimeType            string `json:"mime_type"`
	ContentLength       string `json:"content_length"`
	Encoding            string `json:"encoding"`
	TimeFirstParsed     string `json:"time_first_parsed"`
	HasOldDupes         string `json:"has_old_dupes"`
	InnerdomainRedirect string `json:"innerdomain_redirect"`
	TimeToRead          int    `json:"time_to_read"`
	HasImage            string `json:"has_image"`
	HasVideo            string `json:"has_video"`
	ResolvedID          string `json:"resolved_id"`
	ResolvedURL         string `json:"resolvedUrl"`
	Host                string `json:"host"`
	Title               string `json:"title"`
	DatePublished       string `json:"datePublished"`
	TimePublished       int    `json:"timePublished"`
	ResponseCode        string `json:"responseCode"`
	Excerpt             string `json:"excerpt"`
	Authors             struct {
	} `json:"authors"`
	Images         map[string]Image `json:"images"`
	Videos         string           `json:"videos"`
	WordCount      int              `json:"wordCount"`
	IsArticle      int              `json:"isArticle"`
	IsVideo        int              `json:"isVideo"`
	IsIndex        int              `json:"isIndex"`
	UsedFallback   int              `json:"usedFallback"`
	RequiresLogin  int              `json:"requiresLogin"`
	Lang           string           `json:"lang"`
	TopImageURL    string           `json:"topImageUrl"`
	DomainMetadata *DomainMetadata  `json:"domainMetadata"`
	// The HTML text of the article.
	// Note that images are not given as <img> tags, but instead by an HTML comment <!--IMG_{n}-->
	// where {n} is the kep of the image in the `images` map.
	Article string `json:"article"`
}
