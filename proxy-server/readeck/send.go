package readeck

import (
	"errors"
	"time"
)

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
