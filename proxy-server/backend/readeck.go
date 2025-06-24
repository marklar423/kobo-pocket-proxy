package backend

import (
	"errors"
	"proxyserver/pocketapi"
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

func (conn *ReadeckConn) Get(req pocketapi.GetRequest) (pocketapi.GetResponse, error) {
	return pocketapi.GetResponse{}, errors.New("not implemented")
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
