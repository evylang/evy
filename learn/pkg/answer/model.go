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

// Renderer generates output. It renders either plain text or image/SVG
// content.
//
// The rendered output is used to verify questions against answers according
// to the answerkey provided in the frontmatter.
//
// - Markdown Codeblocks can render text output directly
// - Markdown Codeblocks can contain Evy source code that renders text or image/SVG output
// - Markdown images can link to and render image/SVG output.
// - Markdown links can point to Evy source code which can be executed to output text or image/SVG
type Renderer interface {
	Render() string
}

type Model struct {
	Question      Renderer
	AnswerChoices []Renderer
	RenderType    RenderType
	Filename      string

	// track answer Block so we can render it with a/b/c/d checkboxes or radiobuttons
	mdAnswerList markdown.Block
	// Track to URLs to .evy files and images .evy.svg of so that their
	// contents or in the case of evy files the execution output (text or
	// SVG) can be embedded in the HTML output. Prefix with output type:
	// "svg:dot/dot.a.evy.svg", "source:dot/dot.a.evy", "svg:dot/dot.a.evy", "text:print/print.a.evy"
	renderers map[string]Renderer
}

type RenderType uint

const (
	UnknownOutput RenderType = iota
	TextOutput
	SVGOutput
)

var RenderTypeToString = map[RenderType]string{
	UnknownOutput: "invalid",
	TextOutput:    "text",
	SVGOutput:     "svg",
}

type FieldType uint

const (
	QuestionField FieldType = iota
	AnswerField
)

var FieldTypeToString = map[FieldType]string{
	QuestionField: "question",
	AnswerField:   "answer",
}

func NewModel(qmd *QuestionMarkdown) (*Model, error) {
	answerType := qmd.Frontmatter.AnswerType
	if answerType != "single-choice" && answerType != "multiple-choice" {
		return nil, fmt.Errorf("%w: unsupported answerType %q", ErrUnimplemented, answerType)
	}
	return newModelFromChoice(qmd)
}

// newModelFromChoice operates on single-choice, multiple-choice question documents
// and expects the following structure:
//
// - question: top level evyContent or top level paragraph with image ending in .evy.svg.
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
	if err := model.inferRenderType(); err != nil {
		return nil, err
	}
	return model, nil
}

func (m *Model) buildFields(b markdown.Block) error {
	var err error
	if m.Question == nil {
		m.Question, err = NewRenderer(b, QuestionField, m.Filename)
		if err != nil {
			return err
		}
	}
	if m.Question != nil && m.AnswerChoices == nil {
		m.AnswerChoices, err = NewRenderersFromList(b, AnswerField, m.Filename)
		if err != nil {
			return err
		}
		m.mdAnswerList = b
	}
	return nil
}

// NewRenderer converts a markdown code block, image or link into an
// Renderer.
//
// Markdown code blocks with an `evy` info tag are interpreted as Evy source
// code that can generate text or image/SVG output. Code blocks with an empty
// info tag are interpreted directly as text output.
//
// Markdown images need to have a relative URL ending in ".evy.svg". The image
// text has to be `question` or `answer`.
//
// Markdown links need to be to relative links to `.evy` files. The link text
// needs to be `question` or `answer`.The The link title indicates whether to
// use evy file as source file, generate text or image/SVG output from it. It
// needs to be one of:
//
//	evy:source | evy:svg | evy:text
//
// Code blocks, links and images that don't meet the conditions listed above
// are ignored. In this case NewRenderer returns nil, nil.
func NewRenderer(block markdown.Block, field FieldType, filename string) (Renderer, error) {
	if cb, ok := block.(*markdown.CodeBlock); ok {
		return rendererFromCodeblock(cb)
	}
	if text := toText(block); text != nil {
		return rendererFromInlines(text.Inline, field, filename)
	}
	return nil, nil
}

func rendererFromCodeblock(cb *markdown.CodeBlock) (Renderer, error) {
	if cb.Info == "evy" {
		text := strings.Join(cb.Text, "\n") + "\n"
		return newEvySource(text), nil
	}
	if cb.Info == "" {
		return newTextContent(cb), nil
	}
	return nil, nil
}

func rendererFromInlines(inlines []markdown.Inline, field FieldType, filename string) (Renderer, error) {
	var renderer Renderer
	for _, inline := range inlines {
		r, err := rendererFromInline(inline, field, filename)
		if err != nil {
			return nil, err
		}
		if r == nil {
			continue
		}
		if renderer != nil {
			return nil, fmt.Errorf("%w: found second output Renderer in one markdown text block", ErrBadMarkdownStructure)
		}
		renderer = r
	}
	if renderer == nil && len(inlines) == 1 {
		return rendererFromSingleInlineCode(inlines[0])
	}
	return renderer, nil
}

