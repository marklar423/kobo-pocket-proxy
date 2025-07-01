package backend

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"proxyserver/pocketapi"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func numPointer[T int | int64](value T) *T {
	return &value
}

func TestReadeck_GetRequest(t *testing.T) {
	testCases := []struct {
		name    string
		request pocketapi.GetRequest
		want    url.Values
	}{
		{
			name:    "Defaults",
			request: pocketapi.GetRequest{},
			want: url.Values{
				"sort": []string{"-created"},
				"type": []string{"article"},
			},
		},
		{
			name: "Limit & Offset",
			request: pocketapi.GetRequest{
				Count:  numPointer(30),
				Offset: numPointer(10),
			},
			want: url.Values{
				"sort":   []string{"-created"},
				"type":   []string{"article"},
				"limit":  []string{"30"},
				"offset": []string{"10"},
			},
		},
		{
			name: "Since",
			request: pocketapi.GetRequest{
				Since: numPointer(int64(0)),
			},
			want: url.Values{
				"sort":          []string{"-created"},
				"type":          []string{"article"},
				"updated_since": []string{time.Unix(0, 0).Format(time.RFC3339)},
			},
		},
		{
			name: "Favorite",
			request: pocketapi.GetRequest{
				Favorite: "1",
			},
			want: url.Values{
				"sort":      []string{"-created"},
				"type":      []string{"article"},
				"is_marked": []string{"1"},
			},
		},
		{
			name: "Not Favorite",
			request: pocketapi.GetRequest{
				Favorite: "0",
			},
			want: url.Values{
				"sort":      []string{"-created"},
				"type":      []string{"article"},
				"is_marked": []string{"0"},
			},
		},
		{
			name: "Archived",
			request: pocketapi.GetRequest{
				State: "archive",
			},
			want: url.Values{
				"sort":        []string{"-created"},
				"type":        []string{"article"},
				"is_archived": []string{"1"},
			},
		},
		{
			name: "Not Archived",
			request: pocketapi.GetRequest{
				State: "unread",
			},
			want: url.Values{
				"sort":        []string{"-created"},
				"type":        []string{"article"},
				"is_archived": []string{"0"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				if r.Method != http.MethodGet {
					t.Errorf("Unexpected HTTP method, want GET got %s", r.Method)
				}

				wantToken := "Bearer token"
				if r.Header.Get("Authorization") != wantToken {
					t.Errorf("Unexepcted authorization header: want %s got %s", wantToken, r.Header.Get("Authorization"))
				}

				if diff := cmp.Diff(tc.want, r.URL.Query()); diff != "" {
					t.Errorf("GET query mismatch (-want +got):\n%s", diff)
				}
			}))
			defer server.Close()

			readeck := NewReadeckConn(server.URL, "token")
			readeck.Get(tc.request)
		})
	}
}

func TestReadeck_GetResponse(t *testing.T) {
	testCases := []struct {
		name         string
		jsonResponse string
		responseCode int
		want         pocketapi.GetResponse
		wantError    bool
	}{
		{
			name:         "Empty",
			responseCode: http.StatusOK,
			jsonResponse: "[]",
			want: pocketapi.GetResponse{
				Status: 1,
			},
		},
		{
			name:         "Error",
			responseCode: http.StatusNotFound,
			wantError:    true,
		},
		{
			name:         "All Fields One Item",
			responseCode: http.StatusOK,
			jsonResponse: `
			[{
			  "id": "id123",
				"is_marked": false,
				"created": "2025-07-01T22:17:08+00:00",
				"updated": "2025-07-01T22:17:08+00:00",
				"reading_time": 10,
				"href": "http://example.com"
		  }]
			`,
			want: pocketapi.GetResponse{
				Status:   1,
				Complete: 1,
				List: map[string]pocketapi.GetResponseItem{
					"id123": pocketapi.GetResponseItem{},
				},
				Total: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.responseCode)
				w.Write([]byte(tc.jsonResponse))
			}))
			defer server.Close()

			readeck := NewReadeckConn(server.URL, "token")
			res, err := readeck.Get(pocketapi.GetRequest{})

			if tc.wantError && err == nil {
				t.Error("Wanted error, got nil instead")
			}

			if !tc.wantError && err != nil {
				t.Errorf("Wanted nil error, got %v instead", err)
			}

			if diff := cmp.Diff(tc.want, res); diff != "" {
				t.Errorf("GET response mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
