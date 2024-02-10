package main

import (
	"bytes"

	"rsc.io/markdown"
)

// node is a subset of markdown.Block and markdown.Inline interfaces.
type node interface {
	PrintHTML(*bytes.Buffer)
}

type walkFunc func(node) node

func walk(node node, f walkFunc) node {
	switch n := node.(type) {
	case *markdown.Paragraph:
		n.Text = walk(n.Text, f).(*markdown.Text)
	case *markdown.Heading:
		n.Text = walk(n.Text, f).(*markdown.Text)
	case *markdown.Del:
		n.Inner = walkNodes(n.Inner, f)
	case *markdown.Document:
		n.Blocks = walkNodes(n.Blocks, f)
	case *markdown.Emph:
		n.Inner = walkNodes(n.Inner, f)
	case *markdown.Image:
		n.Inner = walkNodes(n.Inner, f)
	case *markdown.Item:
		n.Blocks = walkNodes(n.Blocks, f)
	case *markdown.Link:
		n.Inner = walkNodes(n.Inner, f)
	case *markdown.List:
		n.Items = walkNodes(n.Items, f)
	case *markdown.Quote:
		n.Blocks = walkNodes(n.Blocks, f)
	case *markdown.Strong:
		n.Inner = walkNodes(n.Inner, f)
	case *markdown.Table:
		n.Header = walkNodes(n.Header, f)
		for i, row := range n.Rows {
			n.Rows[i] = walkNodes(row, f)
		}
	case *markdown.Text:
		n.Inline = walkNodes(n.Inline, f)
	}
	return f(node)
}

func walkNodes[T node](nodes []T, f walkFunc) []T {
	result := make([]T, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, walk(n, f).(T))
	}
	return result
}