func rendererFromSingleInlineCode(inline markdown.Inline) (Renderer, error) {
	inlineCode, ok := inline.(*markdown.Code)
	if !ok {
		return nil, nil
	}
	return TextContent(inlineCode.Text + "\n"), nil
}

func rendererFromInline(inline markdown.Inline, field FieldType, filename string) (Renderer, error) {
	mdImg, ok := inline.(*markdown.Image)
	if ok && strings.HasSuffix(mdImg.URL, ".evy.svg") {
		if err := ensureField(mdImg.Inner, field); err != nil {
			return nil, fmt.Errorf("%w (image: %s)", err, inlineToString(mdImg))
		}
		return newSVGContentFromFile(filepath.Join(filepath.Dir(filename), mdImg.URL))
	}
	link, ok := inline.(*markdown.Link)
	if ok && strings.HasPrefix(link.Title, "evy:") {
		if err := ensureField(link.Inner, field); err != nil {
			return nil, fmt.Errorf("%w (link: %s)", err, inlineToString(link))
		}
		renderType, err := getRenderTypeFromLink(link)
		if err != nil {
			return nil, fmt.Errorf("%w (link: %s)", err, inlineToString(link))
		}
		filename := filepath.Join(filepath.Dir(filename), link.URL)
		return NewRendererFromEvyFile(filename, renderType)
	}
	return nil, nil
}

func ensureField(inner []markdown.Inline, field FieldType) error {
	if len(inner) != 1 {
		return fmt.Errorf("%w: found %d inner elements, expected 1", ErrBadMarkdownStructure, len(inner))
	}
	got := inlineToString(inner[0])
	want := FieldTypeToString[field]
	if got != want {
		return fmt.Errorf("%w: expected %q text, found %q", ErrBadMarkdownStructure, want, got)
	}
	return nil
}

func inlineToString(inline markdown.Inline) string {
	b := &bytes.Buffer{}
	inline.PrintText(b)
	return b.String()
}

// NewRenderersFromList converts a markdown block into a list of Renderers if
// the block is markdown list and all list items can be converted to
// Renderers.
func NewRenderersFromList(b markdown.Block, field FieldType, filename string) ([]Renderer, error) {
	list, ok := b.(*markdown.List)
	if !ok {
		return nil, nil
	}
	var Renderers []Renderer
	for _, item := range list.Items {
		foundRendererInItem := false
		for _, block := range item.(*markdown.Item).Blocks {
			Renderer, err := NewRenderer(block, field, filename)
			if err != nil {
				return nil, err
			}
			if Renderer != nil && foundRendererInItem {
				return nil, fmt.Errorf("%w: found second Renderer in one list item", ErrBadMarkdownStructure)
			}
			if Renderer != nil {
				foundRendererInItem = true
				Renderers = append(Renderers, Renderer)
			}
		}
	}
	if len(Renderers) != 0 && len(Renderers) != len(list.Items) {
		return nil, fmt.Errorf("%w: found %d output rendering items, expected %d (one per list item)", ErrBadMarkdownStructure, len(Renderers), len(list.Items))
	}
	return Renderers, nil
}

// Verify checks if the given answer, provided by answerkey or frontmatter, is
// correct. Verify compares the given answer against generated output of
// questions and answers markdown code blocks and images.
func (m *Model) Verify(answer Answer) error {
	if answer.Type != "single-choice" && answer.Type != "multiple-choice" {
		return fmt.Errorf("%w: unsupported answerType %q", ErrUnimplemented, answer.Type)
	}
	correctByIndex := answer.correctAnswerIndices()
	generated := m.Question.Render()
	for i, choice := range m.AnswerChoices {
		choiceGen := choice.Render()
		if correctByIndex[i] && generated != choiceGen {
			return fmt.Errorf("%w: expected answers %q: answer %q does not match question: %q != %q", ErrInvalidAnswer, answer.correctAnswers(), indexToLetter(i), choiceGen, generated)
		}
		if !correctByIndex[i] && generated == choiceGen {
			return fmt.Errorf("%w: expected answer %q: answer %q matches question: %q == %q", ErrInvalidAnswer, answer.correctAnswers(), indexToLetter(i), choiceGen, generated)
		}
	}
	return nil
}

