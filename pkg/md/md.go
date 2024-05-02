// Package md provides common markdown utilities, used by build-tools and learn
// platform.
package md

import (
	"bytes"

	"rsc.io/markdown"
)

// Node is a subset of markdown.Block and markdown.Inline interfaces.
type Node interface {
	PrintHTML(*bytes.Buffer)
}

// Walk visits a markdown AST node, executes function f on it and recursively
// visits all children.
func Walk(node Node, f func(Node)) {
	f(node)
	switch n := node.(type) {
	case *markdown.Del:
		walkNodes(n.Inner, f)
	case *markdown.Document:
		walkNodes(n.Blocks, f)
	case *markdown.Emph:
		walkNodes(n.Inner, f)
	case *markdown.Heading:
		Walk(n.Text, f)
	case *markdown.Image:
		walkNodes(n.Inner, f)
	case *markdown.Item:
		walkNodes(n.Blocks, f)
	case *markdown.Link:
		walkNodes(n.Inner, f)
	case *markdown.List:
		walkNodes(n.Items, f)
	case *markdown.Paragraph:
		Walk(n.Text, f)
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

func walkNodes[T Node](nodes []T, f func(Node)) {
	for _, n := range nodes {
		Walk(n, f)
	}
}
