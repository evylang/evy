package learn

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strconv"
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

	Question   Renderer
	ResultType ResultType

	// AnswerType determines model of choice vs text
	AnswerChoices []Renderer
	AnswerText    Renderer

	embeds      map[markdown.Block]Renderer // use to replace markdown Link or Image with codeBlock or inline SVG
	answerBlock markdown.Block              // use to output checkbox, radio buttons for choice questions, textarea for text questions

	subQuestions   []*QuestionModel // Derived questions, generated from txtar question and links.
	parentQuestion *QuestionModel   // Initial Model from which sub-question was generated.
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
		embeds:            map[markdown.Block]Renderer{},
		configurableModel: newConfigurableModel(filename, options),
	}
	if err := question.parseFrontmatterMD(); err != nil {
		return nil, fmt.Errorf("%w (%s)", err, question.Filename())
	}
	if err := question.buildQuestionModel(); err != nil {
		return nil, fmt.Errorf("%w (%s)", err, question.Filename())
	}
	subs, err := newSubQuestions(question)
	if err != nil {
		return nil, err
	}
	question.subQuestions = subs
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
	if m.hasSubQuestions() {
		answerKey := AnswerKey{}
		for _, q := range m.subQuestions {
			ak, err := q.ExportAnswerKey()
			if err != nil {
				return AnswerKey{}, err
			}
			answerKey.merge(ak)
		}
		return answerKey, nil
	}
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
		embed, ok := m.embeds[block]
		var err error
		switch {
		case block == m.answerBlock && m.Frontmatter.AnswerType == "text":
			err = m.printTextAnswerHTML(buf, withAnswersMarked)
		case block == m.answerBlock && m.isSubQuestion():
			err = m.printTxtarAnswerChoicesHTML(buf, withAnswersMarked)
		case block == m.answerBlock && m.isParserErrorQuestion():
			err = m.printTxtarAnswerChoicesHTML(buf, withAnswersMarked)
		case block == m.answerBlock:
			err = m.printAnswerChoicesHTML(block.(*markdown.List), buf, withAnswersMarked)
		case ok: // question block (answers are covered in the cases above)
			embed.RenderHTML(buf)
		default:
			printHTML(block, buf)
		}
		if err != nil {
			return err
		}
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
	correctAnswers, err := m.correctAnswerIndices(withAnswersMarked)
	if err != nil {
		return err
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
				embed.RenderHTML(buf)
			} else {
				printHTML(block, buf)
			}
		}
		buf.WriteString("</div>\n")
	}
	buf.WriteString("</fieldset>\n")
	return nil
}

func (m *QuestionModel) printTxtarAnswerChoicesHTML(buf *bytes.Buffer, withAnswersMarked bool) error {
	buf.WriteString("<fieldset>\n")
	inputType := "radio"
	if m.Frontmatter.AnswerType == "multiple-choice" {
		inputType = "checkbox"
	}
	correctAnswers, err := m.correctAnswerIndices(withAnswersMarked)
	if err != nil {
		return err
	}
	txtarRenderer := m.AnswerChoices[0].(*txtarContent)
	files := txtarRenderer.archive.Files
	for i, file := range files {
		checked := ""
		if correctAnswers[i] {
			checked = "checked "
		}
		letter := indexToLetter(i)
		buf.WriteString("<div>\n")
		buf.WriteString(`<label for="` + letter + `">` + letter + "</label>\n")
		buf.WriteString(`<input type="` + inputType + `" value="` + letter + `" name="answer" ` + checked + "/>\n")
		if m.isParserErrorQuestion() {
			buf.WriteString(`<pre><code class="language-evy">`)
			buf.Write(file.Data)
			buf.WriteString("</code></pre>")
		} else {
			r := newRendererFromEvyBytes(file.Data, txtarRenderer.ResultType)
			r.RenderHTML(buf)
		}
		buf.WriteString("</div>\n")
	}
	buf.WriteString("</fieldset>\n")
	return nil
}

