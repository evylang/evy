package evaluator

import (
	"fmt"
	"time"
)

// The Platform interface must be implemented by an environment in order
// to execute Evy source code and Evy builtins. To create a new platform
// implementation, you can embed [UnimplementedPlatform] and override
// the methods that your platform can provide. For example, there is a
// jsPlatform implementation for the browser, which has full support for
// all graphics built-in functions. There is also a command-line
// environment implementation, which does not have graphics function
// support. For more details on the built-in functions, see the
// [built-ins documentation].
//
// [built-ins documentation]: https://github.com/evylang/evy/blob/main/docs/builtins.md
type Platform interface {
	GraphicsPlatform
	Print(string)
	Read() string
	Cls()
	Sleep(dur time.Duration)
	Yielder() Yielder
}

// The GraphicsPlatform interface contains all methods that are required
// by the graphics built-ins. For more details see the
// [graphics built-ins] documentation.
//
// [graphics built-ins]: https://github.com/evylang/evy/blob/main/docs/builtins.md#graphics
type GraphicsPlatform interface {
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

// UnimplementedPlatform implements Platform with no-ops and prints a "<func> not implemented" message.
type UnimplementedPlatform struct {
	print func(string)
}

// Print prints to os.Stdout.
func (rt *UnimplementedPlatform) Print(s string) {
	if rt.print != nil {
		rt.print(s)
	} else {
		print(s)
	}
}

func (rt *UnimplementedPlatform) unimplemented(s string) {
	rt.Print(fmt.Sprintf("%q not implemented\n", s))
}

// Cls is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Cls() { rt.unimplemented("cls") }

// Read is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Read() string { rt.unimplemented("read"); return "" }

// Sleep is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Sleep(time.Duration) { rt.unimplemented("sleep") }

// Yielder is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Yielder() Yielder { rt.unimplemented("yielder"); return nil }

// Move is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Move(float64, float64) { rt.unimplemented("move") }

// Line is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Line(float64, float64) { rt.unimplemented("line") }

// Rect is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Rect(float64, float64) { rt.unimplemented("rect") }

// Circle is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Circle(float64) { rt.unimplemented("circle") }

// Width is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Width(float64) { rt.unimplemented("width") }

// Color is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Color(string) { rt.unimplemented("color") }

// Clear is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Clear(string) { rt.unimplemented("clear") }

// Gridn is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Gridn(float64, string) { rt.unimplemented("gridn") }

// Poly is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Poly([][]float64) { rt.unimplemented("poly") }

// Stroke is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Stroke(string) { rt.unimplemented("stroke") }

// Fill is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Fill(string) { rt.unimplemented("fill") }

// Dash is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Dash([]float64) { rt.unimplemented("dash") }

// Linecap is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Linecap(string) { rt.unimplemented("linecap") }

// Text is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Text(string) { rt.unimplemented("text") }

// Font is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Font(map[string]any) { rt.unimplemented("font") }

// Ellipse is a no-op that prints an "unimplemented" message.
func (rt *UnimplementedPlatform) Ellipse(float64, float64, float64, float64, float64, float64, float64) {
	rt.unimplemented("ellipse")
}
