package readeck

import (
	"fmt"
	"io"
	"net/http"
	"proxyserver/pocketapi"
	"strings"
	"time"
)

func copyFromGetItem(item getResponseItem, article *pocketapi.ArticleTextResponse) {
	article.ItemID = item.ID
	article.ResolvedID = item.ID
	article.GivenURL = item.URL
	article.NormalURL = item.URL
	article.ResolvedNormalURL = item.URL
	article.DateResolved = item.Created.Format(time.RFC3339)
	article.TimeToRead = &item.ReadingTime

	article.HasVideo = "0"
	article.Host = item.Site
	article.Title = item.Title
	article.DatePublished = item.Published.Format(time.RFC3339)
	article.ResponseCode = "200"
	article.Excerpt = item.Description

	pocketItem := item.toPocketItem()
	article.HasImage = pocketItem.HasImage
	article.Authors = pocketItem.Authors
	article.WordCount = &item.WordCount
	one := 1
	article.IsArticle = &one
	zero := 0
	article.IsIndex = &zero
	article.IsVideo = &zero
	article.Lang = item.Lang
}

func (conn *ReadeckConn) getArticleHTML(itemID string) (string, error) {
	deckReq, err := conn.createRequest(http.MethodGet, fmt.Sprintf("bookmarks/%s/article", itemID))
	if err != nil {
		return "", err
	}

	deckReq.Header.Set("Accept", "text/html")
	deckRes, err := http.DefaultClient.Do(deckReq)
	if err != nil {
		return "", err
	}
	if deckRes.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error calling Readeck API: [%d] %s", deckRes.StatusCode, deckRes.Status)
	}

	buf := new(strings.Builder)
	if _, err := io.Copy(buf, deckReq.Body); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func parseArticleText(articleText string, article *pocketapi.ArticleTextResponse) string {
	// We need to separate the <img> tags and replace them with HTML comments
	// of the form <!--IMG_n-->, since that what Pocket clients expect.

	// TODO: contentlength
}

func (conn *ReadeckConn) ArticleText(url string) (pocketapi.ArticleTextResponse, error) {
	id, cached := conn.urlIDCache[url]
	if !cached {
		return pocketapi.ArticleTextResponse{}, fmt.Errorf("unable to find URL %s in cache. Try calling /get first (and paginate all items) to refresh the cache", url)
	}

	item, err := conn.getOneItem(id)
	if err != nil {
		return pocketapi.ArticleTextResponse{}, err
	}
	article := pocketapi.ArticleTextResponse{}
	copyFromGetItem(item, &article)

	article.Encoding = "utf-8"
	articleText, err := conn.getArticleHTML(id)
	if err != nil {
		return pocketapi.ArticleTextResponse{}, fmt.Errorf("error getting article HTML: %v", err)
	}

	parseArticleText(articleText, &article)

	return article, nil
}
