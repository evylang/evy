package md

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
	"rsc.io/markdown"
)

func TestUpdateRelLink(t *testing.T) {
	tests := []struct {
		name string
		mdl  *markdown.Link
		want string
	}{
		{
			name: "empty",
			mdl:  &markdown.Link{},
			want: "",
		},
		{
			name: "relative",
			mdl:  &markdown.Link{URL: "banana.md"},
			want: "banana.html",
		},
		{
			name: "relative-README.md",
			mdl:  &markdown.Link{URL: "banana/README.md"},
			want: "banana/index.html",
		},
		{
			name: "absolute",
			mdl:  &markdown.Link{URL: "http://example.com/banana.md"},
			want: "http://example.com/banana.md",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RewriteLink(tt.mdl)
			assert.Equal(t, tt.want, tt.mdl.URL)
		})
	}
}
