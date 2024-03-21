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
	"regexp"
	"slices"
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
	asts, err := a.makeASTs(mdFiles)
	if err != nil {
		return err
	}
	// Add anchors, fill in sidebar, replace .md with .html in links
	updateASTs(asts)
	return a.genHTMLFiles(asts)
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

// makeASTs creates *markdown.Document ASTs from the markdown files in
// `mdFiles` and returns them in a map for easier lookup.
func (a *app) makeASTs(mdFiles []string) (map[string]*markdown.Document, error) {
	p := markdown.Parser{
		AutoLinkText: true, // turn URLs into links even without []()
		Table:        true,
	}
	asts := map[string]*markdown.Document{}
	for _, mdf := range mdFiles {
		mdFile := filepath.Join(a.SrcDir, mdf)
		mdBytes, err := os.ReadFile(mdFile)
		if err != nil {
			return nil, err
		}
		asts[mdf] = p.Parse(string(mdBytes))
	}
	return asts, nil
}

// updateASTs changes the markdown asts to
//
// - add anchors to headings
// - re-write relative links to .md files to .html files
// - expands sidebar entries with sub-headings.
func updateASTs(asts map[string]*markdown.Document) {
	headings := map[string][]heading{}
	var sidebarFiles []string
	for mdf, ast := range asts {
		if filepath.Base(mdf) == "_sidebar.md" {
			sidebarFiles = append(sidebarFiles, mdf)
			continue
		}
		w := &walker{anchorIDs: map[string]bool{}}
		walk(ast, w.walk)
		headings[mdf] = w.headings
	}
	for _, sbf := range sidebarFiles {
		updateSidebar(sbf, asts[sbf], headings)
		w := &walker{anchorIDs: map[string]bool{}}
		// we need to walk sidebars _after_ sidebar update with heading
		// insertion because we look up the inserted headings by markdown and
		// not html filename.
		walk(asts[sbf], w.walk)
	}
}

func (a *app) genHTMLFiles(asts map[string]*markdown.Document) error {
	for mdf, doc := range asts {
		if filepath.Base(mdf) == "_sidebar.md" || filepath.Base(mdf) == "_header.md" {
			continue
		}
		sidebar := filepath.Join(filepath.Dir(mdf), "_sidebar.md")
		header := filepath.Join(filepath.Dir(mdf), "_header.md")
		htmlFile := filepath.Join(a.DestDir, htmlFilename(mdf))
		err := genHTMLFile(doc, asts[sidebar], asts[header], htmlFile, mdf)
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
	Sidebar string
	Header  string
}

func genHTMLFile(doc, sidebar, header *markdown.Document, htmlFile, mdf string) error {
	data := tmplData{
		Root:    toRoot(mdf),
		Title:   extractTitle(doc),
		Content: docToHTML(doc),
		Sidebar: docToHTML(sidebar),
		Header:  docToHTML(unwrapParagraph(header)),
	}

	out, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(out, data); err != nil {
		out.Close() //nolint:errcheck,gosec // we're returning the more important error
		return err
	}
	return out.Close()
}

func docToHTML(doc *markdown.Document) string {
	if doc == nil {
		return ""
	}
	return markdown.ToHTML(doc)
}

func unwrapParagraph(doc *markdown.Document) *markdown.Document {
	if doc == nil {
		return nil
	}
	p := doc.Blocks[0].(*markdown.Paragraph)
	return &markdown.Document{
		Blocks: []markdown.Block{p.Text},
	}
}

type walker struct {
	anchorIDs map[string]bool
	headings  []heading
}

type heading struct {
	anchorID string
	heading  *markdown.Heading
}

func (w *walker) walk(n node) {
	switch n := n.(type) {
	case *markdown.Document:
		removeTOC(n)
	case *markdown.Link:
		updateLink(n)
	case *markdown.Heading:
		w.updateHeading(n)
	}
}

func removeTOC(doc *markdown.Document) {
	inTOC := false
	var blocks []markdown.Block
	for i, b := range doc.Blocks {
		if h, ok := b.(*markdown.Heading); ok {
			htext := strings.ToLower(strings.TrimSpace(markdown.ToMarkdown(h.Text)))
			inTOC = htext == "table of contents"
		}
		if !inTOC {
			blocks = append(blocks, doc.Blocks[i])
		}
	}
	doc.Blocks = blocks
}

func updateLink(mdl *markdown.Link) {
	u, err := url.Parse(mdl.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing URL %q: %v\n", mdl.URL, err)
		return
	}
	if u.IsAbs() {
		if rootDir, found := strings.CutSuffix(u.Hostname(), ".evy.dev"); found { //  subdomain link
			u.Path, err = url.JoinPath("/", rootDir, u.Path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error creating URL %q %q %q: %v\n", "/", rootDir, u.Path, err)
			}
			u.Host = ""
			u.Scheme = ""
			mdl.URL = u.String()
		}
		return
	}
	// relative path, fix *.md filenames
	u.Path = htmlFilename(u.Path)
	mdl.URL = u.String()
}

// updateHeading inserts a [markdown.Inline] element at the start of a
// [markdown.Heading]s Text slice that renders a link marker linking
// to the heading, allowing for easily copying links to a heading.
func (w *walker) updateHeading(h *markdown.Heading) {
	if h.Level == 1 || h.Level > 3 {
		return
	}
	text := inlineText(h.Text.Inline)
	var majorHeading string
	id := makeID(text, majorHeading, w.anchorIDs)
	anchor := markdown.Inline(newAnchor(id))
	h.Text.Inline = slices.Insert(h.Text.Inline, 0, anchor)
	ah := heading{anchorID: id, heading: h}
	w.headings = append(w.headings, ah)
}

func newAnchor(id string) *markdown.HTMLTag {
	format := `<a id="%s" href="#%s" class="anchor">#</a>`
	return &markdown.HTMLTag{
		Text: fmt.Sprintf(format, id, id),
	}
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
	return inlineText(titleText.Inline)
}

func inlineText(inlines []markdown.Inline) string {
	buf := &bytes.Buffer{}
	for _, inline := range inlines {
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

var reHeadingID = regexp.MustCompile(`[^\pL\pN]+`)

func makeID(s, majorHeading string, ids map[string]bool) string {
	id := strings.ToLower(s)
	id = reHeadingID.ReplaceAllString(id, "-")
	if ids[id] && majorHeading != "" {
		id = majorHeading + "-" + id
	}
	orig := id
	for i := 1; ids[id]; i++ {
		id = fmt.Sprintf("%s-%d", orig, i)
	}
	ids[id] = true
	return id
}
