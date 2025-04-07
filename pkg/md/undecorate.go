package md

import (
	"fmt"
	"strings"

	"rsc.io/markdown"
)

// Undecorate returns a stripped down markdown representation of the given
// markdown.Inline: Strong has the `**` stripped, Code has the back-ticks
// stripped, Link only uses the text representation, no URL etc. Undecorate is
// used to extract a plain "name" for given inline elements.
func Undecorate(mdInline markdown.Inline) string {
	// Inlines, Plain, Escaped, Code, Strong, Emph, Del, Link, AutoLink, Image, SoftBreak, HardBreak, HTMLTag, Emoji, and Task.
	switch inline := mdInline.(type) {
	case *markdown.Plain:
		return inline.Text
	case *markdown.Escaped:
		return inline.Text
	case *markdown.Code:
		return inline.Text
	case *markdown.Strong:
		return Undecorate(inline.Inner)
	case *markdown.Emph:
		return Undecorate(inline.Inner)
	case *markdown.Del:
		return Undecorate(inline.Inner)
	case *markdown.Link:
		return Undecorate(inline.Inner)
	case *markdown.AutoLink:
		return inline.Text
	case *markdown.Image:
		return inline.Title
	case *markdown.SoftBreak:
		return ""
	case *markdown.HardBreak:
		return "\n"
	case *markdown.HTMLTag:
		return ""
	case *markdown.Emoji:
		return inline.Text
	case *markdown.Task:
		return ""
	case markdown.Inlines:
		s := make([]string, len(inline))
		for i, inl := range inline {
			s[i] = Undecorate(inl)
		}
		return strings.Join(s, "")
	}

	return fmt.Sprintf("cannot undecorate %T", mdInline)
}
