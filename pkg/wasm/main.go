//go:build tinygo

package main

import (
	"strings"
	"unsafe"

	"foxygo.at/evy/pkg/evaluator"
	"foxygo.at/evy/pkg/lexer"
	"foxygo.at/evy/pkg/parser"
)

var version string

func main() {
}

// jsPrint is imported from JS
//export jsPrint
func jsPrint(string)

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

// evaluate evaluates an evy program, after tokenizing and parsing. It
// is exported to wasm and JS. Strings cannot be passed to wasm
// directly so we need to use linear memory arithmetic as workaround.
// See:
// * https://www.wasm.builders/k33g_org/an-essay-on-the-bi-directional-exchange-of-strings-between-the-wasm-module-with-tinygo-and-nodejs-with-wasi-support-3i9h
// * https://www.alcarney.me/blog/2020/passing-strings-between-tinygo-wasm/
//
//export evaluate
func jsEvaluate(ptr *uint32, length int) {
	s := getString(ptr, length)
	builtins := evaluator.DefaultBuiltins(jsRuntime)
	evaluator.RunWithBuiltins(s, builtins)
}

//export tokenize
func jsTokenize(ptr *uint32, length int) {
	s := getString(ptr, length)
	jsPrint(lexer.Run(s))
}

//export parse
func jsParse(ptr *uint32, length int) {
	s := getString(ptr, length)
	builtins := evaluator.DefaultBuiltins(jsRuntime).Decls()
	jsPrint(parser.Run(s, builtins))
}

// alloc pre-allocates memory used in string parameter passing.
//
//export alloc
func alloc(size uint32) *byte {
	buf := make([]byte, size)
	return &buf[0]
}

// getString turns pointers in linear memory into string, see comments
// for evaluate.
func getString(ptr *uint32, length int) string {
	var builder strings.Builder
	uptr := uintptr(unsafe.Pointer(ptr))
	for i := 0; i < length; i++ {
		s := *(*int32)(unsafe.Pointer(uptr + uintptr(i)))
		builder.WriteByte(byte(s))
	}
	return builder.String()
}
