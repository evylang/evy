package question

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rsc.io/markdown"
)

// Errors for the question package.
var (
	ErrBadMarkdownStructure = errors.New("bad Markdown structure")
	ErrInconsistentMdoel    = errors.New("inconsistency")
	ErrWrongAnswer          = errors.New("wrong answer")

	ErrSingleChoice          = errors.New("single-choice answer must be a single character a-z")
	ErrBadDirectoryStructure = errors.New("bad directory structure for course layout")

	ErrNoFrontmatter        = errors.New("no frontmatter found")
	ErrInvalidFrontmatter   = errors.New("invalid frontmatter")
	ErrWrongFrontmatterType = errors.New("wrong frontmatter type")
	ErrNoFrontmatterAnswer  = errors.New("no answer in frontmatter")
	ErrSealedAnswerNoKey    = errors.New("sealed answer without key in frontmatter")
	ErrSealedTooShort       = errors.New("sealed data is too short")
)

// Model represents a question and its answer choices as parsed and derived
// from the original Markdown file with frontmatter.
//
// In a Model the Question may be Evy source code, text output or image
// output. The Model AnswerChoices field is a list of Evy source code, text
// output and image output. In a verified Model, only the correct answer
// choice (output) matches the question (output). "Correct" means as
// specified in the Markdown Frontmatter.
type Model struct {
	Filename    string
	Doc         *markdown.Document
	Frontmatter *frontmatter

	Question      Renderer
	AnswerChoices []Renderer
	ResultType    ResultType

	withSealed bool
	privateKey string
	embeds     map[markdown.Block]embed // use to replace markdown Link or Image with codeBlock or inline SVG
	answerList markdown.Block           // use to output checkbox or radio buttons for List in HTML
}

type embed struct {
	id       string
	renderer Renderer
}
type fieldType uint

const (
	questionField fieldType = iota
	answerField
)

var fieldTypeToString = map[fieldType]string{
	questionField: "question",
	answerField:   "answer",
}

// NewModel creates a new Model from a Markdown file.
func NewModel(filename string, options ...Option) (*Model, error) {
	frontmatter, doc, err := newFrontmatterMarkdown(filename)
	if err != nil {
		return nil, err
	}
	answerType := frontmatter.AnswerType
	if answerType != "single-choice" && answerType != "multiple-choice" {
		return nil, fmt.Errorf("%w: unimplemented answerType %q", ErrInvalidFrontmatter, answerType)
	}
	model := &Model{
		Filename:    filename,
		Doc:         doc,
		Frontmatter: frontmatter,
		embeds:      map[markdown.Block]embed{},
	}
	for _, opt := range options {
		opt(model)
	}
	if err := model.buildModelForChoice(); err != nil {
		return nil, err
	}
	return model, nil
}

// Option is used on Model creation to set optional parameters.
type Option func(configurableModel)

type configurableModel interface {
	setPrivateKey(string)
}

// WithPrivateKey sets privateKey and all follow-up method invocations attempt
// to unseal sealed answers.
func WithPrivateKey(privateKey string) Option {
	return func(m configurableModel) {
		m.setPrivateKey(privateKey)
	}
}

// Seal seals the unsealed answer in the Frontmatter using the public key.
// Sealing can only be reverted if the secret private key is available.
// Answers committed to the public Repo should always be sealed.
func (m *Model) Seal(publicKey string) error {
	if err := m.Frontmatter.Seal(publicKey); err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename)
	}
	return nil
}

// Unseal unseals the sealed answer in the Frontmatter using the private key.
func (m *Model) Unseal() error {
	if err := m.Frontmatter.Unseal(m.privateKey); err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename)
	}
	return nil
}

// Verify checks if the given answer, provided by the frontmatter, is correct.
// Verify compares the given Frontmatter answer against generated output of
// questions and answers markdown code blocks and images.
func (m *Model) Verify() error {
	if !m.withSealed && m.Frontmatter.SealedAnswer != "" {
		return nil
	}
	_, err := m.getVerifiedAnswer()
	return err
}

