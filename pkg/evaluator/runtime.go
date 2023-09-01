package evaluator

import (
	"fmt"
	"time"
)

type Runtime interface {
	GraphicsRuntime
	Print(string)
	Read() string
	Cls()
	Sleep(dur time.Duration)
	Yielder() Yielder
}

type GraphicsRuntime interface {
	Move(x, y float64)
	Line(x, y float64)
	Rect(dx, dy float64)
	Circle(radius float64)
	Width(w float64)
	Color(s string)
	Clear(color string)

	// advanced graphics functions
	Poly(vertices [][]float64)
	Ellipse(x, y, radiusX, radiusY, rotation, startAngle, endAngle float64)
	Stroke(s string)
	Fill(s string)
	Dash(segments []float64)
	Linecap(s string)
	Text(s string)
	// font optionally sets font properties such as family, size or weight.
	// font properties match their CSS properties. Here's an exhaustive list
	// of mapped properties:
	//
	//		font {
	//			family: "Georgia, serif"
	//			size: 3 // relative to canvas, numbers only no "12px" etc.
	//			weight: 100 | 200| 300 | 400 == "normal" | 500 | 600 | 700 == "bold" | 800 | 900
	//			style: "italic" | "oblique 35deg" | "normal"
	//			baseline: "top" | "middle" | "bottom"
	//			align: "left" | "center" | "right"
	//			letterspacing: 1 // number, see size. extra inter-character space. negative allowed.
	//		}
	Font(props map[string]any)
	Gridn(unit float64, color string)
}

type UnimplementedRuntime struct {
	print func(string)
}

func (rt *UnimplementedRuntime) Print(s string) {
	if rt.print != nil {
		rt.print(s)
	} else {
		print(s)
	}
}

func (rt *UnimplementedRuntime) unimplemented(s string) {
	rt.Print(fmt.Sprintf("%q not implemented\n", s))
}

func (rt *UnimplementedRuntime) Cls()                  { rt.unimplemented("cls") }
func (rt *UnimplementedRuntime) Read() string          { rt.unimplemented("read"); return "" }
func (rt *UnimplementedRuntime) Sleep(time.Duration)   { rt.unimplemented("sleep") }
func (rt *UnimplementedRuntime) Yielder() Yielder      { rt.unimplemented("yielder"); return nil }
func (rt *UnimplementedRuntime) Move(float64, float64) { rt.unimplemented("move") }
func (rt *UnimplementedRuntime) Line(float64, float64) { rt.unimplemented("line") }
func (rt *UnimplementedRuntime) Rect(float64, float64) { rt.unimplemented("rect") }
func (rt *UnimplementedRuntime) Circle(float64)        { rt.unimplemented("circle") }
func (rt *UnimplementedRuntime) Width(float64)         { rt.unimplemented("width") }
func (rt *UnimplementedRuntime) Color(string)          { rt.unimplemented("color") }
func (rt *UnimplementedRuntime) Clear(string)          { rt.unimplemented("clear") }
func (rt *UnimplementedRuntime) Gridn(float64, string) { rt.unimplemented("gridn") }
func (rt *UnimplementedRuntime) Poly([][]float64)      { rt.unimplemented("poly") }
func (rt *UnimplementedRuntime) Stroke(string)         { rt.unimplemented("stroke") }
func (rt *UnimplementedRuntime) Fill(string)           { rt.unimplemented("fill") }
func (rt *UnimplementedRuntime) Dash([]float64)        { rt.unimplemented("dash") }
func (rt *UnimplementedRuntime) Linecap(string)        { rt.unimplemented("linecap") }
func (rt *UnimplementedRuntime) Text(string)           { rt.unimplemented("text") }
func (rt *UnimplementedRuntime) Font(map[string]any)   { rt.unimplemented("font") }
func (rt *UnimplementedRuntime) Ellipse(float64, float64, float64, float64, float64, float64, float64) {
	rt.unimplemented("ellipse")
}
