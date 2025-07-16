// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	if err := checkResponseCode(deckRes); err != nil {
		return err
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
