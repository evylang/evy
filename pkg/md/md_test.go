package md

import (
	"fmt"
	"os"
	"testing"

	"evylang.dev/evy/pkg/assert"
	"rsc.io/markdown"
)

func TestWalkBasic(t *testing.T) {
	var types []string
	f := func(n Node) {
		t := fmt.Sprintf("%T", n)
		types = append(types, t)
	}
	parser := markdown.Parser{}
	document := parser.Parse(`# Heading 1`)
	Walk(document, f)
	want := []string{
		"*markdown.Document",
		"*markdown.Heading",
		"*markdown.Text",
		"*markdown.Plain",
	}

	assert.Equal(t, want, types)
}

func TestWalk(t *testing.T) {
	var types []string
	f := func(n Node) {
		t := fmt.Sprintf("%T", n)
		types = append(types, t)
	}
	parser := markdown.Parser{}
	mdContent, err := os.ReadFile("testdata/sample/README.md")
	assert.NoError(t, err)
	document := parser.Parse(string(mdContent))
	Walk(document, f)
	want := []string{
		"*markdown.Document",
		"*markdown.Heading",
		"*markdown.Text",
		"*markdown.Plain",
		"*markdown.Heading",
		"*markdown.Text",
		"*markdown.Plain",
		"*markdown.Heading",
		"*markdown.Text",
		"*markdown.Plain",
		"*markdown.Paragraph",
		"*markdown.Text",
		"*markdown.Link",
		"*markdown.Plain",
		"*markdown.CodeBlock",
		"*markdown.Paragraph",
		"*markdown.Text",
		"*markdown.Strong",
		"*markdown.Plain",
		"*markdown.List",
		"*markdown.Item",
		"*markdown.Text",
		"*markdown.Plain",
		"*markdown.Item",
		"*markdown.Text",
		"*markdown.Plain",
		"*markdown.Item",
		"*markdown.Text",
		"*markdown.Plain",
	}

	assert.Equal(t, want, types)
}
