package answer

import (
	"fmt"
	"os"
	"strings"

	"rsc.io/markdown"
)

var (
	ErrBadMarkdownStructure = fmt.Errorf("bad Markdown structure")
)

type Model struct {
	Question      outputGenerator
	AnswerChoices []outputGenerator
	OutputType    OutputType
}

func NewModel(doc *markdown.Document, answerType AnswerType) (*Model, error) {
	if answerType != "single-choice" && answerType != "multiple-choice" {
		return nil, fmt.Errorf("%w: unsupported answerType %q", ErrUnimplemented, answerType)
	}
	return newFromChoice(doc)
}

// newFromChoice operates on single-choice, multiple-choice question documents
// and expects the following structure:
//
// - question: top level codeblock or top level paragraph with image ending in .evy.svg.
// answers: list of codeblocks or paragraphs with images
func newFromChoice(doc *markdown.Document) (*Model, error) {
	var question outputGenerator
	var answers []outputGenerator
	var err error
	for _, b := range doc.Blocks {
		question, answers, err = toModelFields(question, answers, b)
		if err != nil {
			return nil, err
		}
	}
	if question == nil {
		return nil, fmt.Errorf("%w: found no question", ErrBadMarkdownStructure)
	}
	if len(answers) < 2 {
		return nil, fmt.Errorf("%w: found %d questions, expected 1.", ErrBadMarkdownStructure, len(answers))
	}
	outputType, err := getOutputType(append([]outputGenerator{question}, answers...))
	if err != nil {
		return nil, err
	}
	return &Model{
		Question:      question,
		AnswerChoices: answers,
		OutputType:    outputType,
	}, nil
}

func toModelFields(question outputGenerator, answers []outputGenerator, b markdown.Block) (outputGenerator, []outputGenerator, error) {
	outputGenerator, err := toOutputGenerator(b)
	if err != nil {
		return nil, nil, err
	}
	if outputGenerator != nil {
		if question != nil {
			return nil, nil, fmt.Errorf("%w: found second question", ErrBadMarkdownStructure)
		}
		if answers != nil {
			return nil, nil, fmt.Errorf("%w: found question after answers", ErrBadMarkdownStructure)
		}
		return outputGenerator, nil, nil
	}
	outputGeneratorList, err := toOutputGeneratorList(b)
	if err != nil {
		return nil, nil, err
	}
	if outputGeneratorList != nil {
		if question == nil {
			return nil, nil, fmt.Errorf("%w: found answers before question", ErrBadMarkdownStructure)
		}
		if answers != nil {
			return nil, nil, fmt.Errorf("%w: found a second set of answers", ErrBadMarkdownStructure)
		}
		return question, outputGeneratorList, nil
	}
	return nil, nil, nil
}

// toOutputGenerator converts a markdown block into an outputGenerator if the
// block is a code block with an `evy` or `evy-out` info tag. It also turn a
// markdown image with a URL ending in ".evy.svg" into an outputGnerator.
//
// toOutputGenerator only considers top-level code blocks and images inside
// paragraphs. The first matching is returned, all following are ignored.
func toOutputGenerator(b markdown.Block) (outputGenerator, error) {
	if cb, ok := b.(*markdown.CodeBlock); ok {
		if cb.Info == "evy" || cb.Info == "evy-out" {
			return &codeBlock{cb}, nil
		}
	}
	var img outputGenerator
	if p, ok := b.(*markdown.Paragraph); ok {
		for _, inline := range p.Text.Inline {
			if mdImg, ok := inline.(*markdown.Image); ok {
				if strings.HasSuffix(mdImg.URL, ".evy.svg") {
					if img != nil {
						return nil, fmt.Errorf("%w: found second image", ErrBadMarkdownStructure)
					}
					img = image(mdImg.URL)
				}
			}
		}
	}
	return img, nil
}

// toOutputGeneratorList converts a markdown block into a list of
// outputGenerators if the block is markdown list and all list items
// can be converted to outputGenerators.
func toOutputGeneratorList(b markdown.Block) ([]outputGenerator, error) {
	list, ok := b.(*markdown.List)
	if !ok {
		return nil, nil
	}
	var outputGenerators []outputGenerator
	for _, item := range list.Items {
		foundGeneratorInItem := false
		for _, block := range item.(*markdown.Item).Blocks {
			outputGenerator, err := toOutputGenerator(block)
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
	generated, err := m.Question.genOutput(m.OutputType)
	if err != nil {
		return err
	}
	for i, choice := range m.AnswerChoices {
		choiceGen, err := choice.genOutput(m.OutputType)
		if err != nil {
			return err
		}
		if correctByIndex[i] && generated != choiceGen {
			return fmt.Errorf("%w: %q: %q != %q", ErrChoiceNotCorrect, indexToLetter(i), generated, choiceGen)
		}
		if !correctByIndex[i] && generated == choiceGen {
			return fmt.Errorf("%w: %q: %q == %q", ErrChoiceNotIncorrect, indexToLetter(i), generated, choiceGen)
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
	genOutput(OutputType) (string, error) // OutputType needed for generic evy programs that can have text _and_ image output
	outputType() OutputType
}

type image string

func (img image) genOutput(t OutputType) (string, error) {
	if t != imgOutput {
		return "", fmt.Errorf("%w: expected image output type, got %q", ErrBadOutput, OutputTypeToString[t])
	}
	b, err := os.ReadFile(string(img))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (img image) outputType() OutputType { return imgOutput }

type codeBlock struct {
	*markdown.CodeBlock
}

func (b *codeBlock) genOutput(t OutputType) (string, error) {
	switch {
	case b.Info == "evy" && t == textOutput:
	case b.Info == "evy" && t == imgOutput: // TODO
	case b.Info == "evy-out" && t == textOutput:
		return strings.Join(b.Text, "\n"), nil
	}
	return "", fmt.Errorf("%w: info: %q, %q", ErrBadCodeBlock, b.Info, OutputTypeToString[t])
}

func (b *codeBlock) outputType() OutputType {
	if b.Info == "evy-out" {
		return textOutput
	}
	return unknownOutput
}

func getOutputType(generators []outputGenerator) (OutputType, error) {
	outputType := unknownOutput
	for _, g := range generators {
		t := g.outputType()
		if outputType == unknownOutput {
			outputType = t
		} else if outputType != t {
			return unknownOutput, fmt.Errorf("%w: found text and image output, expected only one", ErrBadOutput)
		}
	}
	if outputType == unknownOutput {
		return unknownOutput, fmt.Errorf("%w: found neither text nor image output", ErrBadOutput)
	}
	return outputType, nil
}
