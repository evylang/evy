//go:build !tinygo

package svg

import (
	"bytes"
	"os"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestRuntimeCircleRect(t *testing.T) {
	rt := newTestRuntime()
	rt.Move(20, 0)
	rt.Rect(10, 30)
	rt.Rect(20, 5)
	rt.Move(50, 50)
	rt.Color("red")
	rt.Circle(10)
	rt.Gridn(10, "hsl(0deg 100% 0% / 50%)")

	assertSVG(t, "testdata/circle-rect.svg", rt)
}

func TestRuntimeEllipse(t *testing.T) {
	rt := newTestRuntime()
	rt.Ellipse(50, 85, 30, 10, 0, 0, 0)
	rt.Ellipse(50, 55, 30, 10, 30, 0, 0)

	assertSVG(t, "testdata/ellipse.svg", rt)
}

func TestRuntimeFill(t *testing.T) {
	rt := newTestRuntime()
	rt.Width(2)
	rt.Move(10, 65)
	rt.Color("red")
	rt.Circle(7)
	rt.Move(3, 40)
	rt.Rect(14, 14)
	rt.Move(3, 35)
	rt.Line(17, 35)

	rt.Move(30, 65)
	rt.Stroke("blue")
	rt.Circle(7)
	rt.Move(23, 40)
	rt.Rect(14, 14)
	rt.Move(23, 35)
	rt.Line(37, 35)

	rt.Move(50, 65)
	rt.Color("green")
	rt.Fill("orange")
	rt.Circle(7)
	rt.Move(43, 40)
	rt.Rect(14, 14)
	rt.Move(43, 35)
	rt.Line(57, 35)

	rt.Move(70, 65)
	rt.Stroke("deeppink")
	rt.Fill("cyan")
	rt.Circle(7)
	rt.Move(63, 40)
	rt.Rect(14, 14)
	rt.Move(63, 35)
	rt.Line(77, 35)

	rt.Move(90, 65)
	rt.Stroke("violet")
	rt.Fill("none")
	rt.Circle(7)
	rt.Move(83, 40)
	rt.Rect(14, 14)
	rt.Move(83, 35)
	rt.Line(97, 35)

	assertSVG(t, "testdata/fill.svg", rt)
}

func TestRuntimeLines(t *testing.T) {
	rt := newTestRuntime()
	for i := float64(0); i < 100; i += 3 {
		rt.Move(i, 0)
		rt.Line(100, i)
	}
	assertSVG(t, "testdata/lines.svg", rt)
}

func TestRuntimeLinestyle(t *testing.T) {
	rt := newTestRuntime()

	rt.Width(3)
	rt.Linecap("round")
	rt.Move(5, 80)
	rt.Line(95, 80)

	rt.Linecap("butt")
	rt.Move(5, 70)
	rt.Line(95, 70)

	rt.Linecap("square")
	rt.Move(5, 60)
	rt.Line(95, 60)

	rt.Width(1)
	rt.Move(5, 30)
	rt.Dash([]float64{5, 3, 1, 3})
	rt.Line(95, 30)

	rt.Dash(nil)
	rt.Move(5, 20)
	rt.Line(95, 20)

	assertSVG(t, "testdata/linestyle.svg", rt)
}

func TestRuntimePoly(t *testing.T) {
	rt := newTestRuntime()
	rt.Width(1)
	rt.Color("red")
	rt.Fill("none")
	rt.Poly([][]float64{{10, 80}, {30, 60}, {50, 80}, {70, 60}, {90, 80}})
	rt.Fill("gold")
	rt.Poly([][]float64{{10, 20}, {50, 50}, {20, 10}, {10, 20}})

	assertSVG(t, "testdata/poly.svg", rt)
}

func TestRuntimeText(t *testing.T) {
	rt := newTestRuntime()

	rt.Move(10, 85)
	rt.Text("“Time is an illusion.")
	rt.Move(10, 78)
	rt.Text("Lunchtime doubly so.”")

	rt.Font(map[string]any{
		"size":          float64(4),
		"style":         "italic",
		"family":        "Tahomana, sans-serif",
		"weight":        float64(700),
		"letterspacing": float64(-0.1),
		"align":         "center",
		"baseline":      "middle",
	})
	rt.Move(60, 72)
	rt.Color("dodgerblue")
	rt.Text("― Douglas Adams")

	rt.Color("black")
	rt.Font(map[string]any{"size": float64(6), "style": "normal", "letterspacing": float64(0), "align": "left", "family": "Fira Code, monospace"})
	rt.Fill("none")

	rt.Move(10, 50)
	rt.Line(45, 50)
	rt.Move(10, 50)
	rt.Font(map[string]any{"baseline": "bottom"})
	rt.Text("bottom")

	rt.Move(10, 35)
	rt.Line(45, 35)
	rt.Move(10, 35)
	rt.Font(map[string]any{"baseline": "top"})
	rt.Text("top")
	rt.Move(10, 20)
	rt.Line(45, 20)
	rt.Move(10, 20)
	rt.Font(map[string]any{"baseline": "middle"})
	rt.Text("middle")
	rt.Move(10, 5)
	rt.Line(45, 5)
	rt.Move(10, 5)
	rt.Font(map[string]any{"baseline": "alphabetic"})
	rt.Text("alphabetic")
	rt.Move(70, 48)
	rt.Line(70, 56)
	rt.Move(70, 50)
	rt.Font(map[string]any{"align": "left"})
	rt.Text("left")
	rt.Move(70, 33)
	rt.Line(70, 41)
	rt.Move(70, 35)
	rt.Font(map[string]any{"align": "right"})
	rt.Text("right")
	rt.Move(70, 18)
	rt.Line(70, 26)
	rt.Move(70, 20)
	rt.Font(map[string]any{"align": "center"})
	rt.Text("center")

	assertSVG(t, "testdata/text.svg", rt)
}

func assertSVG(t *testing.T, wantFilename string, gotRT *GraphicsRuntime) {
	t.Helper()
	b, err := os.ReadFile(wantFilename)
	assert.NoError(t, err)
	want := string(b)

	buffer := &bytes.Buffer{}
	err = gotRT.WriteSVG(buffer)
	assert.NoError(t, err)
	got := buffer.String()

	assert.Equal(t, want, got)
}

const testStyle = "border: 1px solid red; width: 400px; height: 400px"

func newTestRuntime() *GraphicsRuntime {
	rt := NewGraphicsRuntime()
	rt.SVG.Style = testStyle
	return rt
}
