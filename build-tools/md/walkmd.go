package main

import (
	"bytes"

	"rsc.io/markdown"
)

// node is a subset of markdown.Block and markdown.Inline interfaces.
type node interface {
	PrintHTML(*bytes.Buffer)
}

func walk(node node, f func(node)) {
	f(node)
	switch n := node.(type) {
	case *markdown.Del:
		walkNodes(n.Inner, f)
	case *markdown.Document:
		walkNodes(n.Blocks, f)
	case *markdown.Emph:
		walkNodes(n.Inner, f)
	case *markdown.Heading:
		walk(n.Text, f)
	case *markdown.Image:
		walkNodes(n.Inner, f)
	case *markdown.Item:
		walkNodes(n.Blocks, f)
	case *markdown.Link:
		walkNodes(n.Inner, f)
	case *markdown.List:
		walkNodes(n.Items, f)
	case *markdown.Paragraph:
		walk(n.Text, f)
	case *markdown.Quote:
		walkNodes(n.Blocks, f)
	case *markdown.Strong:
		walkNodes(n.Inner, f)
	case *markdown.Table:
		walkNodes(n.Header, f)
		for _, row := range n.Rows {
			walkNodes(row, f)
		}
	case *markdown.Text:
		walkNodes(n.Inline, f)
	}
}

func walkNodes[T node](nodes []T, f func(node)) {
	for _, n := range nodes {
		walk(n, f)
	}
}
