//go:build tinygo

package main

import (
	"fmt"
	"time"

	"foxygo.at/evy/pkg/evaluator"
)

var version string
var eval *evaluator.Evaluator

var source = `func target x:num y:num
  colors := ["yellow" "gold" "orange" "red" "maroon"]
  radius := 5

  for c := range colors
    radius = radius - 1
    print "x" x "y" y "color" c "radius" radius
  end
end

target 40 40
`

func main() {
	builtins := evaluator.DefaultBuiltins(jsRuntime)
	eval = evaluator.NewEvaluator(builtins)
	eval.Yield = newSleepingYielder()
	eval.Run(source)
}

type sleepingYielder struct {
	start time.Time
	count int
}

func (y *sleepingYielder) Yield() {
	y.count++
	if y.count > 1000 && time.Since(y.start) > 100*time.Millisecond {
		time.Sleep(time.Millisecond)
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

// --- JS function exported to Go/WASM ---------------------------------

// We cannot take the address of external/exported functions
// (https://golang.org/cmd/cgo/#hdr-Passing_pointers) so we must wrap them in a
// Go function first to put them in this Runtime struct.
var jsRuntime evaluator.Runtime = evaluator.Runtime{
	Print: func(s string) { fmt.Print(s) },
}

// --- Go function exported to JS/WASM runtime -------------------------
