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

func TestIndex(t *testing.T) {
	t.Run("enclosed index", func(t *testing.T) {
		global := NewSymbolTable()
		global.Define("a")
		global.Define("b")
		nested := global.Push()
		symbol := nested.Define("c")
		assert.Equal(t, 0, symbol.Index)
		nested2 := nested.Push()
		s2 := nested2.Define("d")
		assert.Equal(t, 1, s2.Index)
	})
	t.Run("empty outer", func(t *testing.T) {
		global := NewSymbolTable()
		local := global.Push()
		symbol := local.Define("c")
		assert.Equal(t, 0, symbol.Index)
	})
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
