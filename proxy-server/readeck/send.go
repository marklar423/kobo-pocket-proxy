package readeck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var pointerTrue bool = true
var pointerFalse bool = false

type updateRequest struct {
	IsDeleted  *bool `json:"is_deleted,omitempty"`
	IsMarked   *bool `json:"is_marked,omitempty"`
	IsArchived *bool `json:"is_archived,omitempty"`
}

func sendUpdate(conn *ReadeckConn, itemID string, params updateRequest) error {
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(params); err != nil {
		return err
	}

	deckReq, err := conn.createRequest(http.MethodPatch, fmt.Sprintf("bookmarks/%s", itemID), &buffer)
	if err != nil {
		return err
	}
	deckReq.Header.Set("Content-Type", "application/json")

	deckRes, err := http.DefaultClient.Do(deckReq)
	if err != nil {
		return err
	}
	if deckRes.StatusCode != http.StatusOK {
		return fmt.Errorf("error calling Readeck API: [%d] %s", deckRes.StatusCode, deckRes.Status)
	}
	return nil
}

type insertRequest struct {
	Url   string `json:"url"`
	Title string `json:"title,omitempty"`
}

func (conn *ReadeckConn) Add(url string, title string, time time.Time) error {
	body := insertRequest{Url: url, Title: title}
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(body); err != nil {
		return err
	}

	deckReq, err := conn.createRequest(http.MethodPost, "bookmarks", &buffer)
	if err != nil {
		return err
	}
	deckReq.Header.Set("Content-Type", "application/json")

	deckRes, err := http.DefaultClient.Do(deckReq)
	if err != nil {
		return err
	}
	if deckRes.StatusCode != http.StatusOK {
		return fmt.Errorf("error calling Readeck API: [%d] %s", deckRes.StatusCode, deckRes.Status)
	}

	// Cache the returned ID
	itemID := deckRes.Header.Get("Bookmark-Id")
	conn.urlIDCache[url] = itemID

	return nil
}

func (conn *ReadeckConn) Archive(itemID string, time time.Time) error {
	return sendUpdate(conn, itemID, updateRequest{IsArchived: &pointerTrue})
}

func (conn *ReadeckConn) Unarchive(itemID string, time time.Time) error {
	return sendUpdate(conn, itemID, updateRequest{IsArchived: &pointerFalse})
}

func (conn *ReadeckConn) Delete(itemID string, time time.Time) error {
	return sendUpdate(conn, itemID, updateRequest{IsDeleted: &pointerTrue})
}

func (conn *ReadeckConn) Favorite(itemID string, time time.Time) error {
	return sendUpdate(conn, itemID, updateRequest{IsMarked: &pointerTrue})
}

func (conn *ReadeckConn) Unfavorite(itemID string, time time.Time) error {
	return sendUpdate(conn, itemID, updateRequest{IsMarked: &pointerFalse})
}
