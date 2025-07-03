package pocketapi

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
	Since *int64 `json:"since"`
}

type GetResponseItem struct {
	ItemID   string `json:"item_id"`
	Favorite string `json:"favorite"`
	// "1" if the item is archived, "2" if the item should be deleted
	Status        string `json:"status"`
	TimeAdded     string `json:"time_added"`
	TimeUpdated   string `json:"time_updated"`
	TimeFavorited string `json:"time_favorited"`
	Tags          struct {
	} `json:"tags"`
	TopImageURL   string `json:"top_image_url,omitempty"`
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
	TimeToRead             int               `json:"time_to_read"`
	ListenDurationEstimate int               `json:"listen_duration_estimate"`
	DomainMetadata         *DomainMetadata   `json:"domain_metadata"`
	Authors                map[string]Author `json:"authors"`
	Images                 map[string]Image  `json:"images"`
	Image                  *Image            `json:"image"`
}

type GetResponse struct {
	MaxActions int    `json:"maxActions"`
	Cachetype  string `json:"cachetype"`
	Status     int    `json:"status"`
	Error      any    `json:"error"`
	// Unix timestamp.
	Since int                        `json:"since"`
	List  map[string]GetResponseItem `json:"list"`
	Total int                        `json:"total"`
}
