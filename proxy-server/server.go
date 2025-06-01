package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
)

var port = flag.Int("port", 8080, "HTTP port to listen on")

func getArticles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	w.Write([]byte("Get"))
}

func modifyArticles(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Modify"))
}

type articleRequest struct {
	Url          string
	Access_Token string
}

type author struct {
}

type image struct {
}

type articleResponse struct {
	Resolved_id   string            `json:"resolved_id"`
	ResolvedUrl   string            `json:"resolvedUrl"`
	Host          string            `json:"host"`
	Title         string            `json:"title"`
	DatePublished string            `json:"datePublished"`
	TimePublished string            `json:"timePublished"`
	ResponseCode  int               `json:"responseCode"`
	Excerpt       string            `json:"excerpt"`
	Authors       map[string]author `json:"authors"`
	Images        map[string]image  `json:"images"`
	Videos        string            `json:"videos"`
	WordCount     int               `json:"wordCount"`
	IsArticle     int               `json:"isArticle"`
	IsVideo       int               `json:"isVideo"`
	IsIndex       int               `json:"isIndex"`
	UsedFallback  int               `json:"usedFallback"`
	RequiresLogin int               `json:"requiresLogin"`
	Lang          string            `json:"lang"`
	TopImageUrl   string            `json:"topImageUrl"`
	Article       string            `json:"article"`
}

func articleText(w http.ResponseWriter, r *http.Request) {
	var req articleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse request body: %v", err.Error()), http.StatusBadRequest)
		return
	}

	var res articleResponse
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, fmt.Sprintf("Unable to serialize response: %v", err.Error()), http.StatusInternalServerError)
		return
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/v3/get", getArticles)
	mux.HandleFunc("/v3/send", modifyArticles)
	mux.HandleFunc("/v3beta/text", articleText)

	fmt.Printf("Listening on http://localhost:%d\n", *port)

	http.ListenAndServe(fmt.Sprintf(":%d", *port), mux)
}
