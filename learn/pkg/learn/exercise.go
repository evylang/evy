package learn

import (
	"bytes"
	"fmt"
	"path/filepath"

	"evylang.dev/evy/pkg/md"
	"gopkg.in/yaml.v3"
	"rsc.io/markdown"
)

// ExerciseModel represents a single exercise with multiple questions. The
// question composition is defined in the frontmatter, for example there
// maybe 6 easy and 7 hard questions associated with a specific exercise, but
// when the question is presented to a student only 3 easy and then 2 hard
// questions are selected.
type ExerciseModel struct {
	*configurableModel
	Doc                   *markdown.Document
	Frontmatter           *exerciseFrontmatter
	name                  string
	Questions             []*QuestionModel
	QuestionsByDifficulty questionsByDifficulty
}

// NewExerciseModel returns a new exercise model from an exercise Markdown file.
func NewExerciseModel(filename string, options ...Option) (*ExerciseModel, error) {
	exercise := &ExerciseModel{
		QuestionsByDifficulty: map[string][]*QuestionModel{},
		configurableModel:     newConfigurableModel(filename, options),
	}
	exercise.cache[filename] = exercise
	if err := exercise.parseFrontmatterMD(); err != nil {
		return nil, err
	}
	if err := exercise.buildQuestions(); err != nil {
		return nil, err
	}
	if err := exercise.QuestionsByDifficulty.validate(exercise.Frontmatter.Composition); err != nil {
		return nil, fmt.Errorf("%w: %w: %s", ErrExercise, err, exercise.Filename())
	}
	return exercise, nil
}

type exerciseFrontmatter struct {
	Type        frontmatterType   `yaml:"type,omitempty"`
	Composition []DifficultyCount `yaml:"composition,omitempty"`
}

// ToHTML returns a complete standalone HTML document as string.
func (m *ExerciseModel) ToHTML(withMarked bool) (string, error) {
	buf := &bytes.Buffer{}
	md.Walk(m.Doc, md.RewriteLink)
	printHTML(m.Doc, buf)
	printComposition(buf, m.Frontmatter.Composition)
	for _, d := range validDifficulties {
		for _, question := range m.QuestionsByDifficulty[d] {
			if !question.hasSubQuestions() {
				if err := question.PrintHTML(buf, withMarked); err != nil {
					return "", err
				}
				continue
			}
			for _, subQuestion := range question.subQuestions {
				if err := subQuestion.PrintHTML(buf, withMarked); err != nil {
					return "", err
				}
			}
		}
	}
	return buf.String(), nil
}

// Name returns the name of the exercise model derived from the first heading.
func (m *ExerciseModel) Name() string {
	return m.name
}

// Document returns the markdown ast root node.
func (m *ExerciseModel) Document() *markdown.Document {
	return m.Doc
}

func (m *ExerciseModel) buildQuestions() error {
	questionFiles, err := filepath.Glob(filepath.Dir(m.Filename()) + "/*.md")
	if err != nil {
		return fmt.Errorf("%w: cannot glob *.md files: %w", ErrExercise, err)
	}
	questionOpts := newOptions(m.ignoreSealed, m.privateKey, m.cache)
	for _, filename := range questionFiles {
		model, err := newModel(filename, questionOpts, m.cache)
		if err != nil {
			return err
		}
		question, ok := model.(*QuestionModel)
		if !ok || (m.ignoreSealed && question.IsSealed()) {
			continue
		}
		difficulty := string(question.Frontmatter.Difficulty)
		m.QuestionsByDifficulty[difficulty] = append(m.QuestionsByDifficulty[difficulty], question)
		m.Questions = append(m.Questions, question)
	}
	return nil
}

func (m *ExerciseModel) parseFrontmatterMD() error {
	var err error
	if m.rawFrontmatter == "" && m.rawMD == "" {
		m.rawFrontmatter, m.rawMD, err = readSplitMDFile(m.Filename())
		if err != nil {
			return fmt.Errorf("%w (%s)", err, m.Filename())
		}
	}
	m.Frontmatter = &exerciseFrontmatter{}
	if err := yaml.Unmarshal([]byte(m.rawFrontmatter), m.Frontmatter); err != nil {
		return fmt.Errorf("%w: cannot process Exercise Markdown frontmatter: %w", ErrInvalidFrontmatter, err)
	}
	if m.Frontmatter.Type != "exercise" {
		return fmt.Errorf("%w: invalid frontmatter type %q, expected %q", ErrInvalidFrontmatter, m.Frontmatter.Type, "exercise")
	}

	m.Doc = parseMD(m.rawMD)
	if m.name, err = extractName(m.Doc); err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename())
	}
	return nil
}
