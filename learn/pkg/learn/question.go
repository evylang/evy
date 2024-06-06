package learn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rsc.io/markdown"
)

// QuestionModel represents a question and its answer choices as parsed and derived
// from the original Markdown file with frontmatter.
//
// In a QuestionModel the Question may be Evy source code, text output or image
// output. The QuestionModel AnswerChoices field is a list of Evy source code, text
// output and image output. In a verified QuestionModel, only the correct answer
// choice (output) matches the question (output). "Correct" means as
// specified in the Markdown Frontmatter.
type QuestionModel struct {
	Filename    string
	Doc         *markdown.Document
	Frontmatter *frontmatter

	Question      Renderer
	AnswerChoices []Renderer
	ResultType    ResultType

	ignoreSealed bool
	privateKey   string
	embeds       map[markdown.Block]embed // use to replace markdown Link or Image with codeBlock or inline SVG
	answerList   markdown.Block           // use to output checkbox or radio buttons for List in HTML
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

// NewQuestionModel creates a new QuestionModel from a Markdown file.
func NewQuestionModel(filename string, options ...Option) (*QuestionModel, error) {
	frontmatter, doc, err := newFrontmatterMarkdown(filename)
	if err != nil {
		return nil, err
	}
	answerType := frontmatter.AnswerType
	if answerType != "single-choice" && answerType != "multiple-choice" {
		return nil, fmt.Errorf("%w: unimplemented answerType %q", ErrInvalidFrontmatter, answerType)
	}
	model := &QuestionModel{
		Filename:    filename,
		Doc:         doc,
		Frontmatter: frontmatter,
		embeds:      map[markdown.Block]embed{},
	}
	for _, opt := range options {
		opt(model)
	}
	if err := model.buildQuestionModelForChoice(); err != nil {
		return nil, err
	}
	return model, nil
}

// Seal seals the unsealed answer in the Frontmatter using the public key.
// Sealing can only be reverted if the secret private key is available.
// Answers committed to the public Repo should always be sealed.
func (m *QuestionModel) Seal(publicKey string) error {
	if err := m.Frontmatter.Seal(publicKey); err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename)
	}
	return nil
}

// Unseal unseals the sealed answer in the Frontmatter using the private key.
func (m *QuestionModel) Unseal() error {
	if err := m.Frontmatter.Unseal(m.privateKey); err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename)
	}
	return nil
}

// Verify checks if the given answer, provided by the frontmatter, is correct.
// Verify compares the given Frontmatter answer against generated output of
// questions and answers markdown code blocks and images.
func (m *QuestionModel) Verify() error {
	if m.ignoreSealed && m.IsSealed() {
		return nil
	}
	_, err := m.getVerifiedAnswer()
	return err
}

// ExportAnswerKey returns the answerKey for the question Markdown file.
func (m *QuestionModel) ExportAnswerKey() (AnswerKey, error) {
	if m.ignoreSealed && m.IsSealed() {
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
func (m *QuestionModel) ExportAnswerKeyJSON() (string, error) {
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

// IsSealed returns true if the answer is sealed in the Frontmatter.
func (m *QuestionModel) IsSealed() bool {
	return m.Frontmatter.SealedAnswer != ""
}

// WriteFormatted formats YAML frontmatter, fenced by "---", followed by
// formatted markdown content and writes it to original File.
func (m *QuestionModel) WriteFormatted() error {
	b, err := format(m.Frontmatter, m.Doc)
	if err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename)
	}
	return os.WriteFile(m.Filename, b, 0o666)
}

// PrintHTML prints the question and answer choices as HTML form elements.
func (m *QuestionModel) PrintHTML(buf *bytes.Buffer, withMarked bool) error {
	buf.WriteString("<form id=" + baseFilename(m.Filename) + ` class="difficulty-` + string(m.Frontmatter.Difficulty) + `">` + "\n")
	for _, block := range m.Doc.Blocks {
		if block == m.answerList {
			if err := m.printAnswerChoicesHTML(block.(*markdown.List), buf, withMarked); err != nil {
				return err
			}
			continue
		}
		if embed, ok := m.embeds[block]; ok {
			embed.renderer.RenderHTML(buf)
			continue
		}
		block.PrintHTML(buf)
	}
	buf.WriteString("</form>\n")
	return nil
}

// ToHTML returns a complete standalone HTML document as string.
func (m *QuestionModel) ToHTML(withMarked bool) (string, error) {
	buf := &bytes.Buffer{}
	buf.WriteString(questionPrefixHTML)
	if err := m.PrintHTML(buf, withMarked); err != nil {
		return "", err
	}
	buf.WriteString(questionSuffixHTML)
	return buf.String(), nil
}

func baseFilename(filename string) string {
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}

func (m *QuestionModel) printAnswerChoicesHTML(list *markdown.List, buf *bytes.Buffer, withMarked bool) error {
	buf.WriteString("<fieldset>\n")
	inputType := "radio"
	if m.Frontmatter.AnswerType == "multiple-choice" {
		inputType = "checkbox"
	}
	var correctAnswers map[int]bool
	if withMarked && !(m.IsSealed() && m.ignoreSealed) {
		answer, err := m.Frontmatter.getAnswer(m.privateKey)
		if err != nil {
			return err
		}
		correctAnswers = answer.correctAnswerIndices()
	}
	for i, item := range list.Items {
		checked := ""
		if correctAnswers[i] {
			checked = "checked "
		}
		letter := indexToLetter(i)
		buf.WriteString("<div>\n")
		buf.WriteString(`<label for="` + letter + `">` + letter + "</label>\n")
		buf.WriteString(`<input type="` + inputType + `" value="` + letter + `" name="answer" ` + checked + "/>\n")
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
	return nil
}

func (m *QuestionModel) getVerifiedAnswer() (Answer, error) {
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

// buildQuestionModelForChoice operates on single-choice or multiple-choice question documents
// and expects the following structure:
//
// - question: top level evyContent or top level paragraph with image ending in .evy.svg.
// - answers: list of codeblocks / inline-code, paragraphs with images or links.
func (m *QuestionModel) buildQuestionModelForChoice() error {
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

func (m *QuestionModel) buildQuestionField(b markdown.Block) error {
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

func (m *QuestionModel) buildAnswerChoicesField(block markdown.Block) error {
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

func (m *QuestionModel) trackBlocksToReplace(b markdown.Block, renderer Renderer) {
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

func (m *QuestionModel) inferResultType() error {
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

func (m *QuestionModel) setPrivateKey(privateKey string) {
	m.privateKey = privateKey
}

func (m *QuestionModel) setIgnoreSealed() {
	m.ignoreSealed = true
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
      form {
        padding-bottom: 32px;
      }
      form svg {
        width: 200px;
        height: 200px;
        border: 1px solid silver;
      }
      form.difficulty-easy:after {
        content: "Difficulty: 🌶️";
      }
      form.difficulty-medium:after {
        content: "Difficulty: 🌶️🌶️";
      }
      form.difficulty-hard:after {
        content: "Difficulty: 🌶️🌶️🌶️";
      }
      form:not(:first-child) {
        border-top: 1px solid silver;
      }
      table {
        border-collapse: collapse;
        margin-bottom: 24px;
      }
      td {
        border: 1px solid silver;
        padding: 4px 8px;
      }
    </style>
  </head>
  <body>
  `

const questionSuffixHTML = `  </body>
</html>
`
