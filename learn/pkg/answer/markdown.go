package answer

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
	"rsc.io/markdown"
)

var ErrUnimplemented = fmt.Errorf("not yet implemented") // todo: remove when done

// QuestionMarkdown is a markdown file with question frontmatter.
type QuestionMarkdown struct {
	Filename    string
	Frontmatter *QuestionFrontmatter
	Doc         *markdown.Document
}

func NewQuestionMarkdown(filename string) (*QuestionMarkdown, error) {
	frontmatterString, mdString, err := readSplitMDFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%w (%s)", err, filename)
	}
	fm, err := parseQuestionFrontmatter(frontmatterString)
	if err != nil {
		return nil, fmt.Errorf("%w (%s)", err, filename)
	}
	parser := markdown.Parser{AutoLinkText: true, TaskListItems: true}
	doc := parser.Parse(mdString)

	return &QuestionMarkdown{Filename: filename, Frontmatter: fm, Doc: doc}, nil
}

func (md *QuestionMarkdown) Seal(publicKey string) error {
	if err := md.Frontmatter.Seal(publicKey); err != nil {
		fmt.Errorf("%w (%s)", err, md.Filename)
	}
	return nil
}

func (md *QuestionMarkdown) Unseal(privateKey string) error {
	if err := md.Frontmatter.Unseal(privateKey); err != nil {
		fmt.Errorf("%w (%s)", err, md.Filename)
	}
	return nil
}

func (md *QuestionMarkdown) Format() (string, error) {
	return formatMD(md.Frontmatter, md.Doc)
}

func (md *QuestionMarkdown) Verify(key string) error {
	answer, err := md.Frontmatter.GetAnswer(key)
	if err != nil {
		return fmt.Errorf("%w (%s)", err, md.Filename)
	}
	return md.verifyAnswer(answer)
}

func (md *QuestionMarkdown) ExportAnswerkey(key string) (Answerkey, error) {
	answer, err := md.Frontmatter.GetAnswer(key)
	if err != nil {
		return nil, fmt.Errorf("%w (%s)", err, md.Filename)
	}
	if err := md.verifyAnswer(answer); err != nil {
		return nil, fmt.Errorf("%w (%s)", err, md.Filename)
	}
	return NewAnswerkey(md.Filename, answer)
}

func (md *QuestionMarkdown) verifyAnswer(answer Answer) error {
	model, err := NewModel(md)
	if err != nil {
		return fmt.Errorf("%w (%s)", err, md.Filename)
	}
	if err := model.Verify(answer); err != nil {
		return fmt.Errorf("%w (%s)", err, md.Filename)
	}
	return nil
}

func readSplitMDFile(filename string) (frontmatter string, md string, err error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", "", fmt.Errorf("cannot process Question Markdown file: %w", err)
	}
	str := trimLeftSpace(string(b))
	frontmatter, err = extractFrontmatterString(str)
	if err != nil {
		return "", "", err
	}
	md = trimLeftSpace(str[len(frontmatter)+6:])
	return frontmatter, md, nil
}

func parseQuestionFrontmatter(str string) (*QuestionFrontmatter, error) {
	fm := &QuestionFrontmatter{}
	if err := yaml.Unmarshal([]byte(str), fm); err != nil {
		return nil, fmt.Errorf("cannot process Question Markdown frontmatter: %w", err)
	}
	if err := fm.Validate(); err != nil {
		return nil, err
	}
	return fm, nil
}

func formatMD(frontmatter any, doc *markdown.Document) (string, error) {
	bytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", err
	}
	sb := strings.Builder{}
	sb.WriteString("---\n")
	sb.Write(bytes)
	sb.WriteString("---\n\n")
	sb.WriteString(markdown.ToMarkdown(doc))
	return sb.String(), nil
}

func trimLeftSpace(str string) string {
	return strings.TrimLeftFunc(str, unicode.IsSpace)
}

func extractFrontmatterString(str string) (string, error) {
	if !strings.HasPrefix(str, "---") {
		return "", ErrNoFrontmatter
	}
	end := strings.Index(str[3:], "\n---")
	if end == -1 {
		return "", fmt.Errorf("%w: no closing ---", ErrInvalidFrontmatter)
	}
	return str[3 : end+4], nil
}
