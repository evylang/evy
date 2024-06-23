package learn

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"evylang.dev/evy/pkg/md"
)

// ExportOptions contains options for exporting answer key and HTML files.
type ExportOptions struct {
	WriteAnswerKey    bool
	WriteHTML         bool
	WithAnswersMarked bool
	SelfContained     bool /* CSS, JS, Favicon links vs standalone embeds*/
	WriteCatalog      bool
}

func (opts ExportOptions) validate() error {
	if !opts.WriteHTML && !opts.WriteAnswerKey && !opts.WriteCatalog {
		return fmt.Errorf("%w: at least one of WriteHTML, WriteAnswerKey or WriteCatalog must be true", ErrInvalidExportOptions)
	}
	if !opts.WriteHTML && opts.WithAnswersMarked {
		return fmt.Errorf("%w: WithAnswersMarked requires WriteHTML", ErrInvalidExportOptions)
	}
	if !opts.WriteHTML && opts.SelfContained {
		return fmt.Errorf("%w: SelfContained requires WriteHTML", ErrInvalidExportOptions)
	}
	return nil
}

// Export exports answer key and HTML files from srcDir containing Markdown
// files to destDir.
func Export(srcDir, destDir string, exportOpts ExportOptions, modelOpts ...Option) error {
	if err := exportOpts.validate(); err != nil {
		return err
	}
	mdFiles, err := md.FindFiles(srcDir)
	if err != nil {
		return err
	}
	models, err := newModels(srcDir, mdFiles, modelOpts)
	if err != nil {
		return err
	}
	if exportOpts.WriteHTML {
		if err := writeHTMLFiles(models, srcDir, destDir, exportOpts); err != nil {
			return err
		}
	}
	if exportOpts.WriteAnswerKey {
		answerKeyFile := filepath.Join(destDir, "answerkey.json")
		if err := writeAnswerKeyFile(models, answerKeyFile); err != nil {
			return err
		}
	}
	if exportOpts.WriteCatalog {
		catalogFile := filepath.Join(destDir, "catalog.json")
		if err := writeCatalogFile(models, catalogFile); err != nil {
			return err
		}
	}
	return nil
}

func newModels(srcDir string, mdFiles []string, modelOpts []Option) ([]model, error) {
	modelCache := map[string]model{}
	modelOpts = append(modelOpts, withCache(modelCache))
	models := make([]model, 0, len(mdFiles))
	for _, mdFile := range mdFiles {
		mdf := filepath.Join(srcDir, mdFile)
		model, err := newModel(mdf, modelOpts, modelCache)
		if err != nil {
			return nil, err
		}
		qmodel, ok := model.(*QuestionModel)
		if ok && qmodel.hasSubQuestions() {
			for _, qm := range qmodel.subQuestions {
				models = append(models, qm)
			}
			continue
		}
		models = append(models, model)
	}
	if err := validateModelPaths(models); err != nil {
		return nil, err
	}
	return models, nil
}

func validateModelPaths(models []model) error {
	p, err := newPathsByType(models)
	if err != nil {
		return err
	}
	if len(p.questions) == 0 {
		return fmt.Errorf("%w: no question Markdown file found", ErrInvalidFileHierarchy)
	}
	courseDepth := len(splitPath(p.questions[0])) - 2 // relative depth from srcDir
	for idx, paths := range [][]string{p.courses, p.unitsWithQuizzes, p.questions} {
		for _, path := range paths {
			depth := len(splitPath(path))
			if depth != courseDepth+idx {
				return fmt.Errorf("%w: invalid directory depth for %q, expected %d, got %d", ErrInvalidFileHierarchy, path, courseDepth+idx, depth)
			}
		}
	}
	seen := map[string]string{}
	unitCoursePaths := append([]string{}, p.courses...)
	unitCoursePaths = append(unitCoursePaths, p.units...)
	for _, path := range unitCoursePaths {
		dir := filepath.Dir(path)
		if seen[dir] != "" {
			return fmt.Errorf("%w: only one unit or exercise Markdown file allowed per directory found %q and %q", ErrInvalidFileHierarchy, seen[dir], path)
		}
		seen[dir] = path
	}
	return nil
}

type pathsByType struct {
	courses, units, quizzes, unitsWithQuizzes, questions []string
}

