package answer

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"evylang.dev/evy/pkg/cli"
	"evylang.dev/evy/pkg/evaluator"
	"rsc.io/markdown"
)

var (
	ErrBadMarkdownStructure = fmt.Errorf("bad Markdown structure")
	ErrInconsistentMdoel    = fmt.Errorf("inconsistency")
	ErrInvalidAnswer        = fmt.Errorf("invalid answer")
)

type Model struct {
	Question      outputGenerator
	AnswerChoices []outputGenerator
	OutputType    OutputType
	Filename      string
}

func NewModel(qmd *QuestionMarkdown, answerType AnswerType) (*Model, error) {
	if answerType != "single-choice" && answerType != "multiple-choice" {
		return nil, fmt.Errorf("%w: unsupported answerType %q", ErrUnimplemented, answerType)
	}
	return newModelFromChoice(qmd)
}

// newModelFromChoice operates on single-choice, multiple-choice question documents
// and expects the following structure:
//
// - question: top level codeblock or top level paragraph with image ending in .evy.svg.
// answers: list of codeblocks or paragraphs with images
func newModelFromChoice(qmd *QuestionMarkdown) (*Model, error) {
	var err error
	model := &Model{Filename: qmd.Filename}
	for _, b := range qmd.Doc.Blocks {
		if err = model.buildFields(b); err != nil {
			return nil, err
		}
	}
	if model.Question == nil {
		return nil, fmt.Errorf("%w: found no question", ErrBadMarkdownStructure)
	}
	if len(model.AnswerChoices) < 2 {
		return nil, fmt.Errorf("%w: found %d answer, expected 2 or more", ErrBadMarkdownStructure, len(model.AnswerChoices))
	}
	if err := model.inferOutputType(); err != nil {
		return nil, err
	}
	return model, nil
}

func (m *Model) buildFields(b markdown.Block) error {
	if err := m.buildQuestion(b); err != nil {
		return err
	}
	if m.Question != nil {
		return m.buildAnswerChoices(b)
	}
	return nil
}

func (m *Model) buildQuestion(b markdown.Block) error {
	if m.Question != nil {
		return nil
	}
	var err error
	m.Question, err = newOutputGeneratorFromBlock(b, m.Filename)
	return err
}

func (m *Model) buildAnswerChoices(b markdown.Block) error {
	if m.AnswerChoices != nil {
		return nil
	}
	var err error
	m.AnswerChoices, err = newOutputGeneratorsFromList(b, m.Filename)
	return err
}

func (m *Model) setAnswerChoices(answerChoices []outputGenerator) error {
	if m.Question == nil {
		return fmt.Errorf("%w: found answers before question", ErrBadMarkdownStructure)
	}
	if m.AnswerChoices != nil {
		return fmt.Errorf("%w: found a second set of answers", ErrBadMarkdownStructure)
	}
	m.AnswerChoices = answerChoices
	return nil
}

// newOutputGeneratorFromBlock converts a markdown block into an outputGenerator if the
// block is a code block with an `evy` or `evy-out` info tag. It also turn a
// markdown image with a URL ending in ".evy.svg" into an outputGnerator.
//
// newOutputGeneratorFromBlock only considers top-level code blocks and images inside
// paragraphs. The first matching is returned, all following are ignored.
func newOutputGeneratorFromBlock(b markdown.Block, filename string) (outputGenerator, error) {
	if cb, ok := b.(*markdown.CodeBlock); ok {
		if cb.Info == "evy" || cb.Info == "evy-out" || cb.Info == "" {
			return newCodeBlock(cb), nil
		}
	}
	text := toText(b)
	if text == nil {
		return nil, nil
	}
	for _, inline := range text.Inline {
		mdImg, ok := inline.(*markdown.Image)
		if ok && strings.HasSuffix(mdImg.URL, ".evy.svg") {
			return newImageFromFile(filepath.Join(filepath.Dir(filename), mdImg.URL))
		}
		link, ok := inline.(*markdown.Link)
		if ok && isRelativeEvySourceURL(link) {
			return newCodeBlockFromFile(filepath.Join(filepath.Dir(filename), link.URL))
		}
	}
	return nil, nil
}

func isRelativeEvySourceURL(link *markdown.Link) bool {
	u := link.URL
	if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
		return false // not a relative URL
	}
	if !strings.HasSuffix(u, ".evy") {
		return false
	}
	return link.Title == "evy:source"
}

func toText(b markdown.Block) *markdown.Text {
	if t, ok := b.(*markdown.Text); ok {
		return t
	}
	if p, ok := b.(*markdown.Paragraph); ok {
		return p.Text
	}
	return nil

}