// ExportAnswerKey returns the answerKey for the question Markdown file.
func (m *Model) ExportAnswerKey() (AnswerKey, error) {
	if !m.withSealed && m.Frontmatter.SealedAnswer != "" {
		return AnswerKey{}, nil // ignore
	}

	answer, err := m.getVerifiedAnswer()
	if err != nil {
		return nil, err
	}
	return NewAnswerKey(m.Filename, answer)
}

// ExportAnswerKeyJSON returns the answerKey for the question Markdown file as
// JSON string.
func (m *Model) ExportAnswerKeyJSON() (string, error) {
	answerKey, err := m.ExportAnswerKey()
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(answerKey, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

// WriteFormatted formats YAML frontmatter, fenced by "---", followed by
// formatted markdown content and writes it to original File.
func (m *Model) WriteFormatted() error {
	b, err := format(m.Frontmatter, m.Doc)
	if err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename)
	}
	return os.WriteFile(m.Filename, b, 0o666)
}

// PrintHTML prints the question and answer choices as HTML form elements.
func (m *Model) PrintHTML(buf *bytes.Buffer) {
	buf.WriteString("<form id=" + baseFilename(m.Filename) + ">\n")
	for _, block := range m.Doc.Blocks {
		if block == m.answerList {
			m.printAnswerChoicesHTML(block.(*markdown.List), buf)
			continue
		}
		if embed, ok := m.embeds[block]; ok {
			embed.renderer.RenderHTML(buf)
			continue
		}
		block.PrintHTML(buf)
	}
	buf.WriteString("</form>\n")
}

// ToHTML returns a complete standalone HTML document as string.
func (m *Model) ToHTML() string {
	buf := &bytes.Buffer{}
	buf.WriteString(questionPrefixHTML)
	m.PrintHTML(buf)
	buf.WriteString(questionSuffixHTML)
	return buf.String()
}

func baseFilename(filename string) string {
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}

func (m *Model) printAnswerChoicesHTML(list *markdown.List, buf *bytes.Buffer) {
	buf.WriteString("<fieldset>\n")
	for i, item := range list.Items {
		letter := indexToLetter(i)
		buf.WriteString("<div>\n")
		buf.WriteString(`<label for="` + letter + `">` + letter + "</label>\n")
		buf.WriteString(`<input type="radio" id="` + letter + `" name="answer" />` + "\n")
		for _, block := range item.(*markdown.Item).Blocks {
			if embed, ok := m.embeds[block]; ok {
				embed.renderer.RenderHTML(buf)
			} else {
				block.PrintHTML(buf)
			}
		}
		buf.WriteString("</div>\n")
	}
	buf.WriteString("</fieldset>\n")
}

func (m *Model) getVerifiedAnswer() (Answer, error) {
	answer, err := m.Frontmatter.getAnswer(m.privateKey)
	if err != nil {
		return Answer{}, err
	}
	correctByIndex := answer.correctAnswerIndices()
	generated := m.Question.RenderOutput()
	for i, choice := range m.AnswerChoices {
		choiceGen := choice.RenderOutput()
		if correctByIndex[i] && generated != choiceGen {
			return Answer{}, fmt.Errorf("%w: answer %q does not match question: %q != %q", ErrWrongAnswer, indexToLetter(i), strings.TrimSuffix(choiceGen, "\n"), strings.TrimSuffix(generated, "\n"))
		}
		if !correctByIndex[i] && generated == choiceGen {
			return Answer{}, fmt.Errorf("%w: expected %q: answer %q matches question: %q == %q", ErrWrongAnswer, answer.correctAnswers(), indexToLetter(i), strings.TrimSuffix(choiceGen, "\n"), strings.TrimSuffix(generated, "\n"))
		}
	}
	return answer, nil
}

// buildModelForChoice operates on single-choice or multiple-choice question documents
// and expects the following structure:
//
// - question: top level evyContent or top level paragraph with image ending in .evy.svg.
// - answers: list of codeblocks / inline-code, paragraphs with images or links.
func (m *Model) buildModelForChoice() error {
	for _, b := range m.Doc.Blocks {
		if err := m.buildQuestionField(b); err != nil {
			return err
		}
		if err := m.buildAnswerChoicesField(b); err != nil {
			return err
		}
	}
	if m.Question == nil {
		return fmt.Errorf("%w: found no question", ErrBadMarkdownStructure)
	}
	if len(m.AnswerChoices) < 2 {
		return fmt.Errorf("%w: found %d answer, expected 2 or more", ErrBadMarkdownStructure, len(m.AnswerChoices))
	}
	if err := m.inferResultType(); err != nil {
		return err
	}
	return nil
}

