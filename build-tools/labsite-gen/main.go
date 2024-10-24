// Command labsite-gen generates the HTML fragments files from Markdown source
// files.
//
// Usage: labsite-gen <MD-FILE> <HTMLF-FILE>
package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"evylang.dev/evy/pkg/md"
	"rsc.io/markdown"
)

const (
	detailsMarker = "[>]"
)

var nextButton = newNextButton()

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: labsite-gen <MD-FILE> <HTMLF-FILE>")
		os.Exit(1)
	}
	err := run(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(mdFile, htmlfFile string) error {
	md, err := os.ReadFile(mdFile)
	if err != nil {
		return err
	}
	doc := parse(string(md))
	updateImgURL(doc, mdFile)
	updateLabLinks(doc)
	replaceNextButton(doc)
	doc.Blocks = collapse(doc.Blocks)
	html := markdown.ToHTML(doc)

	return os.WriteFile(htmlfFile, []byte(html), 0o666)
}

func parse(md string) *markdown.Document {
	p := markdown.Parser{Table: true}
	return p.Parse(md)
}

// Change relative links to work for frontend lab root which is where it will be
// requested from on https://lab.evy.dev. The generated .htmlf files are loaded
// as fragments and image links relative to these fragments do not work, they need
// to be relative to the initially loaded frontend/lab/index.html site.
func updateImgURL(doc *markdown.Document, mdFile string) {
	md.Walk(doc, func(n md.Node) {
		img, ok := n.(*markdown.Image)
		if !ok {
			return
		}
		u, err := url.Parse(img.URL)
		if err != nil || u.IsAbs() {
			return
		}
		// change img/circle.svg to samples/ifs/img/circle.svg
		u.Path = updateIMGPath(mdFile, u.Path)
		img.URL = u.String()
	})
}

func updateIMGPath(mdPath, imgPath string) string {
	labDir := filepath.Base(filepath.Dir(mdPath))
	return "samples/" + labDir + "/" + imgPath
}

func updateLabLinks(doc *markdown.Document) {
	md.Walk(doc, func(n md.Node) {
		link, ok := n.(*markdown.Link)
		if !ok {
			return
		}
		u, err := url.Parse(link.URL)
		if err != nil || u.IsAbs() || !strings.HasSuffix(u.Path, ".md") {
			return
		}
		// change ../loops/hsl.md and hsl.md to #hsl
		link.URL = "#" + strings.TrimSuffix(filepath.Base(u.Path), ".md")
	})
}

func replaceNextButton(doc *markdown.Document) {
	for i, block := range doc.Blocks {
		if _, ok := block.(*markdown.ThematicBreak); ok {
			doc.Blocks[i] = nextButton
		}
	}
}

func collapse(blocks []markdown.Block) []markdown.Block {
	var result []markdown.Block
	for idx := 0; idx < len(blocks); {
		block := blocks[idx]
		idx++
		if !isCollapsible(block) {
			result = append(result, block)
			continue
		}
		heading := block.(*markdown.Heading)
		deleteCollapseMarker(heading)
		end := findEndIdx(heading.Level, idx, blocks)
		inner := collapse(blocks[idx:end])
		detailsHTML := toDetailsHTML(heading, inner)
		result = append(result, detailsHTML)
		idx = end
		if isThematicBreak(blocks, idx) {
			// skip thematic break, it's only the collapsible end marker
			idx++
		}
	}
	return result
}

func isCollapsible(block markdown.Block) bool {
	heading, ok := block.(*markdown.Heading)
	if !ok || len(heading.Text.Inline) == 0 {
		return false
	}
	plain, ok := heading.Text.Inline[0].(*markdown.Plain)
	return ok && strings.HasPrefix(plain.Text, detailsMarker)
}

func deleteCollapseMarker(heading *markdown.Heading) {
	// assumes isCollapsible returned true
	plain := heading.Text.Inline[0].(*markdown.Plain)
	s := strings.TrimPrefix(plain.Text, detailsMarker)
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	plain.Text = s
}

func findEndIdx(level, start int, blocks []markdown.Block) int {
	for i := start; i < len(blocks); i++ {
		if blocks[i] == nextButton {
			return i
		}
		if _, ok := blocks[i].(*markdown.ThematicBreak); ok {
			return i
		}
		if heading, ok := blocks[i].(*markdown.Heading); ok {
			if heading.Level <= level {
				return i
			}
		}
	}
	return len(blocks)
}

func toDetailsHTML(heading *markdown.Heading, blocks []markdown.Block) *markdown.HTMLBlock {
	buf := bytes.Buffer{}
	buf.WriteString("<details>\n")
	buf.WriteString("<summary>")
	buf.WriteString(markdown.ToHTML(heading.Text))
	buf.WriteString("</summary>\n")
	for _, block := range blocks {
		buf.WriteString(markdown.ToHTML(block))
	}
	buf.WriteString("</details>")
	htmlBlock := &markdown.HTMLBlock{
		Text: strings.Split(buf.String(), "\n"),
	}
	return htmlBlock
}

func newNextButton() *markdown.Paragraph {
	doc := parse(`<button class="next-btn">Next</button>`)
	return doc.Blocks[0].(*markdown.Paragraph)
}

func isThematicBreak(blocks []markdown.Block, idx int) bool {
	if idx >= len(blocks) {
		return false
	}
	_, ok := blocks[idx].(*markdown.ThematicBreak)
	return ok
}
