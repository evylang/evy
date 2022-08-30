//go:build tinygo
package main

import (
	"strings"
	"unsafe"
)

var (
	version string
)

func main() {
}

// jsPrint is imported from JS
func jsPrint(string)

// evaluate evaluates an evy program, after tokenizing and parsing. It
// is exported to wasm and JS. Strings cannot be passed to wasm
// directly so we need to use linear memory arithmetic as workaround.
// See:
// * https://www.wasm.builders/k33g_org/an-essay-on-the-bi-directional-exchange-of-strings-between-the-wasm-module-with-tinygo-and-nodejs-with-wasi-support-3i9h
// * https://www.alcarney.me/blog/2020/passing-strings-between-tinygo-wasm/
//export evaluate
func evaluate(ptr *uint32, length int) {
	s := getString(ptr, length)
	result := strings.ToUpper(s)
	jsPrint(result)
}

//export tokenize
func tokenize(ptr *uint32, length int) {
	s := getString(ptr, length)
	result := "tokenize:\n" + truncate(s, 20) + "\n"
	jsPrint(result)
}

//export parse
func parse(ptr *uint32, length int) {
	s := getString(ptr, length)
	result := "parse:\n" + truncate(s, 20) + "\n"
	jsPrint(result)
}

// alloc pre-allocates memory used in string parameter passing.
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

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
