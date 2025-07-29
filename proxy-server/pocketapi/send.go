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

// SendAction contains the union of all possible fields in an action.
type SendAction struct {
	Action string `json:"action"`
	ItemID string `json:"item_id"`
	Time   int    `json:"time"`
	URL    string `json:"url"`
}

type SendRequest struct {
	AccessToken string       `json:"access_token"`
	Actions     []SendAction `json:"actions"`
	ConsumerKey string       `json:"consumer_key"`
}

type SendError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    int    `json:"code"`
}

type SendResponse struct {
	// 0 = failure, 1 = success
	Status        int          `json:"status"`
	ActionErrors  []*SendError `json:"action_errors"`
	ActionResults []bool       `json:"action_results"`
}
