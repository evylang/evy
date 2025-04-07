package learn

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestExportAll(t *testing.T) {
	destdir := t.TempDir()
	srcdir := "testdata/course1"
	goldendir := "testdata/golden/export/all"

	opts := ExportOptions{
		WriteHTML:         true,
		WriteAnswerKey:    true,
		WithAnswersMarked: true,
		WriteCatalog:      true,
		SelfContained:     true,
	}
	err := Export(srcdir, destdir, opts, WithPrivateKey(testKeyPrivate))
	assert.NoError(t, err)
	assertSameContents(t, goldendir, destdir)
}

func TestExportHTMLNoPrivateKey(t *testing.T) {
	destdir := t.TempDir()
	srcdir := "testdata/course1"
	goldendir := "testdata/golden/export/html-no-private-key"

	opts := ExportOptions{
		WriteHTML:     true,
		SelfContained: true,
	}
	err := Export(srcdir, destdir, opts, WithIgnoreSealed())
	assert.NoError(t, err)
	assertSameContents(t, goldendir, destdir)
}

func assertSameContents(t *testing.T, wantDir, gotDir string) {
	t.Helper()
	wantFiles := findFiles(wantDir)
	gotFiles := findFiles(gotDir)
	if slices.Compare(wantFiles, gotFiles) != 0 {
		missingWantFiles := diff(wantFiles, gotFiles)
		additionalGotFiles := diff(gotFiles, wantFiles)
		t.Errorf("want and got directories do not have the same files.\nmissing files: %v\nextra files:  %v\n", missingWantFiles, additionalGotFiles)
	}

	for _, filename := range wantFiles {
		wantFile := filepath.Join(wantDir, filename)
		want, err := os.ReadFile(wantFile)
		assert.NoError(t, err)
		gotFile := filepath.Join(gotDir, filename)
		got, err := os.ReadFile(gotFile)
		assert.NoError(t, err)
		if !bytes.Equal(want, got) {
			t.Errorf("files %s and %s are not equal", wantFile, gotFile)
		}
	}
}

func diff(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// findFiles finds all files in directory recursively.
func findFiles(root string) []string {
	var files []string
	rootFS := os.DirFS(root)
	_ = fs.WalkDir(rootFS, ".", func(filename string, d fs.DirEntry, _ error) error {
		if !d.IsDir() && filepath.Ext(filename) != ".DS_Store" {
			files = append(files, filename)
		}
		return nil
	})
	return files
}
