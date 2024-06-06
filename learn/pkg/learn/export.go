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
	mdFiles, err := preparePaths(srcDir, destDir, exportOpts.WriteHTML)
	if err != nil {
		return err
	}
	globalAnswerKey := AnswerKey{}
	for _, mdFile := range mdFiles {
		mdf := filepath.Join(srcDir, mdFile)
		model, err := newModel(mdf, modelOpts)
		if err != nil {
			return err
		}
		if exportOpts.WriteAnswerKey {
			answerkey, err := model.ExportAnswerKey()
			if err != nil {
				return err
			}
			globalAnswerKey.merge(answerkey)
		}
		if exportOpts.WriteHTML {
			htmlFile := filepath.Join(destDir, md.HTMLFilename(mdFile))
			if err := writeHTMLFile(model, htmlFile, exportOpts.WithAnswersMarked); err != nil {
				return err
			}
		}
	}
	if exportOpts.WriteAnswerKey {
		answerKeyFile := filepath.Join(destDir, "answerkey.json")
		return writeAnswerKeyFile(globalAnswerKey, answerKeyFile)
	}
	return nil
}

func preparePaths(srcDir, destDir string, writeTree bool) ([]string, error) {
	if writeTree {
		return md.Copy(srcDir, destDir)
	}
	// use MkdirAll in case the directory already exists
	if err := os.MkdirAll(destDir, 0o777); err != nil {
		return nil, err
	}
	return md.FindFiles(srcDir)
}

func writeHTMLFile(model model, htmlFile string, withAnswersMarked bool) error {
	html, err := model.ToHTML(withAnswersMarked)
	if err != nil {
		return err
	}
	return os.WriteFile(htmlFile, []byte(html), 0o666)
}

func writeAnswerKeyFile(answerKey AnswerKey, answerKeyFile string) error {
	b, err := json.MarshalIndent(answerKey, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(answerKeyFile, b, 0o666)
}
