// Command site-gen generates the website files to be deployed to firebase.
//
// Usage: site-gen <src-dir> <dest-dir> <domain>
//
// When deploying to firebase (any other hosting site), we need to make a few
// changes to the HTML, CSS and JS files in the site:
//   - Replace href/values with leading paths of /discord, /docs, /learn and /play
//     with a subdomain instead, so /docs/foo with docs.<domain>/foo
//   - Rename .css, .js and .wasm files to include a short-sha of the SHA256 of the
//     contents of the file and update any references to those files in .html
//     files to include the filename with the short-sha. This is to perform
//     cache busting when the files change.
//   - Update the importmap in .html files to include the short-sha in the
//     javascript imports.
//     e.g. "./module/editor.js": "./module/editor.js"
//     becomes "./module/editor.js": "./module/editor.1a2b3c4d.js"
//   - Update the wasmImports map in .html files to include the short-sha in
//     wasm imports. The wasmImports allows for cache busting hashed filenames
//     for wasm files. The replacements are of the same form as the importmap.
//
// The site generation process copies the source hierarchy to a destination
// directory and performs these updates as it copies the files.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
)

type app struct {
	CacheBust bool   `help:"Rename .css, .js, and .wasm files to include short hash"`
	Domain    string `help:"Rewrite top-level paths to subdomains"`
	SrcDir    string `arg:"" required:""`
	DestDir   string `arg:"" required:""`
}

func main() {
	kctx := kong.Parse(&app{})
	kctx.FatalIfErrorf(kctx.Run())
}

func (a *app) Run() error {
	skippedFiles, renamedFiles, err := a.copyTree()
	if err != nil {
		return err
	}

	return a.copyHTMLFiles(skippedFiles, renamedFiles)
}

// Copy the contents of the `src` filetree to the `dest` directory. When we
// copy it, files with extension `.css`, `.js`, or `.wasm` are renamed to put a
// short sha into the name for cache busting purposes (e.g. foo.css ->
// foo.1a2b3c4d.css). We delay copying html files until after we have walked
// the src filetree and copy them in a second pass afterwards so that we can
// update any references to renamed files in them. Return the skipped files
// (html) and a map of the files we renamed, or an error if something went
// wrong.
func (a *app) copyTree() ([]string, map[string]string, error) { //nolint: gocognit
	skippedFiles := []string{}
	renamedFiles := make(map[string]string)

	srcFS := os.DirFS(a.SrcDir)
	err := fs.WalkDir(srcFS, ".", func(filename string, d fs.DirEntry, err error) error {
		if err != nil {
			// Errors from WalkDir do not include `src` in the path making
			// the error messages not useful. Add src back in.
			var pe *fs.PathError
			if errors.As(err, &pe) {
				pe.Path = filepath.Join(a.SrcDir, pe.Path)
				return pe
			}
			return err
		}
		srcfile := filepath.Join(a.SrcDir, filename)
		destfile := filepath.Join(a.DestDir, filename)
		ext := filepath.Ext(filename)

		switch mode := d.Type() & fs.ModeType; mode {
		case fs.ModeDir:
			return os.Mkdir(destfile, 0o777) //nolint:gosec // erroneous linter
		case fs.ModeSymlink:
			// Treat symlinks as files, duplicating the contents.
			// This is because sometimes we want to embed the site
			// into evy and Go embedding will not embed symlinks.
			// Symlinks to directories are hard (loops, recursion,
			// nested mappings and so on), so error out if we see
			// one.
			target, err := filepath.EvalSymlinks(srcfile)
			if err != nil {
				return err
			}
			fi, err := os.Stat(target)
			if err != nil {
				return err
			}
			if fi.Mode()&fs.ModeType == fs.ModeDir {
				//nolint:goerr113 // dynamic errors in package main is ok
				return fmt.Errorf("symlink dirs not allowed: %s", srcfile)
			}
			// continues past switch to copying files

		case 0: // normal file
			// handled below
		default:
			//nolint:goerr113 // dynamic errors in package main is ok
			return fmt.Errorf("unknown file type: %s: %s", mode, srcfile)
		}

		if ext == ".html" {
			skippedFiles = append(skippedFiles, filename)
			return nil
		}
		if a.CacheBust && (ext == ".js" || ext == ".css" || ext == ".wasm") {
			shortSha, err := hashFile(srcfile)
			if err != nil {
				return err
			}
			basename := strings.TrimSuffix(filepath.Base(filename), ext)
			target := basename + "." + shortSha + ext
			if _, ok := renamedFiles[filename]; ok {
				//nolint:goerr113 // dynamic errors in package main is ok
				return fmt.Errorf("duplicate filename: %s", srcfile)
			}
			renamedFiles[filename] = target
			if ext == ".js" {
				// also keep original JS filename for those who cannot use an `importmap` (e.g. ios 16.2)
				if err := copyFile(srcfile, destfile); err != nil {
					return err
				}
			}
			destfile = filepath.Join(filepath.Dir(destfile), target)
		}
		return copyFile(srcfile, destfile)
	})
	if err != nil {
		return nil, nil, err
	}
	return skippedFiles, renamedFiles, nil
}