// newOutputGeneratorsFromList converts a markdown block into a list of
// outputGenerators if the block is markdown list and all list items
// can be converted to outputGenerators.
func newOutputGeneratorsFromList(b markdown.Block, filename string) ([]outputGenerator, error) {
	list, ok := b.(*markdown.List)
	if !ok {
		return nil, nil
	}
	var outputGenerators []outputGenerator
	for _, item := range list.Items {
		foundGeneratorInItem := false
		for _, block := range item.(*markdown.Item).Blocks {
			outputGenerator, err := newOutputGeneratorFromBlock(block, filename)
			if err != nil {
				return nil, err
			}
			if outputGenerator != nil && foundGeneratorInItem {
				return nil, fmt.Errorf("%w: found second outputGenerator in one list item", ErrBadMarkdownStructure)
			}
			if outputGenerator != nil {
				foundGeneratorInItem = true
				outputGenerators = append(outputGenerators, outputGenerator)
			}
		}
	}
	if len(outputGenerators) != 0 && len(outputGenerators) != len(list.Items) {
		return nil, fmt.Errorf("%w: found %d output generating items, expected %d (one per list item)", ErrBadMarkdownStructure, len(outputGenerators), len(list.Items))
	}
	return outputGenerators, nil
}

// Verify checks if the given answer, provided by answerkey or frontmatter, is
// correct. Verify compares the given answer against generated output of
// questions and answers markdown code blocks and images.
func (m *Model) Verify(answer Answer) error {
	if answer.Type != "single-choice" && answer.Type != "multiple-choice" {
		return fmt.Errorf("%w: unsupported answerType %q", ErrUnimplemented, answer.Type)
	}
	correctByIndex := answer.correctAnswerIndices()
	generated := m.Question.genOutput(m.OutputType)
	for i, choice := range m.AnswerChoices {
		choiceGen := choice.genOutput(m.OutputType)
		if correctByIndex[i] && generated != choiceGen {
			return fmt.Errorf("%w: expected answers %q: answer %q does not match question: %q != %q", ErrInvalidAnswer, answer.correctAnswers(), indexToLetter(i), choiceGen, generated)
		}
		if !correctByIndex[i] && generated == choiceGen {
			return fmt.Errorf("%w: expected answer %q: answer %q matches question: %q == %q", ErrInvalidAnswer, answer.correctAnswers(), indexToLetter(i), choiceGen, generated)
		}
	}
	return nil
}

type OutputType uint

const (
	unknownOutput OutputType = iota
	textOutput
	imgOutput
)

var OutputTypeToString = map[OutputType]string{
	unknownOutput: "unknown",
	textOutput:    "text",
	imgOutput:     "img",
}

type outputGenerator interface {
	genOutput(OutputType) string // OutputType needed for generic evy programs that can have text _and_ image output
	outputType() OutputType
}

type image string

func newImageFromFile(filename string) (image, error) {
	b, err := os.ReadFile(filename)
	if err == nil {
		return image(b), nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	evySourcePath := strings.TrimSuffix(filename, ".svg")
	b, err2 := os.ReadFile(string(evySourcePath))
	if err2 != nil {
		return "", fmt.Errorf("error reading evy source for image: %w: %w (%s)", err, err2, filename)
	}
	svgData := runEvy(string(b), imgOutput)
	if err2 := os.WriteFile(filename, []byte(svgData), 0666); err2 != nil {
		return "", fmt.Errorf("error writing svg output for evy source: %w: %w (%s)", err, err2, filename)
	}
	return image(svgData), nil
}

func (img image) genOutput(t OutputType) string {
	return string(img)
}

func (img image) outputType() OutputType { return imgOutput }

type codeBlock struct {
	outType OutputType
	text    string
}

func newCodeBlock(cb *markdown.CodeBlock) *codeBlock {
	outType := unknownOutput
	if cb.Info == "evy-out" || cb.Info == "" {
		outType = textOutput
	}
	text := strings.Join(cb.Text, "\n") + "\n"
	return &codeBlock{outType: outType, text: text}
}

func newCodeBlockFromFile(filename string) (*codeBlock, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &codeBlock{outType: unknownOutput, text: string(b)}, nil
}

func (b *codeBlock) genOutput(t OutputType) string {
	switch {
	case b.outType == unknownOutput:
		return runEvy(b.text, t)
	case b.outType == textOutput && t == textOutput:
		return b.text
	}
	panic("invalid codeBlock")
}

func (b *codeBlock) outputType() OutputType {
	return b.outType
}

func (m *Model) inferOutputType() error {
	outputType := unknownOutput
	generators := append([]outputGenerator{m.Question}, m.AnswerChoices...)
	for _, g := range generators {
		t := g.outputType()
		if outputType == unknownOutput {
			outputType = t
		} else if outputType != t && t != unknownOutput {
			return fmt.Errorf("%w: found text and image output, expected only one", ErrInconsistentMdoel)
		}
	}
	if outputType == unknownOutput {
		return fmt.Errorf("%w: found neither text nor image output", ErrInconsistentMdoel)
	}
	m.OutputType = outputType
	return nil
}

func runEvy(source string, t OutputType) string {
	textWriter := &bytes.Buffer{}
	opts := []cli.Option{
		cli.WithSkipSleep(true),
		cli.WithOutputWriter(textWriter),
	}
	if t == imgOutput {
		opts = append(opts, cli.WithSVG("" /* root style */))
	}
	rt := cli.NewRuntime(opts...)
	eval := evaluator.NewEvaluator(rt)
	err := eval.Run(source)
	if err != nil {
		return "**ERROR**"
	}
	if t == imgOutput {
		imgWriter := &bytes.Buffer{}
		rt.WriteSVG(imgWriter)
		return imgWriter.String()
	}
	return textWriter.String()
}
