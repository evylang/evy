package learn

import (
	"bytes"
	"fmt"
	"path/filepath"

	"evylang.dev/evy/pkg/md"
	"gopkg.in/yaml.v3"
	"rsc.io/markdown"
)

// CourseModel represents a  single course composed of multiple units.
type CourseModel struct {
	*configurableModel // used by functional options
	Doc                *markdown.Document
	Frontmatter        *courseFrontmatter
	name               string // flat markdown heading
	Units              []*UnitModel
}

// NewCourseModel returns a new unit model from a unit Markdown file or its
// contents.
func NewCourseModel(filename string, options ...Option) (*CourseModel, error) {
	course := &CourseModel{configurableModel: newConfigurableModel(filename, options)}
	course.cache[filename] = course
	if err := course.parseFrontmatterMD(); err != nil {
		return nil, err
	}
	if err := course.buildUnits(); err != nil {
		return nil, err
	}
	return course, nil
}

type courseFrontmatter struct {
	Type frontmatterType `yaml:"type,omitempty"`
}

func (m *CourseModel) buildUnits() error {
	relPaths := collectMDLinks(m.Doc)
	dir := filepath.Dir(m.Filename())
	opts := newOptions(m.ignoreSealed, m.privateKey, m.cache)

	for _, relPath := range relPaths {
		fname := filepath.Join(dir, relPath)
		model, err := newModel(fname, opts, m.cache)
		if err != nil {
			return err
		}
		if unit, ok := model.(*UnitModel); ok {
			m.Units = append(m.Units, unit)
		}
	}
	return nil
}

// ToHTML returns a complete standalone HTML document as string.
func (m *CourseModel) ToHTML(_ bool) (string, error) {
	md.Walk(m.Doc, md.RewriteLink)
	buf := &bytes.Buffer{}
	printHTML(m.Doc, buf)
	if err := m.printUnitBadgesHTML(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Name returns the name of the course model derived from the first heading.
func (m *CourseModel) Name() string {
	return m.name
}

func (m *CourseModel) printUnitBadgesHTML(buf *bytes.Buffer) error {
	courseDir := filepath.Dir(m.Filename())
	for _, unit := range m.Units {
		buf.WriteString("<h2>")
		h, ok := unit.Doc.Blocks[0].(*markdown.Heading)
		if !ok {
			buf.WriteString("Unit:" + unit.Filename() + "\n")
		} else {
			h.Text.PrintHTML(buf)
		}
		buf.WriteString("</h2>\n")
		if err := unit.printBadgesHTML(buf, courseDir); err != nil {
			return err
		}
	}
	return nil
}

func (m *CourseModel) parseFrontmatterMD() error {
	var err error
	if m.rawFrontmatter == "" && m.rawMD == "" {
		m.rawFrontmatter, m.rawMD, err = readSplitMDFile(m.Filename())
		if err != nil {
			return fmt.Errorf("%w (%s)", err, m.Filename())
		}
	}
	m.Frontmatter = &courseFrontmatter{}
	if err := yaml.Unmarshal([]byte(m.rawFrontmatter), m.Frontmatter); err != nil {
		return fmt.Errorf("%w: cannot process Course Markdown frontmatter: %w", ErrInvalidFrontmatter, err)
	}
	if m.Frontmatter.Type != "course" {
		return fmt.Errorf("%w: invalid frontmatter type %q, expected %q", ErrInvalidFrontmatter, m.Frontmatter.Type, "course")
	}

	m.Doc = parseMD(m.rawMD)
	if m.name, err = extractName(m.Doc); err != nil {
		return fmt.Errorf("%w (%s)", err, m.Filename())
	}
	return nil
}
