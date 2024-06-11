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
	"errors"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Errors for the learn package.
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
	ErrSealedAnswerNoKey    = errors.New("sealed answer without key")
	ErrSealedTooShort       = errors.New("sealed data is too short")

	ErrInvalidExportOptions = errors.New("invalid export options")
	ErrExercise             = errors.New("exercise error")

	ErrInvalidFileHierarchy = errors.New("invalid file hierarchy")
)

type model interface {
	ToHTML(withAnswersMarked bool) (string, error)
}

type plainMD string

func (s plainMD) ToHTML(_ bool) (string, error) {
	return md2HTML(string(s)), nil
}

func newModel(mdFile string, opts []Option, modelCache map[string]model) (model, error) {
	mdFile = filepath.Clean(mdFile)
	if m, ok := modelCache[mdFile]; ok {
		return m, nil
	}
	frontmatterString, mdString, err := readSplitMDFile(mdFile)
	if err != nil {
		return nil, err
	}
	var model model
	if frontmatterString == "" {
		model = plainMD(mdString)
	} else {
		model, err = newModelWithFrontmatter(mdFile, frontmatterString, mdString, opts)
		if err != nil {
			return nil, err
		}
	}
	modelCache[mdFile] = model
	return model, nil
}

func newModelWithFrontmatter(mdFile, frontmatterString, mdString string, opts []Option) (model, error) {
	opts = append([]Option{WithRawMD(frontmatterString, mdString)}, opts...)
	fm := &baseFrontmatter{}
	if err := yaml.Unmarshal([]byte(frontmatterString), fm); err != nil {
		return nil, fmt.Errorf("%w: cannot process Markdown frontmatter: %w", ErrInvalidFrontmatter, err)
	}

	switch fm.Type { // "course", "unit", "exercise", "question"
	case "question":
		return NewQuestionModel(mdFile, opts...)
	case "exercise":
		return NewExerciseModel(mdFile, opts...)
	case "unit":
		return NewUnitModel(mdFile, opts...)
	case "quiz":
		return NewQuizModel(mdFile, opts...)
	case "unittest":
		return NewUnittestModel(mdFile, opts...)
	case "course":
		return NewCourseModel(mdFile, opts...)
	}
	return nil, fmt.Errorf("unsupported frontmatter type %q", string(fm.Type)) //nolint:err113 // dynamic errors in main are fine.
}

type baseFrontmatter struct {
	Type frontmatterType `yaml:"type,omitempty"`
}
