package evaluator

import (
	"fmt"
	"time"
)

type Runtime interface {
	GraphicsRuntime
	Print(string)
	Read() string
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
	Textsize(size float64)
	Font(s string)
	Fontfamily(s string)
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

func (rt *UnimplementedRuntime) Unimplemented(s string) {
	rt.Print(fmt.Sprintf("%q not implemented\n", s))
}

func (rt *UnimplementedRuntime) Read() string              { rt.Unimplemented("read"); return "" }
func (rt *UnimplementedRuntime) Sleep(_ time.Duration)     { rt.Unimplemented("sleep") }
func (rt *UnimplementedRuntime) Yielder() Yielder          { rt.Unimplemented("yielder"); return nil }
func (rt *UnimplementedRuntime) Move(x, y float64)         { rt.Unimplemented("move") }
func (rt *UnimplementedRuntime) Line(x, y float64)         { rt.Unimplemented("line") }
func (rt *UnimplementedRuntime) Rect(x, y float64)         { rt.Unimplemented("rect") }
func (rt *UnimplementedRuntime) Circle(r float64)          { rt.Unimplemented("circle") }
func (rt *UnimplementedRuntime) Width(w float64)           { rt.Unimplemented("width") }
func (rt *UnimplementedRuntime) Color(s string)            { rt.Unimplemented("color") }
func (rt *UnimplementedRuntime) Clear(color string)        { rt.Unimplemented("clear") }
func (rt *UnimplementedRuntime) Poly(vertices [][]float64) { rt.Unimplemented("poly") }
func (rt *UnimplementedRuntime) Stroke(s string)           { rt.Unimplemented("stroke") }
func (rt *UnimplementedRuntime) Fill(s string)             { rt.Unimplemented("fill") }
func (rt *UnimplementedRuntime) Dash(segments []float64)   { rt.Unimplemented("dash") }
func (rt *UnimplementedRuntime) Linecap(s string)          { rt.Unimplemented("linecap") }
func (rt *UnimplementedRuntime) Text(s string)             { rt.Unimplemented("text") }
func (rt *UnimplementedRuntime) Textsize(size float64)     { rt.Unimplemented("textsize") }
func (rt *UnimplementedRuntime) Font(s string)             { rt.Unimplemented("font") }
func (rt *UnimplementedRuntime) Fontfamily(s string)       { rt.Unimplemented("fontfamily") }
func (rt *UnimplementedRuntime) Ellipse(x, y, rX, rY, rotation, startAngle, endAngle float64) {
	rt.Unimplemented("ellipse")
}