func newPathsByType(models []model) (pathsByType, error) {
	byType := pathsByType{}
	for _, m := range models {
		switch m := m.(type) { // "course", "unit", "exercise", "question"
		case *QuestionModel:
			byType.questions = append(byType.questions, m.Filename())
		case *ExerciseModel:
			byType.questions = append(byType.questions, m.Filename())
		case *UnitModel:
			byType.units = append(byType.units, m.Filename())
		case *UnittestModel, *QuizModel:
			byType.quizzes = append(byType.quizzes, m.Filename())
		case *CourseModel:
			byType.courses = append(byType.courses, m.Filename())
		case *plainMD: // plain markdown files can be anywhere, no-op
		default:
			return byType, fmt.Errorf("%w: unknown model type: %T", ErrInconsistentMdoel, m)
		}
	}
	byType.unitsWithQuizzes = append([]string{}, byType.units...)
	byType.unitsWithQuizzes = append(byType.unitsWithQuizzes, byType.quizzes...)
	return byType, nil
}

func writeHTMLFiles(models []model, srcDir, destDir string, opts ExportOptions) error {
	if _, err := md.Copy(srcDir, destDir); err != nil {
		return err
	}
	for _, model := range models {
		qmodel, ok := model.(*QuestionModel)
		if ok && qmodel.hasSubQuestions() {
			continue
		}
		mdFile, err := filepath.Rel(srcDir, model.Filename())
		if err != nil {
			return fmt.Errorf("%w: %w: %s", ErrInconsistentMdoel, err, model.Filename())
		}
		htmlFile := filepath.Join(destDir, md.HTMLFilename(mdFile))
		content, err := model.ToHTML(opts.WithAnswersMarked)
		if err != nil {
			return err
		}
		tmplData := newTmplData(mdFile, model.Name(), content, !opts.SelfContained)
		if err := writeHTMLFile(htmlFile, tmplData); err != nil {
			return err
		}
	}
	return nil
}

func writeHTMLFile(htmlFile string, tmplData tmplData) error {
	out, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(out, tmplData); err != nil {
		out.Close() //nolint:errcheck,gosec // we're returning the more important error
		return err
	}
	return out.Close()
}

func writeAnswerKeyFile(models []model, answerKeyFile string) error {
	// use MkdirAll in case the directory already exists
	if err := os.MkdirAll(filepath.Dir(answerKeyFile), 0o777); err != nil {
		return err
	}
	answerKey := AnswerKey{}
	for _, m := range models {
		qmodel, ok := m.(*QuestionModel)
		if ok {
			ak, err := qmodel.ExportAnswerKey()
			if err != nil {
				return err
			}
			answerKey.merge(ak)
		}
	}
	b, err := json.MarshalIndent(answerKey, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(answerKeyFile, b, 0o666)
}

func writeCatalogFile(models []model, catalogFile string) error {
	// use MkdirAll in case the directory already exists
	if err := os.MkdirAll(filepath.Dir(catalogFile), 0o777); err != nil {
		return err
	}
	catalogs := map[string]Course{}
	var err error
	for _, m := range models {
		if cmodel, ok := m.(*CourseModel); ok {
			catalog := NewCourseCatalog(cmodel)
			catalogs[catalog.PartialID] = catalog
		}
	}
	b, err := json.MarshalIndent(catalogs, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(catalogFile, b, 0o666)
}

//go:embed tmpl/*
var tmplFS embed.FS

var (
	tmplFuncMap = template.FuncMap{"indent": indent}
	tmpl        = template.Must(template.New("learn.html.tmpl").Funcs(tmplFuncMap).ParseFS(tmplFS, "tmpl/learn.html.tmpl"))
)

func indent(indentCount int, s string) string {
	parts := strings.Split(s, "\n")
	if len(parts) == 0 {
		return s
	}
	if strings.TrimSpace(parts[len(parts)-1]) == "" {
		parts = parts[:len(parts)-1]
	}
	indent := strings.Repeat("  ", indentCount)
	return strings.Join(parts, "\n"+indent)
}

type tmplData struct {
	Root          string
	Title         string
	Content       string
	DefaultCSS    string
	CSSFiles      []string
	WithHeadLinks bool
}

//go:embed tmpl/default.css
var defaultCSS string

func newTmplData(mdFile, title, content string, withHeadLinks bool) tmplData {
	return tmplData{
		Root:          md.ToRoot(mdFile),
		Title:         title,
		Content:       content,
		DefaultCSS:    defaultCSS,
		CSSFiles:      []string{"index.css"},
		WithHeadLinks: withHeadLinks,
	}
}
