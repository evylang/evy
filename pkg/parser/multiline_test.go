package parser

import (
	"testing"

	"foxygo.at/evy/pkg/assert"
)

func TestArrayLiteralMultiline(t *testing.T) {
	tests := map[string][]multilineItem{
		"[1]":   {multilineEl},
		"[1 2]": {multilineEl, multilineEl},
		`[1
		2]`: {multilineEl, multilineNL, multilineEl},
		`[
		1
		2
		]`: {multilineNL, multilineEl, multilineNL, multilineEl, multilineNL},
		`[1 // comment 1
		// comment 2
		2]`: {multilineEl, multilineComment("// comment 1\n"), multilineComment("  // comment 2   "), multilineEl},
	}
	for input, want := range tests {
		parser := New(input, testBuiltins())
		parser.formatting = newFormatting()
		parser.advanceTo(0)
		scope := newScope(nil, &Program{})

		arrayLit := parser.parseArrayLiteral(scope).(*ArrayLiteral)
		assertNoParseError(t, parser, input)
		got := parser.formatting.multiline[ptr(arrayLit)]
		assert.Equal(t, want, got)
	}
}

func TestMapLiteralMultiline(t *testing.T) {
	tests := map[string][]multilineItem{
		"{a:1}":     {"a"},
		"{a:1 b:2}": {"a", "b"},
		`{a:1
		b:2}`: {"a", multilineNL, "b"},
		`{
		a:1

		b:2
		}`: {multilineNL, "a", multilineNL, multilineNL, "b", multilineNL},
		`{ a:1 // comment 1
		// comment 2
		b:2}`: {"a", multilineComment(" // comment 1 "), multilineComment("// comment 2"), "b"},
	}
	for input, want := range tests {
		parser := New(input, testBuiltins())
		parser.formatting = newFormatting()
		parser.advanceTo(0)
		scope := newScope(nil, &Program{})

		mapLit := parser.parseMapLiteral(scope).(*MapLiteral)
		assertNoParseError(t, parser, input)
		got := parser.formatting.multiline[ptr(mapLit)]
		assert.Equal(t, want, got)
	}
}