type SVGContent string

func newSVGContentFromFile(filename string) (SVGContent, error) {
	b, err := os.ReadFile(filename)
	if err == nil {
		return SVGContent(b), nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	evySourcePath := strings.TrimSuffix(filename, ".svg")
	b, err2 := os.ReadFile(string(evySourcePath))
	if err2 != nil {
		return "", fmt.Errorf("error reading evy source for image: %w: %w (%s)", err, err2, filename)
	}
	svgData := runEvy(string(b), SVGOutput)
	if err2 := os.WriteFile(filename, []byte(svgData), 0666); err2 != nil {
		return "", fmt.Errorf("error writing svg output for evy source: %w: %w (%s)", err, err2, filename)
	}
	return SVGContent(svgData), nil
}

func (s SVGContent) Render() string {
	return string(s)
}

type evySource struct {
	source   string
	out      RenderType
	output   string // cached output
	filename string
}

func newEvySource(text string) *evySource {
	return &evySource{source: text}
}

func (s *evySource) Render() string {
	if s.output == "" {
		s.output = runEvy(s.source, s.out)
	}
	return s.output
}

type TextContent string

func newTextContent(cb *markdown.CodeBlock) TextContent {
	text := strings.Join(cb.Text, "\n") + "\n"
	return TextContent(text)
}

func (s TextContent) Render() string {
	return string(s)
}

func NewRendererFromEvyFile(filename string, renderType RenderType) (Renderer, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	switch renderType {
	case SVGOutput:
		content := runEvy(string(b), SVGOutput)
		return SVGContent(content), nil
	case TextOutput:
		return TextContent(runEvy(string(b), TextOutput)), nil
	}
	return newEvySource(string(b)), nil
}

func getRenderTypeFromLink(link *markdown.Link) (RenderType, error) {
	u := link.URL
	if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
		return UnknownOutput, fmt.Errorf("%w: found external link %q, expected relative link", ErrBadMarkdownStructure, u)
	}
	if !strings.HasSuffix(u, ".evy") {
		return UnknownOutput, fmt.Errorf("%w: found non evy link %q, expected .evy file", ErrBadMarkdownStructure, u)
	}
	if !strings.HasPrefix(link.Title, "evy:") {
		return UnknownOutput, fmt.Errorf("%w: found no evy title %q, expected evy:source | evy:svg | evy:text", ErrBadMarkdownStructure, link.Title)
	}
	switch strings.TrimPrefix(link.Title, "evy:") {
	case "text":
		return TextOutput, nil
	case "source":
		return UnknownOutput, nil
	case "svg":
		return SVGOutput, nil
	}
	return UnknownOutput, fmt.Errorf("%w: found invalid evy title %q, expected evy:source | evy:svg | evy:text", ErrBadMarkdownStructure, link.Title)
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

func (m *Model) inferRenderType() error {
	renderType := UnknownOutput
	renderers := append([]Renderer{m.Question}, m.AnswerChoices...)
	for _, r := range renderers {
		t := renderTypeFromRenderer(r)
		if renderType == UnknownOutput {
			renderType = t
		} else if renderType != t && t != UnknownOutput {
			return fmt.Errorf("%w: found text and image output, expected only one", ErrInconsistentMdoel)
		}
	}
	if renderType == UnknownOutput {
		return fmt.Errorf("%w: found neither text nor image output", ErrInconsistentMdoel)
	}
	m.RenderType = renderType
	for _, r := range renderers {
		if source, ok := r.(*evySource); ok {
			source.out = renderType
		}
	}
	return nil
}

func renderTypeFromRenderer(r Renderer) RenderType {
	switch r.(type) {
	case SVGContent:
		return SVGOutput
	case TextContent:
		return TextOutput
	}
	return UnknownOutput
}

func runEvy(source string, t RenderType) string {
	textWriter := &bytes.Buffer{}
	opts := []cli.Option{
		cli.WithSkipSleep(true),
		cli.WithOutputWriter(textWriter),
		cli.WithSVG("" /* root style */),
	}
	rt := cli.NewRuntime(opts...)
	eval := evaluator.NewEvaluator(rt)
	err := eval.Run(source)
	if err != nil {
		return "**ERROR**"
	}
	if t == SVGOutput {
		imgWriter := &bytes.Buffer{}
		rt.WriteSVG(imgWriter)
		return imgWriter.String()
	}
	return textWriter.String()
}
