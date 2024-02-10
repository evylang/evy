package main

import (
	"bytes"
	"fmt"

	"rsc.io/markdown"
)

// anchoredHeading produces HTML with clickable link anchor, for example:
//
//	<h2 class="anchored">
//	  <a id="variables-and-declarations" href="#variables-and-declarations" class="anchor">
//	    <svg width="20px" height="20px"><use href="#icon-link" /></svg>
//	  </a>
//	  Variables and Declarations
//	</h2>
type anchoredHeading struct {
	*markdown.Heading
	anchor *markdown.HTMLTag
	id     string
}

func newAnchoredHeading(h *markdown.Heading, id string) *anchoredHeading {
	format := `
<a id="%s" href="#%s" class="anchor">
<svg width="20px" height="20px"><use href="#icon-link" /></svg>
</a>
`[1:]
	return &anchoredHeading{
		Heading: h,
		anchor:  &markdown.HTMLTag{Text: fmt.Sprintf(format, id, id)},
		id:      id,
	}
}

func (a *anchoredHeading) PrintHTML(buf *bytes.Buffer) {
	fmt.Fprintf(buf, `<h%d class="anchored">`+"\n", a.Level)
	a.anchor.PrintHTML(buf)
	a.Text.PrintHTML(buf)
	fmt.Fprintf(buf, "</h%d>\n", a.Level)
}
