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

package pocketapi

type DomainMetadata struct {
	Name          string `json:"name,omitempty"`
	Logo          string `json:"logo,omitempty"`
	GreyscaleLogo string `json:"greyscale_logo,omitempty"`
}

type Image struct {
	ItemID  string `json:"item_id"`
	ImageID string `json:"image_id"`
	Src     string `json:"src"`
	Width   string `json:"width,omitempty"`
	Height  string `json:"height,omitempty"`
	Credit  string `json:"credit,omitempty"`
	Caption string `json:"caption,omitempty"`
}

type Author struct {
	AuthorID string `json:"author_id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	ItemID   string `json:"item_id"`
}
