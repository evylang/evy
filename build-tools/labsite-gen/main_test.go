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
	}, {
		name: "codeblocks",
		in: `
# [>] hallo

` + "```evy" + `
print 1

print 2
` + "```" + `
`,
		want: `
<details>
<summary>hallo</summary>
<pre><code class="language-evy">print 1

print 2
</code></pre>
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

var nextButtonTests = []struct {
	name     string
	in       string
	wantMD   string
	wantHTML string
}{
	{
		name:     "no marker",
		in:       "hallo",
		wantMD:   "hallo\n",
		wantHTML: "<p>hallo</p>\n",
	},
	{
		name:     "single button",
		in:       `---`,
		wantMD:   `<button class="next-btn">Next</button>` + "\n",
		wantHTML: `<p><button class="next-btn">Next</button></p>` + "\n",
	},
	{
		name: "multiple buttons",
		in: `# Heading

---

paragraph

---`,
		wantMD: `
# Heading

<button class="next-btn">Next</button>

paragraph

<button class="next-btn">Next</button>
`[1:],
		wantHTML: `
<h1>Heading</h1>
<p><button class="next-btn">Next</button></p>
<p>paragraph</p>
<p><button class="next-btn">Next</button></p>
`[1:],
	},
}

func TestReplaceNextButton(t *testing.T) {
	for _, tt := range nextButtonTests {
		t.Run(tt.name, func(t *testing.T) {
			p := markdown.Parser{}
			doc := p.Parse(tt.in)
			replaceNextButton(doc)
			gotMD := markdown.Format(doc)
			assert.Equal(t, tt.wantMD, gotMD)
			gotHTML := markdown.ToHTML(doc)
			assert.Equal(t, tt.wantHTML, gotHTML)
		})
	}
}
