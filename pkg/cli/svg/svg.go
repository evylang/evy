//go:build !tinygo

// Package svg provides an Evy runtime to generate SVG output for evy programs
// that contain graphics function calls. The SVG elements modeled in this
// package, such as <svg>, <circle> or <g>, and their attributes, "fill"
// or "stroke-width", are a small subset of all SVG elements and attributes.
// They are minimum necessary to output evy drawings as SVG.
package svg

import "encoding/xml"

// SVG represents a top-level SVG element <svg>.
type SVG struct {
	XMLName xml.Name `xml:"svg"`
	Attr
	TextAttr
	Width    string `xml:"width,attr,omitempty"`
	Height   string `xml:"height,attr,omitempty"`
	ViewBox  string `xml:"viewBox,attr,omitempty"`
	Style    string `xml:"style,attr,omitempty"`
	XMLNS    string `xml:"xmlns,attr,omitempty"`
	Elements []any  // group, circle, rect, ...
}

// Attr represents the attributes of text and non-text SVG elements. It is
// embedded in other types representing SVG elements, such as Group or
// Circle.
type Attr struct {
	Fill            string `xml:"fill,attr,omitempty"`
	Stroke          string `xml:"stroke,attr,omitempty"`
	StrokeWidth     string `xml:"stroke-width,attr,omitempty"`
	StrokeLinecap   string `xml:"stroke-linecap,attr,omitempty"`
	StorkeDashArray string `xml:"stroke-dasharray,attr,omitempty"`
}

// TextAttr represents the attributes of text or group SVG elements and
// is embedded in Group and Text types.
type TextAttr struct {
	TextAnchor    string `xml:"text-anchor,attr,omitempty"`
	Baseline      string `xml:"dominant-baseline,attr,omitempty"`
	FontSize      string `xml:"font-size,attr,omitempty"`
	FontWeight    string `xml:"font-weight,attr,omitempty"`
	FontStyle     string `xml:"font-style,attr,omitempty"` // italic, normal
	FontFamily    string `xml:"font-family,attr,omitempty"`
	LetterSpacing string `xml:"letter-spacing,attr,omitempty"`
}

type (
	attrSetter     interface{ setAttr(a Attr) }
	textAttrSetter interface{ setTextAttr(ta TextAttr) }
)

// Group represents a group of SVG elements <g>.
type Group struct {
	XMLName struct{} `xml:"g"`
	Attr
	TextAttr
	Elements []any // circle, rect, ...
}

func (g *Group) setAttr(a Attr)          { g.Attr = a }
func (g *Group) setTextAttr(ta TextAttr) { g.TextAttr = ta }

// Line represents an SVG line element <line>.
type Line struct {
	XMLName struct{} `xml:"line"`
	Attr
	X1 string `xml:"x1,attr"`
	Y1 string `xml:"y1,attr"`
	X2 string `xml:"x2,attr"`
	Y2 string `xml:"y2,attr"`
}

func (l *Line) setAttr(a Attr) { l.Attr = a }

// Circle represents an SVG circle element <circle>.
type Circle struct {
	XMLName struct{} `xml:"circle"`
	Attr
	CX string `xml:"cx,attr"`
	CY string `xml:"cy,attr"`
	R  string `xml:"r,attr"`
}

func (c *Circle) setAttr(a Attr) { c.Attr = a }

// Rect represents an SVG rectangle element <rect>.
type Rect struct {
	XMLName struct{} `xml:"rect"`
	Attr
	X      string `xml:"x,attr"`
	Y      string `xml:"y,attr"`
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
}

func (r *Rect) setAttr(a Attr) { r.Attr = a }

// Polyline represents an SVG polyline element <polyline>.
type Polyline struct {
	XMLName struct{} `xml:"polyline"`
	Attr
	Points string `xml:"points,attr"`
}

func (p *Polyline) setAttr(a Attr) { p.Attr = a }

// Ellipse represents an SVG ellipse element <ellipse>.
type Ellipse struct {
	XMLName struct{} `xml:"ellipse"`
	Attr
	CX        string `xml:"cx,attr"`
	CY        string `xml:"cy,attr"`
	RX        string `xml:"rx,attr"`
	RY        string `xml:"ry,attr"`
	Transform string `xml:"transform,attr,omitempty"`
}

func (p *Ellipse) setAttr(a Attr) { p.Attr = a }

// Text represents an SVG text element <text>.
type Text struct {
	XMLName struct{} `xml:"text"`
	Attr
	TextAttr
	X     string `xml:"x,attr"`
	Y     string `xml:"y,attr"`
	Value string `xml:",chardata"`
}

func (t *Text) setAttr(a Attr) {
	t.Attr = a
	if t.Attr.Fill != t.Attr.Stroke {
		t.Attr.Fill = t.Attr.Stroke
	}
}
func (t *Text) setTextAttr(ta TextAttr) { t.TextAttr = ta }
