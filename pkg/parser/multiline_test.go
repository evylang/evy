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
		got := arrayLit.multilines
		assert.Equal(t, want, got)
	}
}
