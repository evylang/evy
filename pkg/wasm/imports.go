//go:build tinygo

package main

// This file contains JS functions imported into Go/WASM. Functions are
// declared without body, the full definition can be found in the JS
// implementation. The `jsRuntime` struct wraps these functions to
// implement evaluator.Runtime.

import (
	"time"

	"foxygo.at/evy/pkg/evaluator"
)

const minSleepDur = time.Millisecond

// We cannot take the address of external/exported functions
// (https://golang.org/cmd/cgo/#hdr-Passing_pointers) so we must wrap them in a
// Go function first to put them in this Runtime struct.
func newJSRuntime(yielder *sleepingYielder) *evaluator.Runtime {
	return &evaluator.Runtime{
		Print: func(s string) { jsPrint(s) },
		Read:  yielder.Read,
		Sleep: yielder.Sleep,
		Graphics: evaluator.GraphicsRuntime{
			Move:   func(x, y float64) { move(x, y) },
			Line:   func(x, y float64) { line(x, y) },
			Rect:   func(dx, dy float64) { rect(dx, dy) },
			Circle: func(r float64) { circle(r) },
			Width:  func(w float64) { width(w) },
			Color:  func(s string) { color(s) },
		},
	}
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
		time.Sleep(minSleepDur)
		y.Reset()
	}
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

// setEvySource is imported from JS
//
//export setEvySource
func setEvySource(s string)

//export registerEventHandler
func registerEventHandler(eventName string)
