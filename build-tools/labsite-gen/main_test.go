package main

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
	"rsc.io/markdown"
)

var collapseTests = []struct {
	name string
	in   string
	want string
}{
	{
		name: "no marker",
		in:   "hallo",
		want: "hallo\n",
	}, {
		name: "marker, no heading",
		in:   "[>] hallo",
		want: "[>] hallo\n",
	}, {
		name: "marker, no body",
		in:   "# [>] hallo",
		want: `
<details>
<summary>hallo</summary>
</details>
`[1:],
	}, {
		name: "marker, with body",
		in: `# [>] hallo
anybody here?`,
		want: `
<details>
<summary>hallo</summary>
<p>anybody here?</p>
</details>
`[1:],
	}, {
		name: "2 markers and weird white space (weird is good)",
		in: `# [>]hallo
# [>]    	  goodbye`,
		want: `
<details>
<summary>hallo</summary>
</details>

<details>
<summary>goodbye</summary>
</details>
`[1:],
	}, {
		name: "2 nested markers no body",
		in: `
# [>] hallo
## [>] goodbye`[1:],
		want: `
<details>
<summary>hallo</summary>
<details>
<summary>goodbye</summary>
</details>
</details>
`[1:],
	},
}

func TestCollapse(t *testing.T) {
	for _, tt := range collapseTests {
		t.Run(tt.name, func(t *testing.T) {
			p := markdown.Parser{}
			doc := p.Parse(tt.in)
			doc.Blocks = collapse(doc.Blocks)
			got := markdown.Format(doc)
			assert.Equal(t, tt.want, got)
		})
	}
}
