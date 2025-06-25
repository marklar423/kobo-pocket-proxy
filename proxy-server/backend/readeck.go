package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"proxyserver/pocketapi"
	"strconv"
	"strings"
	"time"
)

type ReadeckConn struct {
	endpoint    string
	bearerToken string
}

func NewReadeckConn(endpoint string, bearerToken string) *ReadeckConn {
	return &ReadeckConn{
		endpoint:    endpoint,
		bearerToken: bearerToken,
	}
}

func (conn *ReadeckConn) createRequest(method, action string) (*http.Request, error) {
	apiUrl := fmt.Sprintf("%s/api/%s", conn.endpoint, action)
	deckReq, err := http.NewRequest(method, apiUrl, nil)
	if err != nil {
		return nil, err
	}
	deckReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", conn.bearerToken))
	return deckReq, nil
}

func buildGetQuerystring(req pocketapi.GetRequest) string {
	var query url.Values

	if req.Count != nil {
		query.Set("limit", strconv.Itoa(max(0, *req.Count)))
	}
	if req.Offset != nil {
		query.Set("offset", strconv.Itoa(max(0, *req.Offset)))
	}
	if req.Since != nil {
		query.Set("updated_since", time.Unix(*req.Since, 0).Format(time.RFC3339))
	}

	switch strings.ToLower(req.State) {
	case "unread":
		query.Set("is_archived", "0")
	case "archive":
		query.Set("is_archived", "1")
	case "all":
		fallthrough
	default:
		// Leave it unset.
	}

	switch strings.ToLower(req.Favorite) {
	case "0":
		query.Set("is_marked", "0")
	case "1":
		query.Set("is_marked", "1")
	default:
		// Leave it unset.
	}

	switch strings.ToLower(req.Sort) {
	case "oldest":
		query.Set("sort", "created")
	case "title":
		query.Set("sort", "title")
	case "site":
		query.Set("sort", "domain")
	case "newest":
		fallthrough
	default:
		query.Set("sort", "-created")
	}

	switch strings.ToLower(req.ContentType) {
	case "video":
		query.Set("type", "video")
	case "image":
		query.Set("type", "photo")
	case "article":
		fallthrough
	default:
		query.Set("type", "article")
	}

	return query.Encode()
}

