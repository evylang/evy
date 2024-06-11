package learn

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"evylang.dev/evy/pkg/md"
)

// ExportOptions contains options for exporting answer key and HTML files.
type ExportOptions struct {
	WriteAnswerKey    bool
	WriteHTML         bool
	WithAnswersMarked bool
}

func (opts ExportOptions) validate() error {
	if !opts.WriteHTML && !opts.WriteAnswerKey {
		return fmt.Errorf("%w: at least one of WriteHTML or WriteAnswerKey must be true", ErrInvalidExportOptions)
	}
	if !opts.WriteHTML && opts.WithAnswersMarked {
		return fmt.Errorf("%w: WithAnswersMarked requires WriteHTML", ErrInvalidExportOptions)
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
		if err := writeHTMLFiles(mdFiles, models, srcDir, destDir, exportOpts.WithAnswersMarked); err != nil {
			return err
		}
	}
	if exportOpts.WriteAnswerKey {
		answerKeyFile := filepath.Join(destDir, "answerkey.json")
		return writeAnswerKeyFile(models, answerKeyFile)
	}
	return nil
}

func newModels(srcDir string, mdFiles []string, modelOpts []Option) ([]model, error) {
	modelCache := map[string]model{}
	modelOpts = append(modelOpts, withCache(modelCache))
	models := make([]model, len(mdFiles))
	var err error
	for i, mdFile := range mdFiles {
		mdf := filepath.Join(srcDir, mdFile)
		models[i], err = newModel(mdf, modelOpts, modelCache)
		if err != nil {
			return nil, err
		}
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
			byType.questions = append(byType.questions, m.Filename)
		case *ExerciseModel:
			byType.questions = append(byType.questions, m.Filename)
		case *UnitModel:
			byType.units = append(byType.units, m.Filename)
		case *UnittestModel:
			byType.quizzes = append(byType.quizzes, m.Filename)
		case *QuizModel:
			byType.quizzes = append(byType.quizzes, m.Filename)
		// case *CourseModel:
		// 	byType.courses = append(byType.courses, m.Filename)
		case plainMD: // plain markdown files can be anywhere, no-op
		default:
			return byType, fmt.Errorf("%w: unknown model type: %T", ErrInconsistentMdoel, m)
		}
	}
	byType.unitsWithQuizzes = append([]string{}, byType.units...)
	byType.unitsWithQuizzes = append(byType.unitsWithQuizzes, byType.quizzes...)
	return byType, nil
}

func writeHTMLFiles(mdFiles []string, models []model, srcDir, destDir string, withAnswersMarked bool) error {
	if _, err := md.Copy(srcDir, destDir); err != nil {
		return err
	}
	for i, mdFile := range mdFiles {
		htmlFile := filepath.Join(destDir, md.HTMLFilename(mdFile))
		html, err := models[i].ToHTML(withAnswersMarked)
		if err != nil {
			return err
		}
		if err := os.WriteFile(htmlFile, []byte(html), 0o666); err != nil {
			return err
		}
	}
	return nil
}

func writeAnswerKeyFile(models []model, answerKeyFile string) error {
	// use MkdirAll in case the directory already exists
	if err := os.MkdirAll(filepath.Dir(answerKeyFile), 0o777); err != nil {
		return err
	}
	answerKey := AnswerKey{}
	for _, m := range models {
		if qmodel, ok := m.(*QuestionModel); ok {
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
