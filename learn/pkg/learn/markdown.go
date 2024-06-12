package learn

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"evylang.dev/evy/pkg/md"
	"gopkg.in/yaml.v3"
	"rsc.io/markdown"
)

// format formats YAML frontmatter, fenced by "---", followed by markdown
// content.
func format(frontmatter any, doc *markdown.Document) ([]byte, error) {
	b, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	buf.WriteString("---\n")
	buf.Write(b)
	buf.WriteString("---\n\n")
	buf.WriteString(markdown.ToMarkdown(doc))
	return buf.Bytes(), nil
}

// readSplitMDFile returns contents of filename split into frontmatter and
// markdown string.
func readSplitMDFile(filename string) (string, string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", "", fmt.Errorf("cannot process Markdown file: %w", err)
	}
	str := trimLeftSpace(string(b))

	if !strings.HasPrefix(str, "---\n") {
		return "", str, nil
	}
	end := strings.Index(str, "\n---\n")
	if end == -1 {
		return "", "", fmt.Errorf("%w: no closing ---", ErrInvalidFrontmatter)
	}
	frontmatter := str[:end+1]
	md := trimLeftSpace(str[len(frontmatter)+4:])
	return frontmatter, md, nil
}

func parseMD(rawMD string) *markdown.Document {
	parser := markdown.Parser{AutoLinkText: true, TaskListItems: true}
	return parser.Parse(rawMD)
}

func trimLeftSpace(str string) string {
	return strings.TrimLeftFunc(str, unicode.IsSpace)
}

func collectMDLinks(doc *markdown.Document) []string {
	var mdLinks []string
	md.Walk(doc, func(n md.Node) {
		mdl, ok := n.(*markdown.Link)
		if !ok {
			return
		}
		u, err := url.Parse(mdl.URL)
		if err != nil || u.IsAbs() {
			return
		}
		if filepath.Ext(u.Path) == ".md" {
			mdLinks = append(mdLinks, u.Path)
		}
	})
	return mdLinks
}

func extractName(doc *markdown.Document) (string, error) {
	for _, b := range doc.Blocks {
		h, ok := b.(*markdown.Heading)
		if !ok {
			continue
		}
		buf := &bytes.Buffer{}
		for _, inline := range h.Text.Inline {
			inline.PrintText(buf)
		}
		return buf.String(), nil
	}
	return "", fmt.Errorf("%w: no heading found", ErrBadMarkdownStructure)
}
