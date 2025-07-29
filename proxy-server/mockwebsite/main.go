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

// This file spins up a mock website to be made available to backends for extraction during tests.
package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// A 10x10 blue square PNG.
const testImage = "iVBORw0KGgoAAAANSUhEUgAAAAoAAAAKCAYAAACNMs+9AAAABHNCSVQICAgIfAhkiAAAABhJREFUGJVjZJjy/z8DEYCJGEWjCqmnEAAnjQKmJi5fSQAAAABJRU5ErkJggg=="

var port = flag.Int("port", 9090, "HTTP port to listen on")

func main() {
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		log.Printf("Mock website got request at %s", r.URL.String())
		if strings.HasSuffix(r.URL.Path, "png") {
			w.Header().Set("Content-Type", "image/png")
			imageData, err := base64.StdEncoding.DecodeString(testImage)
			if err != nil {
				log.Printf("Error decoding image base64: %v", err)
			}
			w.Write(imageData)
		}
		if strings.HasSuffix(r.URL.Path, "html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			title := r.URL.Query().Get("title")

			fmt.Fprintf(w, `
			<html>
			<head><title>%s</title></head>
			<body>
			  <h1>%s</h1>
			  <div>
			    <p><img src="/image1.png" /></p>
			  </div>
				<img src="/image2.png" />
			</body>
			</html>
			`, title, title)
		}
	})

	fmt.Printf("Mock website listening on http://localhost:%d\n", *port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Print(err)
	}
}
