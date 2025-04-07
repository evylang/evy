//go:build !tinygo && full

// This file contains the embed directives of the full Evy web content used in
// Evy releases and with the "full" build tag. It is accessed with
//
//	evy serve
//
// The default web content, initialized in main.go, is a placeholder
// index.html with installation and build instructions for the full build.
//
// The full build depends on the evy.wasm file, which needs to be pre-built
// with TinyGo. We don't track the evy.wasm binary in the evy repository and
// its build is comparatively slow, so it's not included in the default
// build. The evy.wasm-dependency-free default build also allows for a
// generic go install to work with
//
//	go install evylang.dev/evy@latest .
//
// To include the full Evy web contents pre-built to the out/embed directory,
// use the 'full' build tag:
//
//	go build -tags full
//	go install -tags full
//
// A full clean build regenerating evy.wasm and out/embed can be executed with:
//
//	make build-full
//	make install-full
package main

import "embed"

//go:embed out/embed
var fullContent embed.FS

func init() {
	content = fullContent
	contentDir = "out/embed"
	fullBuild = true
}