type getResponseItem struct {
	ID            string    `json:"id"`
	Href          string    `json:"href"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
	State         int       `json:"state"`
	Loaded        bool      `json:"loaded"`
	URL           string    `json:"url"`
	Title         string    `json:"title"`
	SiteName      string    `json:"site_name"`
	Site          string    `json:"site"`
	Authors       []string  `json:"authors"`
	Lang          string    `json:"lang"`
	TextDirection string    `json:"text_direction"`
	DocumentType  string    `json:"document_type"`
	Type          string    `json:"type"`
	HasArticle    bool      `json:"has_article"`
	Description   string    `json:"description"`
	IsDeleted     bool      `json:"is_deleted"`
	IsMarked      bool      `json:"is_marked"`
	IsArchived    bool      `json:"is_archived"`
	Labels        []any     `json:"labels"`
	ReadProgress  int       `json:"read_progress"`
	Resources     resources `json:"resources,omitempty"`
	WordCount     int       `json:"word_count,omitempty"`
	ReadingTime   int       `json:"reading_time,omitempty"`
	Published     time.Time `json:"published,omitempty"`
}

type resource struct {
	Src    string `json:"src"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type resources struct {
	Log       resource  `json:"log"`
	Props     resource  `json:"props"`
	Article   *resource `json:"article"`
	Icon      *resource `json:"icon"`
	Image     *resource `json:"image"`
	Thumbnail *resource `json:"thumbnail"`
}

func (m getResponseItem) toPocketItem() pocketapi.GetResponseItem {
	oneIfTrue := func(val bool) string {
		if val {
			return "1"
		}
		return "0"
	}

	timeFavorited := "0"
	if m.IsMarked {
		timeFavorited = strconv.FormatInt(m.Updated.Unix(), 10)
	}

	var domainMeta *pocketapi.DomainMetadata
	if m.SiteName != "" {
		domainMeta = &pocketapi.DomainMetadata{Name: m.SiteName}
	}

	authors := make(map[string]pocketapi.Author)
	for _, a := range m.Authors {
		id := a // TODO: hash
		authors[id] = pocketapi.Author{
			AuthorID: id,
			Name:     a,
			ItemID:   m.ID,
		}
	}

	hasImage := "0"
	var topImageUrl string
	var image *pocketapi.Image
	if m.Resources.Image != nil {
		hasImage = "1"
		topImageUrl = m.Resources.Image.Src
		id := m.Resources.Image.Src // TODO: hash
		image = &pocketapi.Image{
			ItemID:  m.ID,
			ImageID: id,
			Src:     m.Resources.Image.Src,
			Width:   strconv.Itoa(m.Resources.Image.Width),
			Height:  strconv.Itoa(m.Resources.Image.Height),
		}
	}

	return pocketapi.GetResponseItem{
		ItemID:                 m.ID,
		Favorite:               oneIfTrue(m.IsMarked),
		Status:                 "1",
		TimeAdded:              strconv.FormatInt(m.Created.Unix(), 10),
		TimeUpdated:            strconv.FormatInt(m.Updated.Unix(), 10),
		TimeRead:               strconv.Itoa(m.ReadingTime),
		TimeFavorited:          timeFavorited,
		SortID:                 0,
		ResolvedID:             m.ID,
		GivenURL:               m.URL,
		GivenTitle:             m.Title,
		ResolvedTitle:          m.Title,
		ResolvedURL:            m.URL,
		Excerpt:                m.Description,
		IsArticle:              oneIfTrue(m.Type == "article"),
		IsIndex:                "0",
		HasVideo:               "0",
		WordCount:              strconv.Itoa(m.WordCount),
		Lang:                   m.Lang,
		TimeToRead:             m.ReadingTime,
		ListenDurationEstimate: 0,
		DomainMetadata:         domainMeta,
		Authors:                authors,
		HasImage:               hasImage,
		Image:                  image,
		TopImageURL:            topImageUrl,
		// Omit Images since Readeck doesn't give us all the images here.
	}
}

func translateGetResponse(deckRes *http.Response) (pocketapi.GetResponse, error) {
	var deckItems []getResponseItem
	if err := json.NewDecoder(deckRes.Body).Decode(&deckItems); err != nil {
		return pocketapi.GetResponse{}, err
	}

	var pocketRes pocketapi.GetResponse
	pocketRes.Status = 1

	if resTotal := deckRes.Header.Get("Total-Count"); resTotal != "" {
		if t, err := strconv.Atoi(resTotal); err == nil {
			pocketRes.Total = t
		}
	}
	for _, item := range deckItems {
		pocketRes.List[item.ID] = item.toPocketItem()
	}

	return pocketRes, nil
}

func (conn *ReadeckConn) Get(req pocketapi.GetRequest) (pocketapi.GetResponse, error) {
	deckReq, err := conn.createRequest(http.MethodGet, "bookmarks")
	if err != nil {
		return pocketapi.GetResponse{}, err
	}
	deckReq.URL.RawQuery = buildGetQuerystring(req)

	deckRes, err := http.DefaultClient.Do(deckReq)
	if err != nil {
		return pocketapi.GetResponse{}, err
	}

	return translateGetResponse(deckRes)
}

func (conn *ReadeckConn) ArticleText(url string) (pocketapi.ArticleTextResponse, error) {
	return pocketapi.ArticleTextResponse{}, errors.New("not implemented")
}

func (conn *ReadeckConn) Add(url string, title string, time time.Time) error {
	return errors.New("not implemented")
}

func (conn *ReadeckConn) Archive(itemID string, time time.Time) error {
	return errors.New("not implemented")
}

func (conn *ReadeckConn) Unarchive(itemID string, time time.Time) error {
	return errors.New("not implemented")
}

func (conn *ReadeckConn) Delete(itemID string, time time.Time) error {
	return errors.New("not implemented")
}

func (conn *ReadeckConn) Favorite(itemID string, time time.Time) error {
	return errors.New("not implemented")
}

func (conn *ReadeckConn) Unfavorite(itemID string, time time.Time) error {
	return errors.New("not implemented")
}
