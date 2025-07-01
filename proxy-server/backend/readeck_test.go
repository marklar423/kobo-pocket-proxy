package backend

import (
	"net/http"
	"net/http/httptest"
	"proxyserver/pocketapi"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadeck_GetRequest(t *testing.T) {
	testCases := []struct {
		name    string
		request pocketapi.GetRequest
		want    map[string][]string
	}{
		{
			name:    "Empty",
			request: pocketapi.GetRequest{},
			want:    map[string][]string{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				if r.Method != http.MethodGet {
					t.Errorf("Unexpected HTTP method, want GET got %s", r.Method)
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

}
