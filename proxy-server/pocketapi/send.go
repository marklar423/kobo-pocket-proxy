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
