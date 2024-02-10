// Command md is a markdown processing tool
//
// md generates evy frontend code
package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alecthomas/kong"
	"rsc.io/markdown"
)

//go:embed tmpl/*
var tmplFS embed.FS

type app struct {
	SrcDir  string `arg:"" type:"existingdir" help:"source directory" placeholder:"SRCDIR"`
	DestDir string `arg:"" help:"target directory" placeholder:"DESTDIR"`
}

func main() {
	kctx := kong.Parse(&app{})
	kctx.FatalIfErrorf(kctx.Run())
}

func (a *app) Run() error {
	mdFiles, err := a.copy()
	if err != nil {
		return err
	}
	return a.genHTMLFiles(mdFiles)
}

// Copy the contents of the `src` directory to the `dest` directory.
// Skip over *.md files as we will generate *.html files from them.
func (a *app) copy() ([]string, error) {
	mdFiles := []string{}
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

		if d.IsDir() {
			// use MkdirAll in case the directory already exists
			return os.MkdirAll(destfile, 0o777) //nolint:gosec
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

func (a *app) genHTMLFiles(mdFiles []string) error {
	for _, mdf := range mdFiles {
		mdFile := filepath.Join(a.SrcDir, mdf)
		htmlFile := filepath.Join(a.DestDir, htmlFilename(mdf))
		root := toRoot(mdf)
		err := genHTMLFile(mdFile, htmlFile, root)
		if err != nil {
			return err
		}
	}
	return nil
}

var tmpl = template.Must(template.ParseFS(tmplFS, "tmpl/docs.html.tmpl"))

type tmplData struct {
	Root    string
	Title   string
	Content string
}

func genHTMLFile(mdFile, htmlFile, root string) error {
	mdBytes, err := os.ReadFile(mdFile)
	if err != nil {
		return err
	}
	title, htmlContent := md2html(mdBytes)
	out, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	data := tmplData{
		Root:    root,
		Title:   title,
		Content: htmlContent,
	}
	if err := tmpl.Execute(out, data); err != nil {
		out.Close() //nolint:errcheck,gosec // we're returning the more important error
		return err
	}
	return out.Close()
}

// md2html converts markdown to HTML and returns the title and HTML.
func md2html(mdBytes []byte) (string, string) {
	p := markdown.Parser{
		AutoLinkText: true, // turn URLs into links even without []()
	}
	doc := p.Parse(string(mdBytes))
	walk(doc, walkFn)
	title := extractTitle(doc)
	return title, markdown.ToHTML(doc)
}

func walkFn(n node) node {
	if mdl, ok := n.(*markdown.Link); ok {
		return updateLink(mdl)
	}
	return n
}

func updateLink(mdl *markdown.Link) node {
	u, err := url.Parse(mdl.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing URL %q: %v\n", mdl.URL, err)
		return nil
	}
	if u.IsAbs() {
		if rootDir, found := strings.CutSuffix(u.Hostname(), ".evy.dev"); found { //  subdomain link
			mdl.URL, err = url.JoinPath("/", rootDir, u.Path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error creating URL %q %q %q: %v\n", "/", rootDir, u.Path, err)
				return nil
			}
		}
		return mdl
	}
	// relative path, fix *.md filenames
	u.Path = htmlFilename(u.Path)
	mdl.URL = u.String()
	return mdl
}

func extractTitle(doc *markdown.Document) string {
	level := 100
	var titleText *markdown.Text
	for _, block := range doc.Blocks {
		if h, ok := block.(*markdown.Heading); ok {
			if h.Level < level {
				level = h.Level
				titleText = h.Text
			}
		}
	}
	if titleText == nil {
		return ""
	}
	buf := &bytes.Buffer{}
	for _, inline := range titleText.Inline {
		inline.PrintText(buf)
	}
	return buf.String()
}

func htmlFilename(mdf string) string {
	if filepath.Base(mdf) == "README.md" {
		return filepath.Join(filepath.Dir(mdf), "index.html")
	}
	if filename, found := strings.CutSuffix(mdf, ".md"); found {
		return filename + ".html"
	}
	return mdf
}

func toRoot(p string) string {
	if c := strings.Count(p, string(os.PathSeparator)); c > 0 {
		return strings.Repeat("/..", c)[1:]
	}
	return "."
}