func (m *QuestionModel) printTextAnswerHTML(buf *bytes.Buffer, withAnswersMarked bool) error {
	content, err := m.answerContent(withAnswersMarked)
	if err != nil {
		return err
	}
	lineCount := strconv.Itoa(strings.Count(content, "\n") + 1)
	buf.WriteString(`<textarea name="answer" rows="` + lineCount + `" cols="35">` + "\n")
	buf.WriteString(content)
	buf.WriteString("</textarea>\n")
	return nil
}

func (m *QuestionModel) answerContent(withAnswersMarked bool) (string, error) {
	if !withAnswersMarked {
		if evySrc, ok := m.AnswerText.(*evySource); ok {
			return removeCommentTaggedLines(evySrc.source), nil
		}
		return m.AnswerText.RenderOutput(), nil
	}
	answer, err := m.Frontmatter.getAnswer(m.privateKey)
	if err != nil {
		return "", fmt.Errorf("%w: (%s)", err, m.Filename())
	}
	content := removeCommentTags(answer.Text)
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content, nil
}

func removeCommentTaggedLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.HasSuffix(line, " //levy:blank") {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}

func (m *QuestionModel) correctAnswerIndices(withAnswersMarked bool) (map[int]bool, error) {
	if !withAnswersMarked {
		return nil, nil
	}
	if m.IsSealed() && m.ignoreSealed {
		return nil, nil
	}
	answer, err := m.Frontmatter.getAnswer(m.privateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: (%s)", err, m.Filename())
	}
	return answer.correctAnswerIndices(), nil
}

func (m *QuestionModel) getVerifiedAnswer() (Answer, error) {
	answer, err := m.Frontmatter.getAnswer(m.privateKey)
	if err != nil {
		return Answer{}, err
	}
	verification := m.Frontmatter.Verification
	if m.isSubQuestion() || verification == "none" {
		return answer, nil
	}
	if m.isMatchQuestion() {
		if err := m.verifyMatch(answer); err != nil {
			return Answer{}, err
		}
	}
	if verification == "parse-error" {
		if err := m.verifyParseError(answer); err != nil {
			return Answer{}, err
		}
	}
	if verification == "no-parse-error" {
		if err := m.verifyNoParseError(answer); err != nil {
			return Answer{}, err
		}
	}
	return answer, nil
}

func (m *QuestionModel) verifyMatch(answer Answer) error {
	switch m.Frontmatter.AnswerType {
	case "single-choice", "multiple-choice":
		return m.verifyChoiceMatch(answer)
	case "text":
		return m.verifyTextMatch(answer)
	}
	return fmt.Errorf("%w: verify match: unimplemented answer type %q", ErrInvalidFrontmatter, m.Frontmatter.AnswerType)
}

func (m *QuestionModel) verifyChoiceMatch(answer Answer) error {
	correctByIndex := answer.correctAnswerIndices()
	generated := m.Question.RenderOutput()
	outputs := generateAnserOutputs(m.AnswerChoices)
	for i, output := range outputs {
		if correctByIndex[i] && generated != output {
			return fmt.Errorf("%w (%s): answer %q does not match question: %q != %q", ErrWrongAnswer, m.Filename(), indexToLetter(i), strings.TrimSuffix(output, "\n"), strings.TrimSuffix(generated, "\n"))
		}
		if !correctByIndex[i] && generated == output {
			return fmt.Errorf("%w (%s): expected %q: answer %q matches question: %q == %q", ErrWrongAnswer, m.Filename(), answer.correctAnswers(), indexToLetter(i), strings.TrimSuffix(output, "\n"), strings.TrimSuffix(generated, "\n"))
		}
	}
	return nil
}

func (m *QuestionModel) verifyTextMatch(answer Answer) error {
	genFromQuestion := strings.TrimSpace(m.Question.RenderOutput())
	genFromAnswerKey := strings.TrimSpace(answer.Text)
	if _, ok := m.AnswerText.(*evySource); ok {
		genFromAnswerKey = strings.TrimSpace(runEvy(genFromAnswerKey, m.ResultType))
	}
	if genFromQuestion != genFromAnswerKey {
		return fmt.Errorf("%w (%s): want != got \nwant:\n%q\ngot:\n%q", ErrWrongAnswer, m.Filename(), genFromQuestion, genFromAnswerKey)
	}
	return nil
}

