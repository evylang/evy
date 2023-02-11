//go:build tinygo

package main

import (
	"strings"
	"time"
	"unsafe"

	"foxygo.at/evy/pkg/evaluator"
)

var version string
var eval *evaluator.Evaluator
var events []evaluator.Event

const minSleepDur = time.Millisecond

func main() {
	builtins := evaluator.DefaultBuiltins(jsRuntime)
	eval = evaluator.NewEvaluator(builtins)
	eval.Yield = newSleepingYielder()
	eval.Run(getSource())
	handleEvents()
	onStopped()
}

func getSource() string {
	ptr := sourcePtr()
	length := sourceLength()
	return getString(ptr, length)
}

type sleepingYielder struct {
	start time.Time
	count int
}

func (y *sleepingYielder) Yield() {
	y.count++
	if y.count > 1000 && time.Since(y.start) > 100*time.Millisecond {
		time.Sleep(minSleepDur)
		y.start = time.Now()
		y.count = 0
	}
}

// newSleepingYielder yields the CPU so that JavaScript/browser events
// get a chance to be processed. Currently(Feb 2023) it seems that you
// can only yield to JS by sleeping for at least 1ms but having that
// delay is not ideal. Other methods of yielding can be explored by
// implementing a different yield function.
func newSleepingYielder() func() {
	count := 0
	start := time.Now()
	return func() {
		if count > 1000 && time.Since(start) > 100*time.Millisecond {
			time.Sleep(time.Millisecond)
			start = time.Now()
			count = 0
		}
	}
}

func handleEvents() {
	if eval == nil || len(eval.EventHandlerNames()) == 0 {
		return
	}
	for _, name := range eval.EventHandlerNames() {
		registerEventHandler(name)
	}
	for {
		if eval.Stopped {
			return
		}
		// unsynchronized access to events - ok in WASM as single threaded.
		if len(events) > 0 {
			event := events[0]
			events = events[1:]
			eval.HandleEvent(event)
		} else {
			time.Sleep(minSleepDur)
		}
	}
}

// xyEvent implements evaluator.Event.
type xyEvent struct {
	name string
	x    float64
	y    float64
}

func (e *xyEvent) Name() string {
	return e.name
}

func (e *xyEvent) Params() []any {
	return []any{e.x, e.y}
}

// stringEvent implements evaluator.Event.
type stringEvent struct {
	name string
	str  string
}

func (e *stringEvent) Name() string {
	return e.name
}

func (e *stringEvent) Params() []any {
	return []any{e.str}
}

// --- JS function exported to Go/WASM ---------------------------------

//export sourcePtr
func sourcePtr() *uint32

//export sourceLength
func sourceLength() int

// jsPrint is imported from JS
//export jsPrint
func jsPrint(string)

// onStopped is imported from JS
//export onStopped
func onStopped()

// move is imported from JS
//export move
func move(x, y float64)

// line is imported from JS
//export line
func line(x, y float64)

// rect is imported from JS
//export rect
func rect(dx, dy float64)

// circle is imported from JS
//export circle
func circle(r float64)

// width is imported from JS, setting the lineWidth
//export width
func width(w float64)

// color is imported from JS
//export color
func color(s string)

// We cannot take the address of external/exported functions
// (https://golang.org/cmd/cgo/#hdr-Passing_pointers) so we must wrap them in a
// Go function first to put them in this Runtime struct.
var jsRuntime evaluator.Runtime = evaluator.Runtime{
	Print: func(s string) { jsPrint(s) },
	Graphics: evaluator.GraphicsRuntime{
		Move:   func(x, y float64) { move(x, y) },
		Line:   func(x, y float64) { line(x, y) },
		Rect:   func(dx, dy float64) { rect(dx, dy) },
		Circle: func(r float64) { circle(r) },
		Width:  func(w float64) { width(w) },
		Color:  func(s string) { color(s) },
	},
}

//export registerEventHandler
func registerEventHandler(eventName string)

// --- Go function exported to JS/WASM runtime -------------------------

// alloc pre-allocates memory used in string parameter passing.
//
//export alloc
func alloc(size uint32) *byte {
	buf := make([]byte, size)
	return &buf[0]
}

// getString turns pointer and length in linear memory into string
// Strings cannot be passed to or returned from wasm directly so we
// need to use linear memory arithmetic as workaround.
// See:
// * https://www.wasm.builders/k33g_org/an-essay-on-the-bi-directional-exchange-of-strings-between-the-wasm-module-with-tinygo-and-nodejs-with-wasi-support-3i9h
// * https://www.alcarney.me/blog/2020/passing-strings-between-tinygo-wasm
func getString(ptr *uint32, length int) string {
	var builder strings.Builder
	uptr := uintptr(unsafe.Pointer(ptr))
	for i := 0; i < length; i++ {
		s := *(*int32)(unsafe.Pointer(uptr + uintptr(i)))
		builder.WriteByte(byte(s))
	}
	return builder.String()
}

//export stop
func stop() {
	// unsynchronized access to eval.Stopped - ok in WASM as single threaded.
	if eval != nil {
		eval.Stopped = true
	}
}

//export onUp
func onUp(x, y float64) {
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, &xyEvent{name: "up", x: x, y: y})
}

//export onDown
func onDown(x, y float64) {
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, &xyEvent{name: "down", x: x, y: y})
}

//export onMove
func onMove(x, y float64) {
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, &xyEvent{name: "move", x: x, y: y})
}

//export onKey
func onKey(ptr *uint32, length int) {
	str := getString(ptr, length)
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, &stringEvent{name: "key", str: str})
}
