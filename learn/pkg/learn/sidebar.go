package learn

import (
	"fmt"
	"path/filepath"

	"evylang.dev/evy/pkg/md"
	"rsc.io/markdown"
)

func newSidebar(course *CourseModel, rootDir string) (*markdown.Document, error) {
	doc := &markdown.Document{}
	courseDir := filepath.Dir(course.Filename())
	courseRootDir := rootDir + filepath.Base(courseDir) + "/"
	courseHeading, err := sidebarHeading(course, courseRootDir)
	if err != nil {
		return nil, err
	}
	unitList, err := newUnitList(course, courseDir, courseRootDir)
	if err != nil {
		return nil, err
	}
	doc.Blocks = append(doc.Blocks, courseHeading, unitList)
	md.Walk(doc, md.RewriteLink)
	return doc, nil
}

func sidebarHeading(course *CourseModel, rootDir string) (*markdown.Heading, error) {
	heading, err := extractFirstHeading(course.Doc)
	if err != nil {
		return nil, err
	}
	courseLink := &markdown.Link{
		Inner: heading.Text.Inline,
		URL:   rootDir + filepath.Base(course.Filename()),
	}
	heading = &markdown.Heading{
		Level: 1,
		Text: &markdown.Text{
			Inline: []markdown.Inline{courseLink},
		},
	}
	return heading, nil
}

func newUnitList(course *CourseModel, courseDir, courseRootDir string) (markdown.Block, error) {
	list := &markdown.List{Bullet: '-'}
	for _, unit := range course.Units {
		unitHeading, err := extractFirstHeading(unit.Doc)
		if err != nil {
			return nil, err
		}

		relPath, err := filepath.Rel(courseDir, unit.Filename())
		if err != nil {
			return nil, fmt.Errorf("%w: %q not relative to %q: %w", ErrBadDirectoryStructure, unit.Filename(), courseDir, err)
		}
		item := newHeadingItem(unitHeading, courseRootDir+relPath)
		list.Items = append(list.Items, item)

		exerciseList, err := newSubList(unit, courseDir, courseRootDir)
		if err != nil {
			return nil, err
		}
		if exerciseList != nil {
			expander := &markdown.HTMLBlock{Text: []string{`<div class="expander"></div>`}}
			item.Blocks = append(item.Blocks, expander, exerciseList)
		}
	}
	return list, nil
}

// newSubList returns list of exercise, quiz, plain (reading) and unit test links.
func newSubList(unit *UnitModel, courseDir, courseRootDir string) (markdown.Block, error) {
	list := &markdown.List{Bullet: '-'}
	for _, m := range unit.OrderedModelsWithPlain {
		unitItemHeading, err := extractFirstHeading(m.Document())
		if err != nil {
			return nil, err
		}
		relPath, err := filepath.Rel(courseDir, m.Filename())
		if err != nil {
			return nil, fmt.Errorf("%w: %q not relative to %q: %w", ErrBadDirectoryStructure, unit.Filename(), courseDir, err)
		}
		emoji := modelEmoji(m)
		item := newHeadingItemWithEmoji(unitItemHeading, courseRootDir+relPath, emoji)
		list.Items = append(list.Items, item)
	}
	if len(list.Items) == 0 {
		return nil, nil
	}
	return list, nil
}

func newHeadingItem(heading *markdown.Heading, relPath string) *markdown.Item {
	link := &markdown.Link{
		Inner: heading.Text.Inline,
		URL:   relPath,
	}
	return &markdown.Item{
		Blocks: []markdown.Block{
			&markdown.Text{Inline: []markdown.Inline{link}},
		},
	}
}

func newHeadingItemWithEmoji(heading *markdown.Heading, relPath string, emoji string) *markdown.Item {
	link := &markdown.Link{
		Inner: []markdown.Inline{&markdown.Plain{Text: emoji + " "}},
		URL:   relPath,
	}
	link.Inner = append(link.Inner, heading.Text.Inline...)
	return &markdown.Item{
		Blocks: []markdown.Block{
			&markdown.Text{Inline: []markdown.Inline{link}},
		},
	}
}

func modelEmoji(m model) string {
	switch m.(type) {
	case *ExerciseModel:
		return "👉"
	case *QuizModel:
		return "✨"
	case *UnittestModel:
		return "⭐️"
	default:
		return "📖"
	}
}
