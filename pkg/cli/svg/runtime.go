//go:build !tinygo

package svg

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

const (
	evyWidth    = 100
	evyHeight   = 100
	scaleFactor = 10
)

var (
	defaultStrokeWidth = 1.0
	defaultAttr        = Attr{
		Fill:            "black",
		Stroke:          "black",
		StrokeWidth:     &defaultStrokeWidth,
		StrokeLinecap:   "round",
		StrokeDashArray: "",
	}
)

var (
	defaultFontSize   = 60.0
	defaultFontWeight = 400.0
	defaultTextAttr   = TextAttr{
		TextAnchor:    "start",
		Baseline:      "alphabetic",
		FontSize:      &defaultFontSize,
		FontWeight:    &defaultFontWeight,
		FontStyle:     "normal",
		FontFamily:    `"Fira Code", monospace`,
		LetterSpacing: "0",
	}
)

// GraphicsRuntime implements evaluator.GraphcisRuntime for SVG output.
type GraphicsRuntime struct {
	x float64 // current cursor position
	y float64 // current cursor position

	attr     Attr
	textAttr TextAttr
	SVG      SVG
	elements []any
}

// NewGraphicsRuntime returns a new GraphicsRuntime with default attributes
// set suitable for Evy drawing output.
func NewGraphicsRuntime() *GraphicsRuntime {
	viewBox := fmt.Sprintf("0 0 %d %d", evyWidth*scaleFactor, evyHeight*scaleFactor)
	rt := &GraphicsRuntime{
		attr:     defaultAttr,
		textAttr: defaultTextAttr,
		SVG: SVG{
			XMLNS:   "http://www.w3.org/2000/svg",
			ViewBox: viewBox,
			Attr: Attr{
				StrokeLinecap: defaultAttr.StrokeLinecap,
				Stroke:        "black",
			},
			TextAttr: TextAttr{
				FontSize: defaultTextAttr.FontSize,
			},
		},
	}
	rt.x = rt.transformX(0)
	rt.y = rt.transformY(0)
	rt.Clear("white")
	return rt
}

func (rt *GraphicsRuntime) transformX(x float64) float64 {
	return rt.scale(x)
}

// transformY flips the y axis. Evy operates in a Cartesian number plain,
// SVG has an inverted y-axis like most computer graphics. We cannot
// directly use SVG `transform` as it would turn all text upside down.
func (rt *GraphicsRuntime) transformY(y float64) float64 {
	return rt.scale(evyHeight) - rt.scale(y)
}

func (rt *GraphicsRuntime) scale(s float64) float64 {
	return scaleFactor * s
}

func (rt *GraphicsRuntime) nonDefaultAttr() Attr {
	attr := rt.attr
	if rt.attr.Fill == defaultAttr.Fill {
		attr.Fill = ""
	}
	if rt.attr.Stroke == defaultAttr.Stroke {
		attr.Stroke = ""
	}
	if rt.attr.StrokeWidth != nil && *rt.attr.StrokeWidth == *defaultAttr.StrokeWidth {
		attr.StrokeWidth = nil
	}
	if rt.attr.StrokeLinecap == defaultAttr.StrokeLinecap {
		attr.StrokeLinecap = ""
	}
	if rt.attr.StrokeDashArray == defaultAttr.StrokeDashArray {
		attr.StrokeDashArray = ""
	}
	return attr
}

func (rt *GraphicsRuntime) nonDefaultTextAttr() TextAttr {
	textAttr := rt.textAttr
	if rt.textAttr.TextAnchor == defaultTextAttr.TextAnchor {
		textAttr.TextAnchor = ""
	}
	if rt.textAttr.Baseline == defaultTextAttr.Baseline {
		textAttr.Baseline = ""
	}
	if rt.textAttr.FontWeight != nil && *rt.textAttr.FontWeight == *defaultTextAttr.FontWeight {
		textAttr.FontWeight = nil
	}
	if rt.textAttr.FontStyle == defaultTextAttr.FontStyle {
		textAttr.FontStyle = ""
	}
	if rt.textAttr.FontSize != nil && *rt.textAttr.FontSize == *defaultTextAttr.FontSize {
		textAttr.FontSize = nil
	}
	if rt.textAttr.FontFamily == defaultTextAttr.FontFamily {
		textAttr.FontFamily = ""
	}
	if rt.textAttr.LetterSpacing == defaultTextAttr.LetterSpacing {
		textAttr.LetterSpacing = ""
	}
	return textAttr
}

