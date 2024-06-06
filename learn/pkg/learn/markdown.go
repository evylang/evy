package learn

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

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

func md2HTML(rawMD string) string {
	doc := parseMD(rawMD)
	buf := &bytes.Buffer{}
	buf.WriteString(questionPrefixHTML)
	doc.PrintHTML(buf)
	buf.WriteString(questionSuffixHTML)
	return buf.String()
}

func trimLeftSpace(str string) string {
	return strings.TrimLeftFunc(str, unicode.IsSpace)
}
