package learn

import (
	"bytes"
	"fmt"
	"path/filepath"

	"evylang.dev/evy/pkg/md"
	"gopkg.in/yaml.v3"
	"rsc.io/markdown"
)

// UnittestModel represents a unit test which has to be completed without
// mistakes to "master" a unit.
type UnittestModel struct {
	*configurableModel    // used by functional options
	Filename              string
	Doc                   *markdown.Document
	Frontmatter           *unittestFrontmatter
	QuestionsByDifficulty questionsByDifficulty
}

// NewUnittestModel returns a new unit model from a unit Markdown file or its
// contents.
func NewUnittestModel(filename string, options ...Option) (*UnittestModel, error) {
	unittest := &UnittestModel{
		Filename:              filename,
		QuestionsByDifficulty: questionsByDifficulty{},
		configurableModel:     newConfigurableModel(options),
	}
	unittest.cache[filename] = unittest
	if err := unittest.parseFrontmatterMD(); err != nil {
		return nil, err
	}
	if err := unittest.buildExercises(); err != nil {
		return nil, err
	}
	return unittest, nil
}

type unittestFrontmatter struct {
	Type            frontmatterType   `yaml:"type,omitempty"`
	Composition     []DifficultyCount `yaml:"composition,omitempty"`
	IgnoreExercises []string          `yaml:"ignore-exercises,omitempty"`
}

func (m *UnittestModel) buildExercises() error {
	mdFiles, err := filepath.Glob(filepath.Dir(m.Filename) + "/*/*.md")
	if err != nil {
		return fmt.Errorf("%w: cannot glob exercise and question *.md files: %w", ErrExercise, err)
	}
	opts := newOptions(m.ignoreSealed, m.privateKey, m.cache)
	ignoreExercises := map[string]bool{}
	for _, path := range m.Frontmatter.IgnoreExercises {
		p := filepath.Clean(filepath.Join(filepath.Dir(m.Filename), path))
		ignoreExercises[p] = true
	}
	for _, filename := range mdFiles {
		if ignoreExercises[filepath.Dir(filename)] {
			continue
		}
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
func (m *UnittestModel) ToHTML(withAnswersMarked bool) (string, error) {
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

func (m *UnittestModel) parseFrontmatterMD() error {
	var err error
	if m.rawFrontmatter == "" && m.rawMD == "" {
		m.rawFrontmatter, m.rawMD, err = readSplitMDFile(m.Filename)
		if err != nil {
			return fmt.Errorf("%w (%s)", err, m.Filename)
		}
	}
	m.Frontmatter = &unittestFrontmatter{}
	if err := yaml.Unmarshal([]byte(m.rawFrontmatter), m.Frontmatter); err != nil {
		return fmt.Errorf("%w: cannot process Unit Test Markdown frontmatter: %w", ErrInvalidFrontmatter, err)
	}
	if m.Frontmatter.Type != "unittest" {
		return fmt.Errorf("%w: invalid frontmatter type %q, expected %q", ErrInvalidFrontmatter, m.Frontmatter.Type, "unittest")
	}
	m.Doc = parseMD(m.rawMD)
	return nil
}
