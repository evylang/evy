package learn

import (
	"bytes"
	"fmt"
	"path/filepath"

	"evylang.dev/evy/pkg/md"
	"gopkg.in/yaml.v3"
	"rsc.io/markdown"
)

// QuizModel represents a quiz which can be taken optionally to improve the
// exercise score. If there is already familiarity with the exercise content,
// the quiz can be taken instead of individual exercises.
type QuizModel struct {
	*configurableModel    // used by functional options
	Filename              string
	Doc                   *markdown.Document
	Frontmatter           *quizFrontmatter
	QuestionsByDifficulty questionsByDifficulty
}

// NewQuizModel returns a new unit model from a unit Markdown file or its
// contents.
func NewQuizModel(filename string, options ...Option) (*QuizModel, error) {
	quiz := &QuizModel{
		Filename:              filename,
		QuestionsByDifficulty: questionsByDifficulty{},
		configurableModel:     newConfigurableModel(options),
	}
	quiz.cache[filename] = quiz
	if err := quiz.parseFrontmatterMD(); err != nil {
		return nil, err
	}
	if err := quiz.buildExercises(); err != nil {
		return nil, err
	}
	return quiz, nil
}

type quizFrontmatter struct {
	Type        frontmatterType   `yaml:"type,omitempty"`
	Composition []DifficultyCount `yaml:"composition,omitempty"`
	Exercises   []string          `yaml:"exercises,omitempty"`
}

func (m *QuizModel) buildExercises() error {
	dir := filepath.Dir(m.Filename)
	var mdFiles []string
	for _, path := range m.Frontmatter.Exercises {
		files, err := filepath.Glob(filepath.Join(dir, path) + "/*.md")
		if err != nil {
			return fmt.Errorf("%w: cannot glob *.md files: %w", ErrExercise, err)
		}
		mdFiles = append(mdFiles, files...)
	}
	opts := newOptions(m.ignoreSealed, m.privateKey, m.cache)
	for _, filename := range mdFiles {
		model, err := newModel(filename, opts, m.cache)
		if err != nil {
			return err
		}
		exercise, ok := model.(*ExerciseModel)
		if !ok {
			continue
		}
		m.QuestionsByDifficulty.merge(exercise.QuestionsByDifficulty)
	}
	return m.QuestionsByDifficulty.validate(m.Frontmatter.Composition)
}

// ToHTML returns a complete standalone HTML document as string.
func (m *QuizModel) ToHTML(withAnswersMarked bool) (string, error) {
	md.Walk(m.Doc, md.RewriteLink)
	buf := &bytes.Buffer{}
	buf.WriteString(prefixHTML)
	m.Doc.PrintHTML(buf)
	if withAnswersMarked {
		printComposition(buf, m.Frontmatter.Composition)
		m.QuestionsByDifficulty.PrintHTML(buf)
	}
	buf.WriteString(suffixHTML)
	return buf.String(), nil
}

func (m *QuizModel) parseFrontmatterMD() error {
	var err error
	if m.rawFrontmatter == "" && m.rawMD == "" {
		m.rawFrontmatter, m.rawMD, err = readSplitMDFile(m.Filename)
		if err != nil {
			return fmt.Errorf("%w (%s)", err, m.Filename)
		}
	}
	m.Frontmatter = &quizFrontmatter{}
	if err := yaml.Unmarshal([]byte(m.rawFrontmatter), m.Frontmatter); err != nil {
		return fmt.Errorf("%w: cannot process Exercise Markdown frontmatter: %w", ErrInvalidFrontmatter, err)
	}
	if m.Frontmatter.Type != "quiz" {
		return fmt.Errorf("%w: invalid frontmatter type %q, expected %q", ErrInvalidFrontmatter, m.Frontmatter.Type, "quiz")
	}

	m.Doc = parseMD(m.rawMD)
	return nil
}