func (m *Model) buildQuestionField(b markdown.Block) error {
	if m.Question != nil {
		return nil
	}
	var err error
	m.Question, err = NewRenderer(b, questionField, m.Filename)
	if err != nil {
		return err
	}
	m.trackBlocksToReplace(b, m.Question)
	return nil
}

func (m *Model) buildAnswerChoicesField(block markdown.Block) error {
	list, ok := block.(*markdown.List)
	if m.Question == nil || m.AnswerChoices != nil || !ok {
		return nil
	}
	for _, item := range list.Items {
		found := false
		for _, b := range item.(*markdown.Item).Blocks {
			renderer, err := NewRenderer(b, answerField, m.Filename)
			if err != nil {
				return err
			}
			if renderer == nil {
				continue
			}
			if found {
				return fmt.Errorf("%w: found second answer in one list item", ErrBadMarkdownStructure)
			}
			found = true
			m.AnswerChoices = append(m.AnswerChoices, renderer)
			m.answerList = block
			m.trackBlocksToReplace(b, renderer)
		}
	}
	if len(m.AnswerChoices) != 0 && len(m.AnswerChoices) != len(list.Items) {
		return fmt.Errorf("%w: found %d answers, expected %d (one per list item)", ErrBadMarkdownStructure, len(m.AnswerChoices), len(list.Items))
	}
	return nil
}

func (m *Model) trackBlocksToReplace(b markdown.Block, renderer Renderer) {
	text := toText(b)
	if renderer == nil || text == nil || len(text.Inline) != 1 {
		return
	}
	id := idFromInline(text.Inline[0])
	if id == "" {
		return
	}
	m.embeds[b] = embed{id: id, renderer: renderer}
}

func idFromInline(inline markdown.Inline) string {
	switch i := inline.(type) {
	case *markdown.Link:
		return escape(i.URL)
	case *markdown.Image:
		return escape(i.URL)
	}
	return ""
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "/", "-")
	return strings.ReplaceAll(s, ".", "-")
}

func (m *Model) inferResultType() error {
	resultType := UnknownOutput
	renderers := append([]Renderer{m.Question}, m.AnswerChoices...)
	for _, r := range renderers {
		t := resultTypeFromRenderer(r)
		if resultType == UnknownOutput {
			resultType = t
		} else if resultType != t && t != UnknownOutput {
			return fmt.Errorf("%w: found text and image output, expected only one", ErrInconsistentMdoel)
		}
	}
	if resultType == UnknownOutput {
		return fmt.Errorf("%w: found neither text nor image output", ErrInconsistentMdoel)
	}
	m.ResultType = resultType
	for _, r := range renderers {
		if source, ok := r.(*evySource); ok {
			source.ResultType = resultType
		}
	}
	return nil
}

func (m *Model) setPrivateKey(privateKey string) {
	m.privateKey = privateKey
	m.withSealed = true
}

const questionPrefixHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>evy · Question</title>
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>⚡️</text></svg>" />
    <style>
      body {
        padding: 8px 32px;
        margin: 0;
      }
      fieldset {
        display: grid;
        grid-template: repeat(2, min-content) / repeat(2, min-content);
        grid-auto-flow: column;
        column-gap: 32px;
        row-gap: 8px;
        border: none;
      }
      pre {
        border: 1px solid silver;
        padding: 8px 12px;
        border-radius: 2px;
        background: whitesmoke;
        width: fit-content;
      }
      fieldset > div {
        display: flex;
        align-items: start;
        gap: 8px;
        border: 1px solid silver;
        border-radius: 6px;
        padding: 8px;
        background: whitesmoke;
      }
      fieldset pre {
        margin: 0;
        padding: 4px 12px;
        align-self: center;
        background: white;
        border: 1px solid silver
      }
      form svg {
        width: 200px;
        height: 200px;
        border: 1px solid silver;
      }
    </style>
  </head>
  <body>
  `

const questionSuffixHTML = `  </body>
</html>
`
