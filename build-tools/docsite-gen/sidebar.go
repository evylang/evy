package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"rsc.io/markdown"
)

// updateSidebar adds the headings of linked documents, linked via relative
// path to Markdown file, to the sidebar.
func updateSidebar(sidebarFile string, sidebar *markdown.Document, headings map[string][]heading) {
	sidebarDir := filepath.Dir(sidebarFile)
	for _, b := range sidebar.Blocks {
		list, ok := b.(*markdown.List)
		if !ok {
			continue
		}
		for _, block := range list.Items {
			item := block.(*markdown.Item)
			relMDF := relPath(item)
			if relMDF == "" {
				fmt.Fprintf(os.Stderr, "sidebar item not a relative link: %#v\n", item)
				continue
			}
			mdf := filepath.Join(sidebarDir, relMDF)
			addHeadingsToItem(item, relMDF, headings[mdf])
		}
	}
}

type levelUl struct {
	ul    *markdown.List
	level int
}

type ulStack []levelUl

func (s *ulStack) push(ul *markdown.List, level int) {
	*s = append(*s, levelUl{ul, level})
}

func (s *ulStack) pop() levelUl {
	res := s.peek()
	*s = (*s)[:len(*s)-1]
	return res
}

func (s *ulStack) peek() levelUl {
	return (*s)[len(*s)-1]
}

// addHeadingsToItem appends a list of headings to an item in the sidebar.
func addHeadingsToItem(item *markdown.Item, relMDF string, headings []heading) {
	var uls ulStack
	prevLevel := 0 // we will never pop the first ul, there are no headings with level 0
	for _, h := range headings {
		level := h.heading.Level
		if level > 3 {
			continue
		}
		if prevLevel < level {
			expander := &markdown.HTMLBlock{Text: []string{`<div class="expander"></div>`}}
			ul := &markdown.List{Bullet: '-'}
			uls.push(ul, prevLevel)
			item.Blocks = append(item.Blocks, expander, ul)
		}
		for uls.peek().level >= level {
			uls.pop()
		}

		item = newHeadingItem(h, relMDF)
		ul := uls.peek().ul
		ul.Items = append(ul.Items, item)

		prevLevel = level
	}
}

func newHeadingItem(h heading, relPath string) *markdown.Item {
	inner := h.heading.Text.Inline[1:] // slice off previously inserted anchor *markdown.HTMLTag
	link := &markdown.Link{
		Inner: inner,
		URL:   relPath + "#" + h.anchorID,
	}
	return &markdown.Item{
		Blocks: []markdown.Block{
			&markdown.Text{Inline: []markdown.Inline{link}},
		},
	}
}

func relPath(item *markdown.Item) string {
	text, ok := item.Blocks[0].(*markdown.Text)
	if !ok || len(text.Inline) == 0 {
		return ""
	}
	link, ok := text.Inline[0].(*markdown.Link)
	if !ok {
		return ""
	}
	u, err := url.Parse(link.URL)
	if err != nil {
		return ""
	}
	if u.IsAbs() {
		return ""
	}
	return u.Path
}