func (a *app) copyHTMLFiles(skippedFiles []string, renamedFiles map[string]string) error {
	for _, filename := range skippedFiles {
		in, out, err := openInOut(filepath.Join(a.SrcDir, filename), filepath.Join(a.DestDir, filename))
		if err != nil {
			return err
		}
		defer in.Close() //nolint:errcheck // don't care about close failing on read-only files
		err = a.updateHTMLFile(out, in, filename, renamedFiles)
		if err != nil {
			out.Close() //nolint:errcheck,gosec // we're returning the more important error
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
	}
	return nil
}

// hashFile returns a short hash of the contents of filename. The short hash is
// 32 bits, or 8 chars[0-9a-f] and with 100 file changes in a year (cache
// expiry is one year) has a collision probability of less than 0.0000000005%.
func hashFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close() //nolint:errcheck // don't care about close failing on read-only files
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	sha := h.Sum(nil)
	return hex.EncodeToString(sha[:4]), nil
}

func openInOut(src, dest string) (io.ReadCloser, io.WriteCloser, error) {
	in, err := os.Open(src)
	if err != nil {
		return nil, nil, err
	}
	out, err := os.Create(dest)
	if err != nil {
		in.Close() //nolint:errcheck,gosec // we're returning the more important error
		return nil, nil, err
	}
	return in, out, nil
}

func copyFile(src, dest string) error {
	in, out, err := openInOut(src, dest)
	if err != nil {
		return err
	}
	defer in.Close() //nolint:errcheck // don't care about close failing on read-only files
	_, err = io.Copy(out, in)
	if err != nil {
		out.Close() //nolint:errcheck,gosec // we're returning the more important error
		return err
	}
	return out.Close()
}

var (
	subdomainRE = regexp.MustCompile(`(href|value)="/(discord|docs|learn|play)`)
	jscssRefRE  = regexp.MustCompile(`(href|src)="(.*\.(?:css|js))"`)
	importmapRE = regexp.MustCompile(`"(.*\.js)": "(.*\.js)"`)
	wasmmapRE   = regexp.MustCompile(`"(.*\.wasm)": "(.*\.wasm)"`)
)

// updateHTMLFile reads an HTML file from `r` and writes it to `w` making the
// following alterations:
//   - href and value attributes referencing /discord, /docs, /learn and /play
//     are transformed to top-level domains - discord.<domain>, docs.<domain>,
//     etc.
//   - href and src attributes referencing .css or .js files that have been
//     renamed to include their hash are updated to that name with the hash
//   - The .js files referenced in an importmap are updated if the referenced
//     .js file was renamed to include a hash.
//   - The .wasm files referenced in the wasmImports map are updated if the
//     referenced .wasm file was renamed to include a hash.
func (a *app) updateHTMLFile(w io.Writer, r io.Reader, filename string, renamedFiles map[string]string) error {
	inImportmap := false
	inWASMImports := false
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Rewrite top-level path to subdomain reference
		if a.Domain != "" {
			line = subdomainRE.ReplaceAllString(line, `$1="https://$2.`+a.Domain)
		}

		if a.CacheBust {
			// Track if we are in an importmap or wasmimports
			if strings.Contains(line, `<script type="importmap">`) {
				inImportmap = true
			}
			if inImportmap && strings.Contains(line, `</script>`) {
				inImportmap = false
			}
			if strings.Contains(line, `const wasmImports = {`) {
				inWASMImports = true
			}
			if inWASMImports && strings.Contains(line, `</script>`) {
				inWASMImports = false
			}

			line = updateRefs(filename, line, renamedFiles)
			if inImportmap {
				line = updateImportMap(filename, line, renamedFiles)
			}
			if inWASMImports {
				line = updateWASMImports(filename, line, renamedFiles)
			}
		}

		_, err := w.Write([]byte(line + "\n"))
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

func updateRefs(filename, line string, renamedFiles map[string]string) string {
	// Rewrite .js and .css in href and src attributes
	if matches := jscssRefRE.FindStringSubmatch(line); len(matches) > 0 {
		newname := getNewName(filename, matches[2], renamedFiles)
		if newname != "" {
			replacement := `$1="` + newname + `"`
			line = jscssRefRE.ReplaceAllString(line, replacement)
		}
	}
	return line
}

func updateImportMap(filename, line string, renamedFiles map[string]string) string {
	// Rewrite .js filenames in importmap
	if matches := importmapRE.FindStringSubmatch(line); len(matches) > 0 {
		newname := getNewName(filename, matches[2], renamedFiles)
		if newname != "" {
			replacement := `"$1": "./` + newname + `"`
			line = importmapRE.ReplaceAllString(line, replacement)
		}
	}
	return line
}

func updateWASMImports(filename, line string, renamedFiles map[string]string) string {
	// Rewrite .wasm filenames in wasm map
	if matches := wasmmapRE.FindStringSubmatch(line); len(matches) > 0 {
		newname := getNewName(filename, matches[2], renamedFiles)
		if newname != "" {
			replacement := `"$1": "./` + newname + `"`
			line = wasmmapRE.ReplaceAllString(line, replacement)
		}
	}
	return line
}

// getNewName returns the filename in `match` that appeared in `filename` as a
// renamed filename if it appears in `renamedFiles`. e.g. If the file
// `./play/index.html` contained a match of `../css/fonts.css` and the file
// `./css/fonts.css` was renamed to `fonts.12345678.css`, getNewName will
// return `../css/fonts.12345678.css`. If the file referenced by `match` was
// not renamed, an empty string is returned.
func getNewName(filename, match string, renamedFiles map[string]string) string {
	src := filepath.Join(filepath.Dir(filename), filepath.FromSlash(match))
	target := filepath.Clean(src)
	hashedName, ok := renamedFiles[target]
	if !ok {
		return ""
	}
	return path.Join(path.Dir(match), hashedName)
}
