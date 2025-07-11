package readeck

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ReadeckConn struct {
	endpoint    string
	bearerToken string

	// A mapping article URLs to Readeck IDs.
	// Pocket just needs a URL to get article text, but Readeck requires an item ID,
	// so the solution here is to cache URLs and IDs in memory when Get() is called.
	//
	// It's not ideal, but the Kobo will usually list articles before downloading, and
	// so long as the proxy server isn't restarted in between it'll still have the ID.
	//
	// If this is insufficient, it should be easy to add code to automatically refresh
	// the cache on startup.
	urlIDCache map[string]string
}

func NewReadeckConn(endpoint string, bearerToken string) *ReadeckConn {
	return &ReadeckConn{
		endpoint:    endpoint,
		bearerToken: bearerToken,
		urlIDCache:  make(map[string]string),
	}
}

func (conn *ReadeckConn) createRequest(method, action string, body io.Reader) (*http.Request, error) {
	apiUrl := fmt.Sprintf("%s/api/%s", conn.endpoint, action)
	deckReq, err := http.NewRequest(method, apiUrl, body)
	if err != nil {
		return nil, err
	}
	deckReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", conn.bearerToken))
	return deckReq, nil
}

func digest(val string) string {
	h := sha1.New()
	h.Write([]byte(val))
	return hex.EncodeToString(h.Sum(nil))
}

type errorBody struct {
	Message string
	Status  int
}

func checkResponseCode(deckRes *http.Response) error {
	if deckRes.StatusCode >= 200 || deckRes.StatusCode <= 299 {
		return nil
	}
	var body errorBody
	if err := json.NewDecoder(deckRes.Body).Decode(&body); err != nil {
		return fmt.Errorf("error calling Readeck API: [%d] %s", deckRes.StatusCode, deckRes.Status)
	}
	return fmt.Errorf("error calling Readeck API: [%d] %s, More details: [%d] %s", deckRes.StatusCode, deckRes.Status, body.Status, body.Message)
}
