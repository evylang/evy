// Command labsite-gen generates the HTML fragments files from Markdown source
// files.
//
// Usage: labsite-gen <MD-FILE> <HTMLF-FILE>
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

	"rsc.io/markdown"
)

const marker = "[>]"

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
	doc.Blocks = collapse(doc.Blocks)
	html := markdown.ToHTML(doc)

	return os.WriteFile(htmlfFile, []byte(html), 0o666)
}

func parse(md string) *markdown.Document {
	p := markdown.Parser{Table: true}
	return p.Parse(md)
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
	return ok && strings.HasPrefix(plain.Text, marker)
}

func deleteCollapseMarker(heading *markdown.Heading) {
	// assumes isCollapsible returned true
	plain := heading.Text.Inline[0].(*markdown.Plain)
	s := strings.TrimPrefix(plain.Text, marker)
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	plain.Text = s
}

func findEndIdx(level, start int, blocks []markdown.Block) int {
	for i := start; i < len(blocks); i++ {
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
	doc := parse(buf.String())
	return doc.Blocks[0].(*markdown.HTMLBlock)
}

func isThematicBreak(blocks []markdown.Block, idx int) bool {
	if idx >= len(blocks) {
		return false
	}
	_, ok := blocks[idx].(*markdown.ThematicBreak)
	return ok
}
