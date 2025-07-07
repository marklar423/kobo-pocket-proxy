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
