package readeck

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"proxyserver/pocketapi"
	"strconv"
	"time"

	"golang.org/x/net/html"
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

func (conn *ReadeckConn) getArticleHTML(itemID string, received func(io.ReadCloser) error) error {
	deckReq, err := conn.createRequest(http.MethodGet, fmt.Sprintf("bookmarks/%s/article", itemID), nil)
	if err != nil {
		return err
	}

	deckReq.Header.Set("Accept", "text/html")
	deckRes, err := http.DefaultClient.Do(deckReq)
	if err != nil {
		return err
	}
	if deckRes.StatusCode != http.StatusOK {
		return fmt.Errorf("error calling Readeck API: [%d] %s", deckRes.StatusCode, deckRes.Status)
	}

	return received(deckReq.Body)
}

func parseArticleText(articleText io.ReadCloser, article *pocketapi.ArticleTextResponse) error {
	doc, err := html.Parse(articleText)
	if err != nil {
		return err
	}

	// We need to separate the <img> tags and replace them with HTML comments
	// of the form <!--IMG_n-->, since that what Pocket clients expect.
	article.Images = make(map[string]pocketapi.Image)

	var root *html.Node
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode {
			if n.Data == "body" {
				root = n
				// Replace the body with a root <div>, since the existing Pocket API
				// doesn't include a <body> tag.
				root.Data = "div"
			}

			if n.Data == "img" {
				pImg := pocketapi.Image{}
				for _, a := range n.Attr {
					if a.Key == "src" {
						pImg.Src = a.Val
					}
					if a.Key == "height" {
						pImg.Height = a.Val
					}
					if a.Key == "width" {
						pImg.Width = a.Val
					}
				}
				if pImg.Src == "" {
					// No image URL available, skip
					continue
				}
				// Save the URL
				pImg.ImageID = strconv.Itoa(len(article.Images) + 1)
				pImg.ItemID = article.ItemID
				article.Images[pImg.ImageID] = pImg

				// Replace the tag with a comment
				n.Type = html.CommentNode
				n.Data = fmt.Sprintf("IMG_%s", pImg.ImageID)
				n.Attr = nil
			}
		}
	}

	if root == nil {
		return errors.New("unable to parse HTML")
	}

	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, root)
	article.Article = buf.String()
	article.ContentLength = strconv.Itoa(len(article.Article))

	return nil
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
	err = conn.getArticleHTML(id, func(articleText io.ReadCloser) error {
		return parseArticleText(articleText, &article)
	})
	if err != nil {
		return pocketapi.ArticleTextResponse{}, fmt.Errorf("error getting article HTML: %v", err)
	}

	return article, nil
}
