package learn

import (
	"bytes"
	"fmt"
	"path/filepath"

	"evylang.dev/evy/pkg/md"
	"gopkg.in/yaml.v3"
	"rsc.io/markdown"
)

// UnitModel represents a single unit with multiple exercises composed of
// several questions. A unit typically also has an optional quiz after two
// or more exercises which presents and opportunity to level up. At the end
// of the unit there is typically a unit test which needs to be completed
// with perfect score so that the unit is considered mastered.
//
// Exercises, quizzes and the final unit tests appear in the same order as
// listed in the unit.md file, e.g. in sidebar navigation or on badge
// summary. The order is captured in the OrderedModels slice.
type UnitModel struct {
	*configurableModel // used by functional options
	Filename           string
	Doc                *markdown.Document
	Frontmatter        *unitFrontmatter
	OrderedModels      []model // exercises, quizzes, unittests
}

// NewUnitModel returns a new unit model from a unit Markdown file or its
// contents.
func NewUnitModel(filename string, options ...Option) (*UnitModel, error) {
	unit := &UnitModel{
		Filename:          filename,
		configurableModel: newConfigurableModel(options),
	}
	unit.cache[filename] = unit
	if err := unit.parseFrontmatterMD(); err != nil {
		return nil, err
	}
	if err := unit.buildModels(); err != nil {
		return nil, err
	}
	return unit, nil
}

type unitFrontmatter struct {
	Type frontmatterType `yaml:"type,omitempty"`
}

// ToHTML returns a complete standalone HTML document as string.
func (m *UnitModel) ToHTML(_ bool) (string, error) {
	md.Walk(m.Doc, md.RewriteLink)
	buf := &bytes.Buffer{}
	buf.WriteString(prefixHTML)
	m.Doc.Blocks[0].PrintHTML(buf)
	unitDir := filepath.Dir(m.Filename)
	if err := m.printBadgesHTML(buf, unitDir); err != nil {
		return "", err
	}
	for _, block := range m.Doc.Blocks[1:] {
		block.PrintHTML(buf)
	}
	buf.WriteString(suffixHTML)
	return buf.String(), nil
}

func (m *UnitModel) printBadgesHTML(buf *bytes.Buffer, baseDir string) error {
	for _, model := range m.OrderedModels {
		badge, ref, err := badgeURL(model, baseDir)
		if err != nil {
			return err
		}
		html := `<a href="` + ref + `">` + badge + `</a>` + "\n"
		buf.WriteString(html)
	}
	return nil
}

func badgeURL(m model, baseDir string) (string, string, error) {
	var badge, fname string
	switch m := m.(type) {
	case *ExerciseModel:
		badge = "🔲"
		fname = m.Filename
	case *QuizModel:
		badge = "✨"
		fname = m.Filename
	case *UnittestModel:
		badge = "⭐️"
		fname = m.Filename
	default:
		return "", "", fmt.Errorf("%w: unit link: unknown model type %T", ErrInconsistentMdoel, m)
	}
	relPath, err := filepath.Rel(baseDir, fname)
	if err != nil {
		return "", "", fmt.Errorf("%w: cannot create relative path to exercise, quiz or unittest", err)
	}
	ref := md.HTMLFilename(relPath)
	return badge, ref, nil
}

func (m *UnitModel) buildModels() error {
	relPaths := collectMDLinks(m.Doc)
	dir := filepath.Dir(m.Filename)
	opts := newOptions(m.ignoreSealed, m.privateKey, m.cache)

	for _, relPath := range relPaths {
		fname := filepath.Join(dir, relPath)
		model, err := newModel(fname, opts, m.cache)
		if err != nil {
			return fmt.Errorf("%w: %s", err, fname)
		}
		switch model.(type) {
		case *ExerciseModel, *QuizModel, *UnittestModel:
			m.OrderedModels = append(m.OrderedModels, model)
		}
	}
	return nil
}

func (m *UnitModel) parseFrontmatterMD() error {
	var err error
	if m.rawFrontmatter == "" && m.rawMD == "" {
		m.rawFrontmatter, m.rawMD, err = readSplitMDFile(m.Filename)
		if err != nil {
			return fmt.Errorf("%w (%s)", err, m.Filename)
		}
	}
	m.Frontmatter = &unitFrontmatter{}
	if err := yaml.Unmarshal([]byte(m.rawFrontmatter), m.Frontmatter); err != nil {
		return fmt.Errorf("%w: cannot process Unit Markdown frontmatter: %w", ErrInvalidFrontmatter, err)
	}
	if m.Frontmatter.Type != "unit" {
		return fmt.Errorf("%w: invalid frontmatter type %q, expected %q", ErrInvalidFrontmatter, m.Frontmatter.Type, "unit")
	}

	m.Doc = parseMD(m.rawMD)
	if len(m.Doc.Blocks) == 0 {
		return fmt.Errorf("%w: no content in unit Markdown file", ErrBadMarkdownStructure)
	}
	if _, ok := m.Doc.Blocks[0].(*markdown.Heading); !ok {
		return fmt.Errorf("%w: first markdown element in unit Markdown file must be heading", ErrBadMarkdownStructure)
	}
	return nil
}
