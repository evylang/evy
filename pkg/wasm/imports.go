//go:build tinygo

package main

// This file contains JS functions imported into Go/WASM. Functions are
// declared without body, the full definition can be found in the JS
// implementation. The `jsRuntime` struct wraps these functions to
// implement evaluator.Runtime.

import (
	"fmt"
	"strings"
	"time"

	"evylang.dev/evy/pkg/evaluator"
)

const minSleepDur = time.Millisecond

// jsRuntime implements evaluator.Runtime.
type jsRuntime struct {
	yielder *sleepingYielder
}

func newJSRuntime() *jsRuntime {
	return &jsRuntime{yielder: newSleepingYielder()}
}

func (rt *jsRuntime) Yielder() evaluator.Yielder { return rt.yielder }
func (rt *jsRuntime) Print(s string)             { jsPrint(s) }
func (rt *jsRuntime) Cls()                       { jsCls() }
func (rt *jsRuntime) Read() string               { return rt.yielder.Read() }

func (rt *jsRuntime) Sleep(dur time.Duration) {
	// Enforce a lower bound to stop browser tabs from freezing.
	if dur < minSleepDur {
		dur = minSleepDur
	}
	rt.yielder.Sleep(dur)
}

func (rt *jsRuntime) Move(x, y float64)                { move(x, y) }
func (rt *jsRuntime) Line(x, y float64)                { line(x, y) }
func (rt *jsRuntime) Rect(x, y float64)                { rect(x, y) }
func (rt *jsRuntime) Circle(r float64)                 { circle(r) }
func (rt *jsRuntime) Width(w float64)                  { width(w) }
func (rt *jsRuntime) Color(s string)                   { color(s) }
func (rt *jsRuntime) Clear(color string)               { clear(color) }
func (rt *jsRuntime) Gridn(unit float64, color string) { gridn(unit, color) }
func (rt *jsRuntime) Stroke(s string)                  { stroke(s) }
func (rt *jsRuntime) Fill(s string)                    { fill(s) }
func (rt *jsRuntime) Dash(segments []float64)          { dash(floatsToString(segments)) }
func (rt *jsRuntime) Linecap(s string)                 { linecap(s) }
func (rt *jsRuntime) Text(s string)                    { text(s) }
func (rt *jsRuntime) Font(props map[string]any) {
	// We don't use encoding/json here as it adds more than 100K to evy.wasm.
	pairs := make([]string, 0, len(props))
	for key, value := range props {
		switch value.(type) {
		case string:
			pairs = append(pairs, fmt.Sprintf("%q:%q", key, value))
		default:
			pairs = append(pairs, fmt.Sprintf("%q:%v", key, value))
		}
	}
	jsonProps := "{" + strings.Join(pairs, ",") + "}"
	font(jsonProps)
}

func (rt *jsRuntime) Poly(vertices [][]float64) {
	vStrings := make([]string, len(vertices))
	for i, vertex := range vertices {
		vStrings[i] = fmt.Sprintf("%f %f", vertex[0], vertex[1])
	}
	poly(strings.Join(vStrings, " "))
}

func (rt *jsRuntime) Ellipse(x, y, rX, rY, rotation, startAngle, endAngle float64) {
	ellipse(x, y, rX, rY, rotation, startAngle, endAngle)
}

func floatsToString(floats []float64) string {
	if len(floats) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, f := range floats {
		sb.WriteString(fmt.Sprintf(" %f", f))
	}
	return sb.String()[1:]
}

// sleepingYielder yields the CPU so that JavaScript/browser events
// get a chance to be processed. Currently(Feb 2023) it seems that you
// can only yield to JS by sleeping for at least 1ms but having that
// delay is not ideal. Other methods of yielding can be explored by
// implementing a different yield function.
type sleepingYielder struct {
	start time.Time
	count int
}

func newSleepingYielder() *sleepingYielder {
	return &sleepingYielder{start: time.Now()}
}

func (y *sleepingYielder) Yield() {
	y.count++
	if y.count > 1000 && time.Since(y.start) > 100*time.Millisecond {
		y.ForceYield()
	}
}

func (y *sleepingYielder) ForceYield() {
	y.Sleep(minSleepDur)
}

func (y *sleepingYielder) Sleep(dur time.Duration) {
	time.Sleep(dur)
	y.Reset()
}

func (y *sleepingYielder) Read() string {
	for {
		if eval.Stopped {
			return ""
		}
		addr := jsRead()
		if addr != 0 {
			return getStringFromAddr(addr)
		}
		y.Sleep(50 * time.Millisecond)
	}
}

func (y *sleepingYielder) Reset() {
	y.start = time.Now()
	y.count = 0
}

// evySource is imported from JS. The float64 return value encodes the
// ptr (high 32 bits) and length (low 32 bts) of the source string.
//
//export evySource
func evySource() float64

// jsActions is imported from JS. The float64 return value encodes the
// ptr (high 32 bits) and length (low 32 bts) of the actions string.
// The actions string is a comma separate list of actions, e.g.:
// fmt,ui,eval
//
//export jsActions
func jsActions() float64

// jsPrepareUI is imported from JS and sets up UI to suit (e.g. hide/show canvas)
//
//export jsPrepareUI
func jsPrepareUI(string)

// jsRead is imported from JS. The float64 return value encodes the
// ptr (high 32 bits) and length (low 32 bts) of the read string or
// return 0 if no string was read.
//
//export jsRead
func jsRead() float64

// jsPrint is imported from JS
//
//export jsPrint
func jsPrint(string)

// jsCls is imported from JS. It clears all printed output.
//
//export jsCls
func jsCls()

// jsError is imported from JS. jsError is used for setting parse
// errors of format:
//
//	line NUM column NUM: ERROR_DETAILS
//
//export jsError
func jsError(string)

// afterStop is imported from JS
//
//export afterStop
func afterStop()

// move is imported from JS
//
//export move
func move(x, y float64)

// line is imported from JS
//
//export line
func line(x, y float64)

// rect is imported from JS
//
//export rect
func rect(dx, dy float64)

// circle is imported from JS
//
//export circle
func circle(r float64)

// width is imported from JS, setting the lineWidth
//
//export width
func width(w float64)

// color is imported from JS
//
//export color
func color(s string)

// clear is imported from JS
//
//export clear
func clear(s string)

// gridn is imported from JS
//
//export gridn
func gridn(unit float64, s string)

// poly is imported from JS
//
//export poly
func poly(s string)

// ellipse is imported from JS
//
//export ellipse
func ellipse(x, y, rX, rY, rotation, startAngle, endAngle float64)

// stroke is imported from JS
//
//export stroke
func stroke(s string)

// fill is imported from JS
//
//export fill
func fill(s string)

// dash is imported from JS
//
//export dash
func dash(s string)

// linecap is imported from JS
//
//export linecap
func linecap(s string)

// text is imported from JS
//
//export text
func text(s string)

// font is imported from JS
//
//export font
func font(s string)

// setEvySource is imported from JS
//
//export setEvySource
func setEvySource(s string)

//export registerEventHandler
func registerEventHandler(eventName string)
