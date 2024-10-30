package evaluator

import (
	"fmt"
	"time"
)

// The Runtime interface must be implemented by an environment in order
// to execute Evy source code and Evy builtins. To create a new runtime
// implementation, you can embed [UnimplementedRuntime] and override
// the methods that your runtime can provide. For example, there is a
// jsRuntime implementation for the browser, which has full support for
// all graphics built-in functions. There is also a command-line
// environment implementation, which does not have graphics function
// support. For more details on the built-in functions, see the
// [built-ins documentation].
//
// [built-ins documentation]: https://github.com/evylang/evy/blob/main/docs/builtins.md
type Runtime interface {
	GraphicsRuntime
	Print(string)
	Read() string
	Cls()
	Sleep(dur time.Duration)
	Yielder() Yielder
}

// The GraphicsRuntime interface contains all methods that are required
// by the graphics built-ins. For more details see the
// [graphics built-ins] documentation.
//
// [graphics built-ins]: https://github.com/evylang/evy/blob/main/docs/builtins.md#graphics
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
	Transform(a, b, c, d, e, f float64)
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

// UnimplementedRuntime implements Runtime with no-ops and prints a "<func> not implemented" message.
type UnimplementedRuntime struct {
	print func(string)
}

// Print prints to os.Stdout.
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

// Cls is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Cls() { rt.unimplemented("cls") }

// Read is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Read() string { rt.unimplemented("read"); return "" }

// Sleep is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Sleep(time.Duration) { rt.unimplemented("sleep") }

// Yielder is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Yielder() Yielder { rt.unimplemented("yielder"); return nil }

// Move is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Move(float64, float64) { rt.unimplemented("move") }

// Line is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Line(float64, float64) { rt.unimplemented("line") }

// Rect is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Rect(float64, float64) { rt.unimplemented("rect") }

// Circle is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Circle(float64) { rt.unimplemented("circle") }

// Width is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Width(float64) { rt.unimplemented("width") }

// Color is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Color(string) { rt.unimplemented("color") }

// Clear is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Clear(string) { rt.unimplemented("clear") }

// Gridn is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Gridn(float64, string) { rt.unimplemented("gridn") }

// Poly is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Poly([][]float64) { rt.unimplemented("poly") }

// Stroke is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Stroke(string) { rt.unimplemented("stroke") }

// Fill is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Fill(string) { rt.unimplemented("fill") }

// Dash is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Dash([]float64) { rt.unimplemented("dash") }

// Linecap is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Linecap(string) { rt.unimplemented("linecap") }

// Text is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Text(string) { rt.unimplemented("text") }

// Font is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Font(map[string]any) { rt.unimplemented("font") }

// Ellipse is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Ellipse(float64, float64, float64, float64, float64, float64, float64) {
	rt.unimplemented("ellipse")
}

// Transform is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedRuntime) Transform(float64, float64, float64, float64, float64, float64) {
	rt.unimplemented("transform")
}
