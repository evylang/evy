package md

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
	"rsc.io/markdown"
)

func TestUndecorate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain",
			input: "banana",
			want:  "banana",
		},
		{
			name:  "strong",
			input: "**banana**",
			want:  "banana",
		},
		{
			name:  "emph code",
			input: "_`banana`_",
			want:  "banana",
		},
		{
			name:  "code",
			input: "`_banana_`",
			want:  "_banana_",
		},
		{
			name:  "link",
			input: `My evy [link](https://example.com "link title")`,
			want:  "My evy link",
		},
		{
			name:  "formatted link with title",
			input: `[My **evy** link](https://example.com "link title")`,
			want:  "My evy link",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := markdown.Parser{}
			doc := p.Parse(tt.input)
			assert.Equal(t, 1, len(doc.Blocks))
			paragraph, ok := doc.Blocks[0].(*markdown.Paragraph)
			assert.Equal(t, true, ok, "want: *markdown.Paragraph, got: %T", doc.Blocks[0])
			got := Undecorate(paragraph.Text.Inline)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUndecorateWithInline(t *testing.T) {
	tests := []struct {
		name   string
		inline markdown.Inline
		want   string
	}{
		{
			name:   "empty plain",
			inline: &markdown.Plain{},
			want:   "",
		},
		{
			name:   "empty link",
			inline: &markdown.Link{},
			want:   "",
		},
		{
			name:   "plain",
			inline: &markdown.Plain{Text: "banana"},
			want:   "banana",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Undecorate(tt.inline)
			assert.Equal(t, tt.want, got)
		})
	}
}
