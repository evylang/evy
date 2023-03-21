//go:build tinygo

package main

import "foxygo.at/evy/pkg/evaluator"

// This file contains Go/WASM functions exported to and called by JS.

// alloc pre-allocates memory used in string parameter passing.
//
//export alloc
func alloc(size uint32) *byte {
	buf := make([]byte, size)
	return &buf[0]
}

// stop is called when JS/UI interactions triggers stop
//
//export stop
func stop() {
	// unsynchronized access to eval.Stopped - ok in WASM as single threaded.
	if eval != nil {
		eval.Stopped = true
	}
}

// onUp is called when pointerup JS/UI event is triggered
//
//export onUp
func onUp(x, y float64) {
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, newXYEvent("up", x, y))
}

// onUp is called when pointerdown JS/UI event is triggered
//
//export onDown
func onDown(x, y float64) {
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, newXYEvent("down", x, y))
}

// onMove is called when pointermvoe JS/UI event is triggered
//
//export onMove
func onMove(x, y float64) {
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, newXYEvent("move", x, y))
}

// onMove is called when keydown JS/UI event is triggered
//
//export onKey
func onKey(ptr *uint32, length int) {
	str := getString(ptr, length)
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, newKeyEvent(str))
}

// onAnimate is wired to JS requestAnimationFrame
//
//export onAnimate
func onAnimate(elapsed float64) {
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, newAnimateEvent(elapsed))
}

// onInput is called when JS/UI generic input event is triggered.
// this input event can be wired flexibly in JS
//
//export onInput
func onInput(idPtr *uint32, idLength int, valPtr *uint32, valLength int) {
	id := getString(idPtr, idLength)
	val := getString(valPtr, valLength)
	// unsynchronized access to events - ok in WASM as single threaded.
	events = append(events, newInputEvent(id, val))
}

func newXYEvent(name string, x, y float64) evaluator.Event {
	return evaluator.Event{
		Name:   name,
		Params: []any{x, y},
	}
}

func newKeyEvent(key string) evaluator.Event {
	return evaluator.Event{
		Name:   "key",
		Params: []any{key},
	}
}

func newInputEvent(id, val string) evaluator.Event {
	return evaluator.Event{
		Name:   "input",
		Params: []any{id, val},
	}
}

func newAnimateEvent(elapsed float64) evaluator.Event {
	return evaluator.Event{
		Name:   "animate",
		Params: []any{elapsed},
	}
}