// Push combines all previously collected SVG elements and adds them to the
// top-level SVG element. It is typically called when styling attributes
// change (like fill or stroke color) or when the Evy program execution ends.
//
// If there are multiple collected elements, they are wrapped in a group and
// the styles are applied to the group element. If there is only one element,
// the styles are applied directly to that element.
func (rt *GraphicsRuntime) Push() {
	if len(rt.elements) == 0 {
		return
	}
	var el any
	if len(rt.elements) == 1 {
		el = rt.elements[0]
	} else {
		el = &Group{Elements: rt.elements}
	}
	if rt.attr != defaultAttr {
		el.(attrSetter).setAttr(rt.nonDefaultAttr())
	}
	if at, ok := el.(textAttrSetter); ok {
		at.setTextAttr(rt.nonDefaultTextAttr())
	}

	rt.SVG.Elements = append(rt.SVG.Elements, el)
	rt.elements = nil
}

// Move sets the current cursor position.
func (rt *GraphicsRuntime) Move(x, y float64) {
	rt.x = rt.transformX(x)
	rt.y = rt.transformY(y)
}

// Line draws a line from the current cursor position to the given x, y.
func (rt *GraphicsRuntime) Line(x, y float64) {
	x = rt.transformX(x)
	y = rt.transformY(y)
	line := Line{X1: rt.x, Y1: rt.y, X2: x, Y2: y}
	rt.x = x
	rt.y = y
	rt.elements = append(rt.elements, &line)
}

// Rect draws a rectangle from the current cursor position for given width and
// height. Negative values are permitted.
func (rt *GraphicsRuntime) Rect(width, height float64) {
	x := rt.x
	y := rt.y
	width = rt.scale(width)
	height = -rt.scale(height)
	rt.x += width
	rt.y += height
	rect := Rect{
		X:      min(x, rt.x),
		Y:      min(y, rt.y),
		Width:  ftoa(math.Abs(width)),
		Height: ftoa(math.Abs(height)),
	}
	rt.elements = append(rt.elements, &rect)
}

// Circle draws a circle at the current cursor position with the given radius.
func (rt *GraphicsRuntime) Circle(radius float64) {
	radius = rt.scale(radius)
	circle := Circle{CX: rt.x, CY: rt.y, R: radius}
	rt.elements = append(rt.elements, &circle)
}

// Clear sets the background color of the SVG canvas. We cannot simply remove
// all previous SVG elements as the background color could be
// semi-transparent overlaying previous elements.
func (rt *GraphicsRuntime) Clear(color string) {
	if color == "" {
		color = "white"
	}
	rect := Rect{
		X:      0,
		Y:      0,
		Width:  "100%",
		Height: "100%",
		Attr: Attr{
			Fill:   color,
			Stroke: color,
		},
	}
	rt.elements = append(rt.elements, &rect)
}

// Poly draws a polygon or polyline with the given vertices.
func (rt *GraphicsRuntime) Poly(vertices [][]float64) {
	points := make([]string, len(vertices))
	for i, v := range vertices {
		x := rt.transformX(v[0])
		y := rt.transformY(v[1])
		points[i] = ftoa(x) + "," + ftoa(y)
	}
	poly := Polyline{
		Points: strings.Join(points, " "),
	}
	rt.elements = append(rt.elements, &poly)
}

// Ellipse draws an ellipse at the given x, y with the given radii.
// Note: startAngle, endAngle are not implemented.
func (rt *GraphicsRuntime) Ellipse(x, y, radiusX, radiusY, rotation, _, _ float64) {
	// TODO: implement the last two parameters: startAngle, endAngle.
	x = rt.transformX(x)
	y = rt.transformX(y)
	var transform string
	if rotation != 0 {
		transform = fmt.Sprintf("rotate(%f %f %f)", rotation, x, y)
	}
	ellipse := Ellipse{
		CX:        x,
		CY:        y,
		RX:        rt.scale(radiusX),
		RY:        rt.scale(radiusY),
		Transform: transform,
	}
	rt.elements = append(rt.elements, &ellipse)
}

// Text draws a text at the current cursor position.
func (rt *GraphicsRuntime) Text(str string) {
	text := Text{
		X:     rt.x,
		Y:     rt.y,
		Value: str,
	}
	if rt.attr.Fill != rt.attr.Stroke {
		text.Fill = rt.attr.Stroke
	}
	rt.elements = append(rt.elements, &text)
}

