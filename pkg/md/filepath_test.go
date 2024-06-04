package md

import (
	"io/fs"
	"os"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestCopy(t *testing.T) {
	destdir := t.TempDir()
	gotMD, err := Copy("testdata/sample", destdir)
	assert.NoError(t, err)

	wantMD := []string{"README.md", "dir1/README.md", "dir1/dir2/file2.md", "dir1/file1.md"}
	assert.Equal(t, wantMD, gotMD)

	gotCopied := findFiles(destdir)
	wantCopied := []string{"dir1/bar.txt"}
	assert.Equal(t, wantCopied, gotCopied)

	gotHTML := make([]string, len(gotMD))
	for i, md := range gotMD {
		gotHTML[i] = HTMLFilename(md)
	}
	wantHTML := []string{"index.html", "dir1/index.html", "dir1/dir2/file2.html", "dir1/file1.html"}
	assert.Equal(t, wantHTML, gotHTML)
}

// findFiles finds recursively all file directory.
func findFiles(root string) []string {
	var files []string
	rootFS := os.DirFS(root)
	_ = fs.WalkDir(rootFS, ".", func(filename string, d fs.DirEntry, _ error) error {
		if !d.IsDir() {
			files = append(files, filename)
		}
		return nil
	})
	return files
}
