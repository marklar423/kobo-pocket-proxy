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
		name string
		text string
		want pocketapi.ArticleTextResponse
	}{
		{
			name: "Basic",
			text: "<div><img src=\"http://test.com/img.png\" />",
			want: pocketapi.ArticleTextResponse{
				Article: "<div><!--IMG_1--></div>",
				Images: map[string]pocketapi.Image{
					"1": {
						ImageID: "1",
						Src:     "http://test.com/img.png",
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got pocketapi.ArticleTextResponse
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
