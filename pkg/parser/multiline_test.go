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
		parser := newParser(input, testBuiltins())
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
		parser := newParser(input, testBuiltins())
		parser.formatting = newFormatting()
		parser.advanceTo(0)
		scope := newScope(nil, &Program{})

		mapLit := parser.parseMapLiteral(scope).(*MapLiteral)
		assertNoParseError(t, parser, input)
		got := parser.formatting.multiline[ptr(mapLit)]
		assert.Equal(t, want, got)
	}
}

func TestNLIndices(t *testing.T) {
	tests := map[string]map[int]bool{
		`func f1
			print 1
		end // needs nl after this line, stmt IDX 0
		func f2
			print 2
		end`: {0: true},
		`
		func f1
			print 1
		end // needs nl after this line, stmt IDX 0
		func f2
			print 2
		end`: {1: true},
		`
		func f1
			print 1
		end // needs nl after this line, stmt IDX 0
		// f2 comment
		func f2
			print 2
		end`: {1: true},
		`
		func f1
			print 1
		end // needs nl after this line, stmt IDX 0
		// f2 comment
		// f2 comment continued
		on down
			print 2
		end`: {1: true},
		`
		// f1 comment
		func f1
			print 1
		end // needs nl after this line, stmt IDX 0
		print 1
		// f2 comment
		// f2 comment continued
		on down
			print 2
		end`: {2: true, 3: true},
	}
	for input, want := range tests {
		parser := newParser(input, testBuiltins())
		prog := parser.Parse()
		assertNoParseError(t, parser, input)
		got := nlAfter(prog.Statements, prog.formatting.comments)
		// cannot equality test maps in tinygo
		// panic: unimplemented: (reflect.Value).MapKeys()
		// so test maps entries individually.
		assert.Equal(t, len(want), len(got))
		for key := range got {
			assert.Equal(t, want[key], got[key])
		}
	}
}
