package learn

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"evylang.dev/evy/pkg/md"
	"gopkg.in/yaml.v3"
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
	*configurableModel
	Doc         *markdown.Document
	Frontmatter *questionFrontmatter

	Question      Renderer
	AnswerChoices []Renderer
	ResultType    ResultType

	embeds     map[markdown.Block]embedMD // use to replace markdown Link or Image with codeBlock or inline SVG
	answerList markdown.Block             // use to output checkbox or radio buttons for List in HTML
}

type embedMD struct {
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

// NewQuestionModel returns a new question model from a question Markdown file
// or its contents.
func NewQuestionModel(filename string, options ...Option) (*QuestionModel, error) {
	question := &QuestionModel{
		embeds:            map[markdown.Block]embedMD{},
		configurableModel: newConfigurableModel(filename, options),
	}
	if err := question.parseFrontmatterMD(); err != nil {
		return nil, err
	}
	if err := question.buildQuestionModelForChoice(); err != nil {
		return nil, fmt.Errorf("%w (%s)", err, question.Filename())
	}
	return question, nil
}

// Seal seals the unsealed answer in the Frontmatter using the public key.
// Sealing can only be reverted if the secret private key is available.
// Answers committed to the public Repo should always be sealed.
func (m *QuestionModel) Seal(publicKey string) error {
	if err := m.Frontmatter.Seal(publicKey); err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename())
	}
	return nil
}

// Unseal unseals the sealed answer in the Frontmatter using the private key.
func (m *QuestionModel) Unseal() error {
	if err := m.Frontmatter.Unseal(m.privateKey); err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename())
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
	return NewAnswerKey(m.Filename(), answer)
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
		return fmt.Errorf("%w (%s)", err, m.Filename())
	}
	return os.WriteFile(m.Filename(), b, 0o666)
}

// PrintHTML prints the question and answer choices as HTML form elements.
func (m *QuestionModel) PrintHTML(buf *bytes.Buffer, withAnswersMarked bool) error {
	md.Walk(m.Doc, md.RewriteLink)
	buf.WriteString("<form id=" + baseNoExt(m.Filename()) + ` class="difficulty-` + string(m.Frontmatter.Difficulty) + `">` + "\n")
	for _, block := range m.Doc.Blocks {
		if block == m.answerList {
			if err := m.printAnswerChoicesHTML(block.(*markdown.List), buf, withAnswersMarked); err != nil {
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
func (m *QuestionModel) ToHTML(withAnswersMarked bool) (string, error) {
	buf := &bytes.Buffer{}
	if err := m.PrintHTML(buf, withAnswersMarked); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Name returns the name of the question model as the base filename without extension.
func (m *QuestionModel) Name() string {
	return baseNoExt(m.Filename())
}

func (m *QuestionModel) printAnswerChoicesHTML(list *markdown.List, buf *bytes.Buffer, withAnswersMarked bool) error {
	buf.WriteString("<fieldset>\n")
	inputType := "radio"
	if m.Frontmatter.AnswerType == "multiple-choice" {
		inputType = "checkbox"
	}
	var correctAnswers map[int]bool
	if withAnswersMarked && !(m.IsSealed() && m.ignoreSealed) {
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
			return Answer{}, fmt.Errorf("%w: %s: answer %q does not match question: %q != %q", ErrWrongAnswer, m.Filename(), indexToLetter(i), strings.TrimSuffix(choiceGen, "\n"), strings.TrimSuffix(generated, "\n"))
		}
		if !correctByIndex[i] && generated == choiceGen {
			return Answer{}, fmt.Errorf("%w: %s: expected %q: answer %q matches question: %q == %q", ErrWrongAnswer, m.Filename(), answer.correctAnswers(), indexToLetter(i), strings.TrimSuffix(choiceGen, "\n"), strings.TrimSuffix(generated, "\n"))
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
	m.Question, err = NewRenderer(b, questionField, m.Filename())
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
			renderer, err := NewRenderer(b, answerField, m.Filename())
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
	m.embeds[b] = embedMD{id: id, renderer: renderer}
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

func (m *QuestionModel) parseFrontmatterMD() error {
	var err error
	if m.rawFrontmatter == "" && m.rawMD == "" {
		m.rawFrontmatter, m.rawMD, err = readSplitMDFile(m.Filename())
		if err != nil {
			return fmt.Errorf("%w (%s)", err, m.Filename())
		}
	}
	m.Frontmatter = &questionFrontmatter{}
	if err := yaml.Unmarshal([]byte(m.rawFrontmatter), m.Frontmatter); err != nil {
		return fmt.Errorf("%w: cannot process Markdown frontmatter: %w (%s)", ErrInvalidFrontmatter, err, m.Filename())
	}
	if err := m.Frontmatter.validate(); err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename())
	}
	if m.Frontmatter.AnswerType != "single-choice" && m.Frontmatter.AnswerType != "multiple-choice" {
		return fmt.Errorf("%w: unimplemented answerType %q", ErrInvalidFrontmatter, m.Frontmatter.AnswerType)
	}
	parser := markdown.Parser{AutoLinkText: true, TaskListItems: true}
	m.Doc = parser.Parse(m.rawMD)
	return nil
}
