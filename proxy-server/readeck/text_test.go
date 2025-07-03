package readeck

import (
	"io"
	"proxyserver/pocketapi"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseArticleText(t *testing.T) {
	testCases := []struct {
		name   string
		text   string
		itemID string
		want   pocketapi.ArticleTextResponse
	}{
		{
			name:   "Empty",
			itemID: "item123",
			text:   "",
			want: pocketapi.ArticleTextResponse{
				ItemID:  "item123",
				Article: "<div></div>",
				Images:  map[string]pocketapi.Image{},
			},
		},
		{
			name:   "Basic",
			itemID: "item123",
			text:   "<div><img src=\"http://test.com/img.png\" /></div>",
			want: pocketapi.ArticleTextResponse{
				ItemID:  "item123",
				Article: "<div><div><!--IMG_1--></div></div>",
				Images: map[string]pocketapi.Image{
					"1": {
						ItemID:  "item123",
						ImageID: "1",
						Src:     "http://test.com/img.png",
					},
				},
			},
		},
		{
			name:   "Malformed",
			itemID: "item123",
			text:   "<div><img src=\"http://test.com/img.png\" />",
			want: pocketapi.ArticleTextResponse{
				ItemID:  "item123",
				Article: "<div><div><!--IMG_1--></div></div>",
				Images: map[string]pocketapi.Image{
					"1": {
						ItemID:  "item123",
						ImageID: "1",
						Src:     "http://test.com/img.png",
					},
				},
			},
		},
		{
			name:   "Multiple Elements",
			itemID: "item123",
			text:   "<div>test</div><div><img src=\"http://test.com/img.png\" /></div>",
			want: pocketapi.ArticleTextResponse{
				ItemID:  "item123",
				Article: "<div><div>test</div><div><!--IMG_1--></div></div>",
				Images: map[string]pocketapi.Image{
					"1": {
						ItemID:  "item123",
						ImageID: "1",
						Src:     "http://test.com/img.png",
					},
				},
			},
		},
		{
			name:   "Multiple Images",
			itemID: "item123",
			text: `
			<div>test</div>
			<p><img src="http://test.com/img.png" /></p>
			<div><figure><img src="http://test.com/img2.png" height="100" width="200" /></figure></div>
			`,
			want: pocketapi.ArticleTextResponse{
				ItemID: "item123",
				Article: `<div><div>test</div>
			<p><!--IMG_1--></p>
			<div><figure><!--IMG_2--></figure></div>
			</div>`,
				Images: map[string]pocketapi.Image{
					"1": {
						ItemID:  "item123",
						ImageID: "1",
						Src:     "http://test.com/img.png",
					},
					"2": {
						ItemID:  "item123",
						ImageID: "2",
						Src:     "http://test.com/img2.png",
						Height:  "100",
						Width:   "200",
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := pocketapi.ArticleTextResponse{
				ItemID: tc.itemID,
			}
			if err := parseArticleText(io.NopCloser(strings.NewReader(tc.text)), &got); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			tc.want.ContentLength = strconv.Itoa(len(tc.want.Article))

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("parseArticleText mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
