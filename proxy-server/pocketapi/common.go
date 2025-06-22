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