func (m *QuestionModel) verifyParseError(answer Answer) error {
	correctByIndex := answer.correctAnswerIndices()
	parseErrors := generateParseErrors(m.AnswerChoices[0].(*txtarContent))
	for i, parseError := range parseErrors {
		if correctByIndex[i] && !parseError {
			return fmt.Errorf("%w: %s: answer %s should have parse error", ErrWrongAnswer, m.Filename(), indexToLetter(i))
		}
		if !correctByIndex[i] && parseError {
			return fmt.Errorf("%w: %s: answer %s should not have parse error", ErrWrongAnswer, m.Filename(), indexToLetter(i))
		}
	}
	return nil
}

func (m *QuestionModel) verifyNoParseError(answer Answer) error {
	correctByIndex := answer.correctAnswerIndices()
	parseErrors := generateParseErrors(m.AnswerChoices[0].(*txtarContent))
	for i, parseError := range parseErrors {
		if correctByIndex[i] && parseError {
			return fmt.Errorf("%w: %s: answer %s should not have parse error", ErrWrongAnswer, m.Filename(), indexToLetter(i))
		}
		if !correctByIndex[i] && !parseError {
			return fmt.Errorf("%w: %s: answer %s should have parse error", ErrWrongAnswer, m.Filename(), indexToLetter(i))
		}
	}
	return nil
}

func generateAnserOutputs(renderers []Renderer) []string {
	if textar, ok := renderers[0].(*txtarContent); ok {
		files := textar.archive.Files
		outputs := make([]string, len(files))
		for i, file := range files {
			if !strings.HasSuffix(file.Name, ".evy") {
				outputs[i] = string(file.Data)
				continue
			}
			r := newRendererFromEvyBytes(file.Data, textar.ResultType)
			outputs[i] = r.RenderOutput()
		}
		return outputs
	}
	outputs := make([]string, len(renderers))
	for i, r := range renderers {
		outputs[i] = r.RenderOutput()
	}
	return outputs
}

func generateParseErrors(t *txtarContent) []bool {
	files := t.archive.Files
	parseErrors := make([]bool, len(files))
	for i, file := range files {
		parseErrors[i] = hasParseError(string(file.Data))
	}
	return parseErrors
}

func (m *QuestionModel) buildQuestionModel() error {
	switch m.Frontmatter.AnswerType {
	case "single-choice", "multiple-choice":
		return m.buildQuestionModelForChoice()
	case "text":
		return m.buildQuestionModelForText()
	}
	return fmt.Errorf("%w: parse question: unimplemented answer type %q", ErrInvalidFrontmatter, m.Frontmatter.AnswerType)
}

func (m *QuestionModel) buildQuestionModelForText() error {
	for _, b := range m.Doc.Blocks {
		if m.Question == nil {
			if err := m.buildQuestionField(b); err != nil {
				return err
			}
			continue
		}
		if m.AnswerText == nil {
			if err := m.buildAnswerTextField(b); err != nil {
				return err
			}
		}
	}
	if m.Question == nil {
		return fmt.Errorf("%w: found no question", ErrBadMarkdownStructure)
	}
	if m.AnswerText == nil {
		return fmt.Errorf("%w: found no text answer section", ErrBadMarkdownStructure)
	}
	if err := m.inferResultType(); err != nil {
		return err
	}
	return nil
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
		if m.isMatchQuestion() {
			return fmt.Errorf("%w: found no question", ErrBadMarkdownStructure)
		}
	}
	if len(m.AnswerChoices) == 0 {
		return fmt.Errorf("%w: found no answers", ErrBadMarkdownStructure)
	}
	if len(m.AnswerChoices) == 1 {
		if _, ok := m.AnswerChoices[0].(*txtarContent); !ok {
			return fmt.Errorf("%w: found only 1 answer", ErrBadMarkdownStructure)
		}
	}
	if m.Frontmatter.GenerateQuestions != "" || m.isParserErrorQuestion() {
		// Single .txtar question required for question generation and
		// (no) parse error verification.
		if len(m.AnswerChoices) != 1 {
			return fmt.Errorf("%w: found %d answer, expected exactly 1", ErrBadMarkdownStructure, len(m.AnswerChoices))
		}
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
	m.Question, err = AsRenderer(b, questionField, m.Filename())
	if err != nil {
		return err
	}
	if m.Question == nil {
		return nil
	}
	if m.Frontmatter.GenerateQuestions != "" {
		if _, ok := m.Question.(*txtarContent); !ok {
			return fmt.Errorf("%w: expected txtar question link, got %T", ErrBadMarkdownStructure, m.Question)
		}
	}
	m.trackBlocksToReplace(b, m.Question)
	return nil
}

