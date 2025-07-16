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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func boolPointer(value bool) *bool {
	return &value
}

func TestReadeck_SendUpdate(t *testing.T) {
	const itemID = "id123"

	testCases := []struct {
		name       string
		update     func(*ReadeckConn) error
		statusCode int
		wantError  bool
		wantBody   updateRequest
	}{
		{
			name:       "Archive",
			statusCode: http.StatusOK,
			update: func(conn *ReadeckConn) error {
				return conn.Archive(itemID, time.Time{})
			},
			wantBody: updateRequest{IsDeleted: nil, IsMarked: nil, IsArchived: boolPointer(true)},
		},
		{
			name:       "Unarchive",
			statusCode: http.StatusOK,
			update: func(conn *ReadeckConn) error {
				return conn.Unarchive(itemID, time.Time{})
			},
			wantBody: updateRequest{IsDeleted: nil, IsMarked: nil, IsArchived: boolPointer(false)},
		},
		{
			name:       "Favorite",
			statusCode: http.StatusOK,
			update: func(conn *ReadeckConn) error {
				return conn.Favorite(itemID, time.Time{})
			},
			wantBody: updateRequest{IsDeleted: nil, IsMarked: boolPointer(true), IsArchived: nil},
		},
		{
			name:       "Unfavorite",
			statusCode: http.StatusOK,
			update: func(conn *ReadeckConn) error {
				return conn.Unfavorite(itemID, time.Time{})
			},
			wantBody: updateRequest{IsDeleted: nil, IsMarked: boolPointer(false), IsArchived: nil},
		},
		{
			name:       "Delete",
			statusCode: http.StatusOK,
			update: func(conn *ReadeckConn) error {
				return conn.Delete(itemID, time.Time{})
			},
			wantBody: updateRequest{IsDeleted: boolPointer(true), IsMarked: nil, IsArchived: nil},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)

				if r.Method != http.MethodPatch {
					t.Errorf("Unexpected HTTP method, want PATCH got %s", r.Method)
				}

				wantToken := "Bearer token123"
				if r.Header.Get("Authorization") != wantToken {
					t.Errorf("Unexepcted authorization header: want %s got %s", wantToken, r.Header.Get("Authorization"))
				}

				wantUrl := fmt.Sprintf("/api/bookmarks/%s", itemID)
				if r.URL.Path != wantUrl {
					t.Errorf("Unexpected URL path, got %s want %s", r.URL.Path, wantUrl)
				}

				var gotBody updateRequest
				if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
					t.Errorf("Unexpected error parsing body: %v", err)
				}

				if diff := cmp.Diff(tc.wantBody, gotBody); diff != "" {
					t.Errorf("Body JSON mismatch (-want +got):\n%s", diff)
				}
			}))
			defer server.Close()

			readeck := NewReadeckConn(server.URL, "token123")
			err := tc.update(readeck)

			if tc.wantError && err == nil {
				t.Error("Unexpected response from action: want error got nil")
			} else if !tc.wantError && err != nil {
				t.Errorf("Unexpected response from action: want nil got error %v", err)
			}
		})
	}
}

func TestReadeck_Add(t *testing.T) {
	wantBody := insertRequest{Url: "http://example.com/path-to-file?key=value"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPost {
			t.Errorf("Unexpected HTTP method, want POST got %s", r.Method)
		}

		wantToken := "Bearer token123"
		if r.Header.Get("Authorization") != wantToken {
			t.Errorf("Unexepcted authorization header: want %s got %s", wantToken, r.Header.Get("Authorization"))
		}

		wantUrl := "/api/bookmarks"
		if r.URL.Path != wantUrl {
			t.Errorf("Unexpected URL path, got %s want %s", r.URL.Path, wantUrl)
		}

		var gotBody insertRequest
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("Unexpected error parsing body: %v", err)
		}

		if diff := cmp.Diff(wantBody, gotBody); diff != "" {
			t.Errorf("Body JSON mismatch (-want +got):\n%s", diff)
		}
	}))
	defer server.Close()

	readeck := NewReadeckConn(server.URL, "token123")
	if err := readeck.Add("http://example.com/path-to-file?key=value", "", time.Time{}); err != nil {
		t.Errorf("Unexpected error from Add(): want nil got %v", err)
	}

}
