package md

import (
	"net/url"

	"rsc.io/markdown"
)

// RewriteLink replace a relative link to a .md file the HTML filename
// equivalent, see [HTMLFilename].
func RewriteLink(n Node) {
	mdl, ok := n.(*markdown.Link)
	if !ok {
		return
	}
	u, err := url.Parse(mdl.URL)
	if err != nil || u.IsAbs() {
		return
	}
	// relative path, fix *.md filenames
	u.Path = HTMLFilename(u.Path)
	mdl.URL = u.String()
}
