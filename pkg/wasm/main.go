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
	eval.Yielder = newSleepingYielder()
	eval.Run(getEvySource())
	handleEvents()
	onStopped()
}

func getEvySource() string {
	addr := evySource()
	ptr, length := decodePtrLen(uint64(addr))
	return getString(ptr, length)
}

func read() string {
	for {
		if eval.Stopped {
			return ""
		}
		addr := jsRead()
		if addr != 0 {
			ptr, length := decodePtrLen(uint64(addr))
			return getString(ptr, length)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// newSleepingYielder yields the CPU so that JavaScript/browser events
// get a chance to be processed. Currently(Feb 2023) it seems that you
// can only yield to JS by sleeping for at least 1ms but having that
// delay is not ideal. Other methods of yielding can be explored by
// implementing a different yield function.
func newSleepingYielder() *sleepingYielder {
	return &sleepingYielder{start: time.Now()}
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

func newXYEvent(name string, x, y float64) evaluator.Event {
	return evaluator.Event{
		Name:   name,
		Params: []any{x, y},
	}
}

func newStringEvent(name, str string) evaluator.Event {
	return evaluator.Event{
		Name:   name,
		Params: []any{str},
	}
}

// --- JS function exported to Go/WASM ---------------------------------

// evySource is imported from JS. The float64 return value encodes the
// ptr (high 32 bits) and length (low 32 bts) of the source string.
//export evySource
func evySource() float64

// jsRead is imported from JS. The float64 return value encodes the
// ptr (high 32 bits) and length (low 32 bts) of the read string or
// return 0 if no string was read.
//export jsRead
func jsRead() float64

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
	Read:  read,
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

func decodePtrLen(ptrLen uint64) (ptr *uint32, length int) {
	ptr = (*uint32)(unsafe.Pointer(uintptr(ptrLen >> 32)))
	length = int(uint32(ptrLen))
	return
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
	events = append(events, newXYEvent("up", x, y))
}

//export onDown
func onDown(x, y float64) {
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, newXYEvent("down", x, y))
}

//export onMove
func onMove(x, y float64) {
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, newXYEvent("move", x, y))
}

//export onKey
func onKey(ptr *uint32, length int) {
	str := getString(ptr, length)
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, newStringEvent("key", str))
}
