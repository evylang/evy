//go:build tinygo

package main

import (
	"strings"
	"unsafe"
)

// getString turns pointer and length in linear memory into string.
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

// getStringFromAddr retrieves a string for a single float64 value that
// encodes ptr and length of the string in memory. The single ptrLen
// value is used when returning strings from JS to go/Was as only
// single number values can be returned.
func getStringFromAddr(ptrLen float64) string {
	ptr, length := decodePtrLen(uint64(ptrLen))
	return getString(ptr, length)
}

// decodePtrLen decodes a single uint64 into a ptr, higher 32 bits, and
// a length, lower 32 bits. Ptr and length are then used to turn memory
// bytes into strings.
func decodePtrLen(ptrLen uint64) (ptr *uint32, length int) {
	ptr = (*uint32)(unsafe.Pointer(uintptr(ptrLen >> 32)))
	length = int(uint32(ptrLen))
	return
}
