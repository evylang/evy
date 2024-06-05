// Package learn provides data structures and tools for Evy learn resources.
// Question, exercises, units and courses are parsed from Markdown files with
// YAML frontmatter. The frontmatter serves as a small set of structured data
// associated with the unstructured Markdown file.
//
// Question can be verified to have the expected correct answer output match
// the question output. Questions, can seal (encrypt) their answers in the
// Frontmatter or unsealed (decrypted) them. We use this to avoid openly
// publishing the answerKey. Questions can also export their AnswerKeys into
// single big JSON object as used in Evy's persistent data store(Firestore).
// See the testdata/ directory for sample question and answers.
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

// newFrontmatterMarkdown creates a new frontmatter and markdown document for
// filename. Frontmatter contains the initial, parsed YAML data section  of
// the file, the markdown document contains the parsed the markdown AST for
// the remainder.
func newFrontmatterMarkdown(filename string) (*frontmatter, *markdown.Document, error) {
	frontmatterString, mdString, err := readSplitMDFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("%w (%s)", err, filename)
	}
	fm, err := parseFrontmatter(frontmatterString)
	if err != nil {
		return nil, nil, fmt.Errorf("%w (%s)", err, filename)
	}
	parser := markdown.Parser{AutoLinkText: true, TaskListItems: true}
	doc := parser.Parse(mdString)

	return fm, doc, nil
}

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
		return "", "", fmt.Errorf("cannot process Question Markdown file: %w", err)
	}
	str := trimLeftSpace(string(b))

	if !strings.HasPrefix(str, "---\n") {
		return "", "", ErrNoFrontmatter
	}
	end := strings.Index(str, "\n---\n")
	if end == -1 {
		return "", "", fmt.Errorf("%w: no closing ---", ErrInvalidFrontmatter)
	}
	frontmatter := str[:end+1]
	md := trimLeftSpace(str[len(frontmatter)+4:])
	return frontmatter, md, nil
}

func trimLeftSpace(str string) string {
	return strings.TrimLeftFunc(str, unicode.IsSpace)
}
