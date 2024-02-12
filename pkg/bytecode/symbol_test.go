package bytecode

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestDefine(t *testing.T) {
	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "b", Scope: GlobalScope, Index: 1},
	}
	global := NewSymbolTable()
	for _, sym := range expected {
		actual := global.Define(sym.Name)
		assert.Equal(t, sym, actual)
	}
}

func TestResolveGlobal(t *testing.T) {
	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
	}
	global := NewSymbolTable()
	for _, sym := range expected {
		global.Define(sym.Name)
		result, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		assert.Equal(t, sym, result, "wrong value for %s")
	}
}
