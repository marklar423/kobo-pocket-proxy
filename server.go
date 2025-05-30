package main

import (
	"flag"
	"fmt"
	"net/http"
)

var port = flag.Int("port", 8080, "HTTP port to listen on")

func getArticles(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get"))
}

func modifyArticles(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Modify"))
}

func articleText(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Text"))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/v3/get", getArticles)
	mux.HandleFunc("/v3/send", modifyArticles)
	mux.HandleFunc("/v3beta/text", articleText)

	fmt.Printf("Listening on http://localhost:%d\n", *port)

	http.ListenAndServe(fmt.Sprintf(":%d", *port), mux)
}