// Gridn draws a grid with the given unit and color.
func (rt *GraphicsRuntime) Gridn(unit float64, color string) {
	unit = rt.transformX(unit)
	group := Group{Attr: Attr{Stroke: color}}
	lineCnt := 0
	thickWdith := 2.0
	height := float64(evyHeight * scaleFactor)
	width := float64(evyWidth * scaleFactor)
	for i := float64(0); i <= 1000; i += unit {
		hLine := &Line{X1: i, Y1: 0, X2: i, Y2: height}
		vLine := &Line{X1: 0, Y1: i, X2: width, Y2: i}
		if lineCnt%5 == 0 {
			hLine.StrokeWidth = &thickWdith
			vLine.StrokeWidth = &thickWdith
		}
		lineCnt++
		group.Elements = append(group.Elements, hLine, vLine)
	}
	rt.elements = append(rt.elements, &group)
}

// Transform sets the transform for all successive shapes.
func (rt *GraphicsRuntime) Transform(a, b, c, d, e, f float64) {
	rt.Push()
	rt.attr.Transform = fmt.Sprintf("matrix(%f %f %f %f %f %f)", a, b, c, d, e, f)
}

// Width sets the stroke width.
func (rt *GraphicsRuntime) Width(w float64) {
	rt.Push()
	strokeWidth := rt.scale(w)
	rt.attr.StrokeWidth = &strokeWidth
}

// Color sets the stroke and fill color.
func (rt *GraphicsRuntime) Color(color string) {
	rt.Push()
	rt.attr.Stroke = color
	rt.attr.Fill = color
}

// Stroke sets the stroke color only.
func (rt *GraphicsRuntime) Stroke(str string) {
	rt.Push()
	rt.attr.Stroke = str
}

// Fill sets the fill color only.
func (rt *GraphicsRuntime) Fill(str string) {
	rt.Push()
	rt.attr.Fill = str
}

// Dash sets the stroke dash array.
func (rt *GraphicsRuntime) Dash(segments []float64) {
	rt.Push()
	segmentStrings := make([]string, len(segments))
	for i, segment := range segments {
		segmentStrings[i] = ftoa(rt.scale(segment))
	}
	rt.attr.StrokeDashArray = strings.Join(segmentStrings, " ")
}

// Linecap sets the stroke linecap style.
func (rt *GraphicsRuntime) Linecap(str string) {
	rt.Push()
	rt.attr.StrokeLinecap = str
}

// Font sets the text font properties. The following property keys with sample
// values are supported:
//
//	"family": "Georgia, serif",
//	"size": 3, // relative to canvas, numbers only no "12px" etc.
//	"weight": 100, //| 200| 300 | 400 == "normal" | 500 | 600 | 700 == "bold" | 800 | 900
//	"style": "italic", | "oblique 35deg" | "normal"
//	"baseline": "top", // | "middle" | "bottom" | "alphabetic"
//	"align": "left", // | "center" | "right"
//	"letterspacing": 1 // number, see size. extra inter-character space. negative allowed.
func (rt *GraphicsRuntime) Font(props map[string]any) {
	rt.Push()

	if family, ok := props["family"].(string); ok {
		rt.textAttr.FontFamily = family
	}
	if size, ok := props["size"].(float64); ok {
		fontSize := rt.scale(size)
		rt.textAttr.FontSize = &fontSize
	}
	if weight, ok := props["weight"].(float64); ok {
		rt.textAttr.FontWeight = &weight
	}
	if style, ok := props["style"].(string); ok {
		rt.textAttr.FontStyle = style
	}
	if baseline, ok := props["baseline"].(string); ok {
		switch baseline {
		case "top":
			rt.textAttr.Baseline = "hanging"
		case "middle":
			rt.textAttr.Baseline = "middle"
		case "bottom":
			rt.textAttr.Baseline = "ideographic"
		case "alphabetic":
			rt.textAttr.Baseline = "alphabetic"
		}
		rt.textAttr.Baseline = baseline
	}
	if align, ok := props["align"].(string); ok {
		switch align {
		case "left":
			rt.textAttr.TextAnchor = "start"
		case "right":
			rt.textAttr.TextAnchor = "end"
		case "center":
			rt.textAttr.TextAnchor = "middle"
		}
	}
	if letterSpacing, ok := props["letterspacing"].(float64); ok {
		rt.textAttr.LetterSpacing = ftoa(letterSpacing)
	}
}

// WriteSVG writes the SVG output to the given writer.
func (rt *GraphicsRuntime) WriteSVG(w io.Writer) error {
	rt.Push()
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(rt.SVG); err != nil {
		return fmt.Errorf("cannot encode SVG: %w", err)
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("cannot add newline to SVG output: %w", err)
	}
	return nil
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
