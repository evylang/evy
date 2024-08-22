package learn

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
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
	buf.WriteString(markdown.Format(doc))
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
	parser := markdown.Parser{AutoLinkText: true, TaskList: true}
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
	heading, err := extractFirstHeading(doc)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	for _, inline := range heading.Text.Inline {
		buf.WriteString(md.Undecorate(inline))
	}
	return buf.String(), nil
}

func extractFirstHeading(doc *markdown.Document) (*markdown.Heading, error) {
	for _, b := range doc.Blocks {
		if h, ok := b.(*markdown.Heading); ok {
			return h, nil
		}
	}
	return nil, fmt.Errorf("%w: no heading found", ErrBadMarkdownStructure)
}

func toHTML(doc *markdown.Document) string {
	buf := &bytes.Buffer{}
	for _, block := range doc.Blocks {
		printHTML(block, buf)
	}
	return buf.String()
}

func printHTML(block markdown.Block, buf *bytes.Buffer) {
	if alertType, ok := alertBlock(block); ok {
		printAlert(block.(*markdown.Quote), buf, alertType)
	} else {
		buf.WriteString(markdown.ToHTML(block))
	}
}

func alertBlock(block markdown.Block) (string, bool) {
	quote, ok := block.(*markdown.Quote)
	if !ok || len(quote.Blocks) == 0 {
		return "", false
	}
	paragraph, ok := quote.Blocks[0].(*markdown.Paragraph)
	if !ok || len(paragraph.Text.Inline) == 0 {
		return "", false
	}
	plain, ok := paragraph.Text.Inline[0].(*markdown.Plain)
	if !ok {
		return "", false
	}
	text := strings.TrimSpace(plain.Text)
	if !strings.HasPrefix(text, "[!") {
		return "", false
	}
	idx := strings.Index(text, "]")
	if idx == -1 {
		return "", false
	}
	alertType := text[2:idx]
	types := []string{"NOTE", "TIP", "IMPORTANT", "WARNING", "CAUTION"}
	// skip first inline in
	if !slices.Contains(types, alertType) {
		return "", false
	}
	paragraph.Text.Inline = paragraph.Text.Inline[1:]
	return strings.ToLower(alertType), true
}

func printAlert(quote *markdown.Quote, buf *bytes.Buffer, alertType string) {
	buf.WriteString(`<div class="alert alert-` + alertType + `">`)
	buf.WriteString(`<p class="alert-title">`)
	buf.WriteString(`<svg viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="`)
	buf.WriteString(alertIconPath[alertType])
	buf.WriteString(`"></path></svg>`)
	buf.WriteString(strings.Title(alertType)) //nolint: staticcheck // we can savely use it here as we know all strings we want to use and have no punctuation.
	buf.WriteString(`</p>`)
	for _, block := range quote.Blocks {
		buf.WriteString(markdown.ToHTML(block))
	}
	buf.WriteString(`</div>`)
}

var alertIconPath = map[string]string{
	"note":      "M 0 8 a 8 8 0 1 1 16 0 A 8 8 0 0 1 0 8 Z m 8 -6.5 a 6.5 6.5 0 1 0 0 13 a 6.5 6.5 0 0 0 0 -13 Z M 6.5 7.75 A 0.75 0.75 0 0 1 7.25 7 h 1 a 0.75 0.75 0 0 1 0.75 0.75 v 2.75 h 0.25 a 0.75 0.75 0 0 1 0 1.5 h -2 a 0.75 0.75 0 0 1 0 -1.5 h 0.25 v -2 h -0.25 a 0.75 0.75 0 0 1 -0.75 -0.75 Z M 8 6 a 1 1 0 1 1 0 -2 a 1 1 0 0 1 0 2 Z",
	"tip":       "M8 1.5c-2.363 0-4 1.69-4 3.75 0 .984.424 1.625.984 2.304l.214.253c.223.264.47.556.673.848.284.411.537.896.621 1.49a.75.75 0 0 1-1.484.211c-.04-.282-.163-.547-.37-.847a8.456 8.456 0 0 0-.542-.68c-.084-.1-.173-.205-.268-.32C3.201 7.75 2.5 6.766 2.5 5.25 2.5 2.31 4.863 0 8 0s5.5 2.31 5.5 5.25c0 1.516-.701 2.5-1.328 3.259-.095.115-.184.22-.268.319-.207.245-.383.453-.541.681-.208.3-.33.565-.37.847a.751.751 0 0 1-1.485-.212c.084-.593.337-1.078.621-1.489.203-.292.45-.584.673-.848.075-.088.147-.173.213-.253.561-.679.985-1.32.985-2.304 0-2.06-1.637-3.75-4-3.75ZM5.75 12h4.5a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1 0-1.5ZM6 15.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 0 1.5h-2.5a.75.75 0 0 1-.75-.75Z",
	"important": "M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v9.5A1.75 1.75 0 0 1 14.25 13H8.06l-2.573 2.573A1.458 1.458 0 0 1 3 14.543V13H1.75A1.75 1.75 0 0 1 0 11.25Zm1.75-.25a.25.25 0 0 0-.25.25v9.5c0 .138.112.25.25.25h2a.75.75 0 0 1 .75.75v2.19l2.72-2.72a.749.749 0 0 1 .53-.22h6.5a.25.25 0 0 0 .25-.25v-9.5a.25.25 0 0 0-.25-.25Zm7 2.25v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 9a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z",
	"warning":   "M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 11a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z",
	"caution":   "M4.47.22A.749.749 0 0 1 5 0h6c.199 0 .389.079.53.22l4.25 4.25c.141.14.22.331.22.53v6a.749.749 0 0 1-.22.53l-4.25 4.25A.749.749 0 0 1 11 16H5a.749.749 0 0 1-.53-.22L.22 11.53A.749.749 0 0 1 0 11V5c0-.199.079-.389.22-.53Zm.84 1.28L1.5 5.31v5.38l3.81 3.81h5.38l3.81-3.81V5.31L10.69 1.5ZM8 4a.75.75 0 0 1 .75.75v3.5a.75.75 0 0 1-1.5 0v-3.5A.75.75 0 0 1 8 4Zm0 8a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z",
}
