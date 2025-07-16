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
	"net/http"
	"net/http/httptest"
	"net/url"
	"proxyserver/pocketapi"
	"strconv"
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
				"sort": {"-created"},
				"type": {"article"},
			},
		},
		{
			name: "Limit & Offset",
			request: pocketapi.GetRequest{
				Count:  numPointer(30),
				Offset: numPointer(10),
			},
			want: url.Values{
				"sort":   {"-created"},
				"type":   {"article"},
				"limit":  {"30"},
				"offset": {"10"},
			},
		},
		{
			name: "Since",
			request: pocketapi.GetRequest{
				Since: numPointer(int64(0)),
			},
			want: url.Values{
				"sort":          {"-created"},
				"type":          {"article"},
				"updated_since": {time.Unix(0, 0).Format(time.RFC3339)},
			},
		},
		{
			name: "Favorite",
			request: pocketapi.GetRequest{
				Favorite: "1",
			},
			want: url.Values{
				"sort":      {"-created"},
				"type":      {"article"},
				"is_marked": {"1"},
			},
		},
		{
			name: "Not Favorite",
			request: pocketapi.GetRequest{
				Favorite: "0",
			},
			want: url.Values{
				"sort":      {"-created"},
				"type":      {"article"},
				"is_marked": {"0"},
			},
		},
		{
			name: "Archived",
			request: pocketapi.GetRequest{
				State: "archive",
			},
			want: url.Values{
				"sort":        {"-created"},
				"type":        {"article"},
				"is_archived": {"1"},
			},
		},
		{
			name: "Not Archived",
			request: pocketapi.GetRequest{
				State: "unread",
			},
			want: url.Values{
				"sort":        {"-created"},
				"type":        {"article"},
				"is_archived": {"0"},
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
		totalItems   int
		want         pocketapi.GetResponse
		wantError    bool
	}{
		{
			name:         "Empty",
			responseCode: http.StatusOK,
			jsonResponse: "[]",
			want: pocketapi.GetResponse{
				Status: 1,
				List:   map[string]pocketapi.GetResponseItem{},
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
			totalItems:   1,
			jsonResponse: `
			[{
				"id": "Wyuiogb24tc7Tiob24t789yp",    
				"created": "2025-06-30T15:08:09.370896696Z",
				"updated": "2025-06-30T15:08:12.10611184Z",    
				"url": "https://some-news-website.org/something-great-happened/",
				"title": "Something Great & Awesome Happened",
				"site_name": "Awesome Website",
				"site": "some-news-website.org",
				"published": "2015-05-22T11:00:56Z",
				"authors": [
					"John Doe"
				],
				"lang": "en",
				"type": "article",
				"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
				"is_deleted": false,
				"is_marked": false,
				"is_archived": false,    
				"resources": {
					"article": {
						"src": "http://readeck-instance.com:8002/api/bookmarks/Wyuiogb24tc7Tiob24t789yp/article"
					},
					"icon": {
						"src": "http://readeck-instance.com:8002/bm/5C/Wyuiogb24tc7Tiob24t789yp/img/icon.png",
						"width": 48,
						"height": 48
					},
					"image": {
						"src": "http://readeck-instance.com:8002/bm/5C/Wyuiogb24tc7Tiob24t789yp/img/image.jpeg",
						"width": 800,
						"height": 800
					},
					"log": {
						"src": "http://readeck-instance.com:8002/api/bookmarks/Wyuiogb24tc7Tiob24t789yp/x/log"
					},
					"props": {
						"src": "http://readeck-instance.com:8002/api/bookmarks/Wyuiogb24tc7Tiob24t789yp/x/props.json"
					},
					"thumbnail": {
						"src": "http://readeck-instance.com:8002/bm/5C/Wyuiogb24tc7Tiob24t789yp/img/thumbnail.jpeg",
						"width": 380,
						"height": 380
					}
				},
				"word_count": 2137,
				"reading_time": 10
			}]
			`,
			want: pocketapi.GetResponse{
				Status: 1,
				List: map[string]pocketapi.GetResponseItem{
					"Wyuiogb24tc7Tiob24t789yp": {
						ItemID:         "Wyuiogb24tc7Tiob24t789yp",
						Favorite:       "0",
						Status:         "0",
						TimeAdded:      "1751296089",
						TimeUpdated:    "1751296092",
						TimeFavorited:  "0",
						TopImageURL:    "http://readeck-instance.com:8002/bm/5C/Wyuiogb24tc7Tiob24t789yp/img/image.jpeg",
						ResolvedID:     "Wyuiogb24tc7Tiob24t789yp",
						GivenURL:       "https://some-news-website.org/something-great-happened/",
						GivenTitle:     "Something Great & Awesome Happened",
						ResolvedTitle:  "Something Great & Awesome Happened",
						ResolvedURL:    "https://some-news-website.org/something-great-happened/",
						Excerpt:        "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
						IsArticle:      "1",
						IsIndex:        "0",
						HasVideo:       "0",
						HasImage:       "1",
						WordCount:      "2137",
						Lang:           "en",
						TimeToRead:     10,
						DomainMetadata: &pocketapi.DomainMetadata{Name: "Awesome Website"},
						Authors: map[string]pocketapi.Author{
							digest("John Doe"): {AuthorID: digest("John Doe"), Name: "John Doe", ItemID: "Wyuiogb24tc7Tiob24t789yp"},
						},
						Images: nil,
						Image: &pocketapi.Image{
							ItemID:  "Wyuiogb24tc7Tiob24t789yp",
							ImageID: digest("http://readeck-instance.com:8002/bm/5C/Wyuiogb24tc7Tiob24t789yp/img/image.jpeg"),
							Src:     "http://readeck-instance.com:8002/bm/5C/Wyuiogb24tc7Tiob24t789yp/img/image.jpeg",
							Width:   "800",
							Height:  "800",
						},
					},
				},
				Total: 1,
			},
		},
		{
			name:         "Minimal Fields Two Items",
			responseCode: http.StatusOK,
			totalItems:   2,
			jsonResponse: `
			[{
				"id": "Wyuiogb24tc7Tiob24t789yp",    
				"created": "2025-06-30T15:08:09.370896696Z",
				"updated": "2025-06-30T15:08:12.10611184Z",    
				"url": "https://some-news-website.org/something-great-happened/",
				"title": "Something Great & Awesome Happened",
				"site": "some-news-website.org",
				"published": "2015-05-22T11:00:56Z",				
				"lang": "en",
				"type": "article",
				"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
				"is_deleted": false,
				"is_marked": false,
				"is_archived": false,    
				"resources": {
					"article": {
						"src": "http://readeck-instance.com:8002/api/bookmarks/Wyuiogb24tc7Tiob24t789yp/article"
					},
					"log": {
						"src": "http://readeck-instance.com:8002/api/bookmarks/Wyuiogb24tc7Tiob24t789yp/x/log"
					},
					"props": {
						"src": "http://readeck-instance.com:8002/api/bookmarks/Wyuiogb24tc7Tiob24t789yp/x/props.json"
					}
				},
				"word_count": 2137,
				"reading_time": 10
			}, {
				"id": "FGU6tvjhjvbl789568",    
				"created": "2025-06-30T15:08:09.370896696Z",
				"updated": "2025-06-30T15:08:12.10611184Z",    
				"url": "https://some-news-website.org/something-else-happened/",
				"title": "Something Else Great & Awesome Happened",
				"site": "some-news-website.org",
				"published": "2015-05-22T11:00:56Z",
				"lang": "en",
				"type": "article",
				"description": "Nanananananana batman!",
				"is_deleted": false,
				"is_marked": true,
				"is_archived": true,    
				"resources": {
					"article": {
						"src": "http://readeck-instance.com:8002/api/bookmarks/FGU6tvjhjvbl789568/article"
					},
					"log": {
						"src": "http://readeck-instance.com:8002/api/bookmarks/FGU6tvjhjvbl789568/x/log"
					},
					"props": {
						"src": "http://readeck-instance.com:8002/api/bookmarks/FGU6tvjhjvbl789568/x/props.json"
					}
				},
				"word_count": 123,
				"reading_time": 5
			}]
			`,
			want: pocketapi.GetResponse{
				Status: 1,
				List: map[string]pocketapi.GetResponseItem{
					"Wyuiogb24tc7Tiob24t789yp": {
						ItemID:        "Wyuiogb24tc7Tiob24t789yp",
						Favorite:      "0",
						Status:        "0",
						TimeAdded:     "1751296089",
						TimeUpdated:   "1751296092",
						TimeFavorited: "0",
						ResolvedID:    "Wyuiogb24tc7Tiob24t789yp",
						GivenURL:      "https://some-news-website.org/something-great-happened/",
						GivenTitle:    "Something Great & Awesome Happened",
						ResolvedTitle: "Something Great & Awesome Happened",
						ResolvedURL:   "https://some-news-website.org/something-great-happened/",
						Excerpt:       "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
						IsArticle:     "1",
						IsIndex:       "0",
						HasVideo:      "0",
						HasImage:      "0",
						WordCount:     "2137",
						Lang:          "en",
						TimeToRead:    10,
					},
					"FGU6tvjhjvbl789568": {
						ItemID:        "FGU6tvjhjvbl789568",
						Favorite:      "1",
						Status:        "1",
						TimeAdded:     "1751296089",
						TimeUpdated:   "1751296092",
						TimeFavorited: "1751296092",
						ResolvedID:    "FGU6tvjhjvbl789568",
						GivenURL:      "https://some-news-website.org/something-else-happened/",
						GivenTitle:    "Something Else Great & Awesome Happened",
						ResolvedTitle: "Something Else Great & Awesome Happened",
						ResolvedURL:   "https://some-news-website.org/something-else-happened/",
						Excerpt:       "Nanananananana batman!",
						IsArticle:     "1",
						IsIndex:       "0",
						HasVideo:      "0",
						HasImage:      "0",
						WordCount:     "123",
						Lang:          "en",
						TimeToRead:    5,
					},
				},
				Total: 2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Total-Count", strconv.Itoa(tc.totalItems))
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
