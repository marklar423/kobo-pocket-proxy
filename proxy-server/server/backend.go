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

package server

import (
	"proxyserver/pocketapi"
	"time"
)

type Backend interface {
	Get(req pocketapi.GetRequest) (pocketapi.GetResponse, error)
	ArticleText(url string) (pocketapi.ArticleTextResponse, error)
	Add(url string, title string, time time.Time) error
	Archive(itemID string, time time.Time) error
	Unarchive(itemID string, time time.Time) error
	Delete(itemID string, time time.Time) error
	Favorite(itemID string, time time.Time) error
	Unfavorite(itemID string, time time.Time) error
}
