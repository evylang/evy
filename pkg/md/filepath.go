package md

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Copy copies the contents of the `src` directory to the `dest`
// directory. It skips over *.md files and returns their files names in a
// string slice, typically to generate HTML files from it.
func Copy(srcDir, destDir string) ([]string, error) {
	mdFiles := []string{}
	srcFS := os.DirFS(srcDir)
	err := fs.WalkDir(srcFS, ".", func(filename string, d fs.DirEntry, err error) error {
		if err != nil {
			// Errors from WalkDir do not include `src` in the path making
			// the error messages not useful. Add src back in.
			var pe *fs.PathError
			if errors.As(err, &pe) {
				pe.Path = filepath.Join(srcDir, pe.Path)
				return pe
			}
			return err
		}
		srcfile := filepath.Join(srcDir, filename)
		destfile := filepath.Join(destDir, filename)

		if d.IsDir() {
			// use MkdirAll in case the directory already exists
			return os.MkdirAll(destfile, 0o777)
		}

		if filepath.Ext(filename) == ".md" {
			mdFiles = append(mdFiles, filename)
			return nil
		}
		sf, err := os.Open(srcfile)
		if err != nil {
			return err
		}
		defer sf.Close() //nolint:errcheck // don't care about close failing on read-only files
		df, err := os.Create(destfile)
		if err != nil {
			df.Close() //nolint:errcheck,gosec // we're returning the more important error
			return err
		}
		if _, err := io.Copy(df, sf); err != nil {
			df.Close() //nolint:errcheck,gosec // we're returning the more important error
			return err
		}
		return df.Close()
	})
	if err != nil {
		return nil, err
	}
	return mdFiles, nil
}

// HTMLFilename returns the HTML filename for a given markdown filename.
// README.md is converted to index.html, and all other .md files are
// converted .html file names with the same base name stem.
func HTMLFilename(mdf string) string {
	if filepath.Base(mdf) == "README.md" {
		return filepath.Join(filepath.Dir(mdf), "index.html")
	}
	if filename, found := strings.CutSuffix(mdf, ".md"); found {
		return filename + ".html"
	}
	return mdf
}

// ToRoot returns a relative path to the root of the given filename. It
// requires a final slash to be concatenated with further paths, so it can be
// used in templates more directly, e.g.:
//
//	<link rel="stylesheet" href="{{.Root}}/css/index.css" type="text/css" />
func ToRoot(filename string) string {
	if c := strings.Count(filename, string(os.PathSeparator)); c > 0 {
		return strings.Repeat("/..", c)[1:]
	}
	return "."
}

// FindFiles finds recursively all file with .md extension in given directory
// and returns a slice of their names.
func FindFiles(root string) ([]string, error) {
	var mdFiles []string
	rootFS := os.DirFS(root)
	err := fs.WalkDir(rootFS, ".", func(filename string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(d.Name()) == ".md" {
			mdFiles = append(mdFiles, filename)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mdFiles, nil
}