func (m *QuestionModel) buildAnswerChoicesField(block markdown.Block) error {
	list, ok := block.(*markdown.List)
	if m.AnswerChoices != nil || !ok {
		return nil
	}
	if m.Question == nil && !m.isParserErrorQuestion() {
		return nil
	}
	for _, item := range list.Items {
		found := false
		for _, b := range item.(*markdown.Item).Blocks {
			renderer, err := AsRenderer(b, answerField, m.Filename())
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
			m.answerBlock = block
			m.trackBlocksToReplace(b, renderer)
		}
	}
	if len(m.AnswerChoices) != 0 && len(m.AnswerChoices) != len(list.Items) {
		return fmt.Errorf("%w: found %d answers, expected %d (one per list item)", ErrBadMarkdownStructure, len(m.AnswerChoices), len(list.Items))
	}
	if m.Frontmatter.GenerateQuestions != "" && len(m.AnswerChoices) != 1 {
		return fmt.Errorf("%w: expected exactly 1 answer for question generation, got %d", ErrBadMarkdownStructure, len(m.AnswerChoices))
	}
	return nil
}

func (m *QuestionModel) buildAnswerTextField(block markdown.Block) error {
	var err error
	m.AnswerText, err = AsRenderer(block, answerField, m.Filename())
	if err != nil {
		return err
	}
	if m.AnswerText == nil {
		return nil
	}
	m.answerBlock = block
	m.trackBlocksToReplace(block, m.Question)
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
	m.embeds[b] = renderer
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
	if m.AnswerText != nil {
		renderers = append(renderers, m.AnswerText)
	}
	for _, r := range renderers {
		t := resultTypeFromRenderer(r)
		if resultType == UnknownOutput {
			resultType = t
		} else if resultType != t && t != UnknownOutput {
			return fmt.Errorf("%w: found text and image output, expected only one", ErrInconsistentMdoel)
		}
	}
	if resultType == UnknownOutput && !m.isParserErrorQuestion() {
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
	answerType := m.Frontmatter.AnswerType
	if answerType != "single-choice" && answerType != "multiple-choice" && answerType != "text" {
		return fmt.Errorf("%w: unimplemented answerType %q", ErrInvalidFrontmatter, answerType)
	}
	parser := markdown.Parser{AutoLinkText: true, TaskListItems: true}
	m.Doc = parser.Parse(m.rawMD)
	return nil
}

func (m *QuestionModel) pick() *QuestionModel {
	if !m.hasSubQuestions() {
		return m
	}
	randIdx := rand.Intn(len(m.subQuestions)) //nolint:gosec // we are fine to use "insecure" randomization here.
	return m.subQuestions[randIdx]
}

func (m *QuestionModel) isSubQuestion() bool {
	return m.parentQuestion != nil
}

func (m *QuestionModel) hasSubQuestions() bool {
	return len(m.subQuestions) != 0
}

func (m *QuestionModel) isParserErrorQuestion() bool {
	return m.Frontmatter.Verification == "parse-error" || m.Frontmatter.Verification == "no-parse-error"
}

func (m *QuestionModel) isMatchQuestion() bool {
	return m.Frontmatter.Verification == "match" || m.Frontmatter.Verification == ""
}
