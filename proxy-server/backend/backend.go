package backend

import (
	"proxyserver/pocketapi"
	"time"
)

type Backend interface {
	Get(req pocketapi.GetRequest) (pocketapi.GetResponse, error)
	ArticleText(url string) (pocketapi.ArticleTextResponse, error)
	Add(url string, title string, time time.Time) error
	Archive(itemID string, time time.Time) error
	Unarchive(itemID string, time time.Time) error
	Delete(itemID string, time time.Time) error
	Favorite(itemID string, time time.Time) error
	Unfavorite(itemID string, time time.Time) error
}
