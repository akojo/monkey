package compiler

import "testing"

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()

	a := global.Define("a")
	if a != expected["a"] {
		t.Errorf("a: want %+v, got %+v", expected["a"], a)
	}

	b := global.Define("b")
	if b != expected["b"] {
		t.Errorf("b: want %+v, got %+v", expected["b"], b)
	}

	outer := NewEnclosedSymbolTable(global)

	c := outer.Define("c")
	if c != expected["c"] {
		t.Errorf("c: want %+v, got %+v", expected["c"], c)
	}

	d := outer.Define("d")
	if d != expected["d"] {
		t.Errorf("d: want %+v, got %+v", expected["d"], d)
	}

	inner := NewEnclosedSymbolTable(outer)

	e := inner.Define("e")
	if e != expected["e"] {
		t.Errorf("e: want %+v, got %+v", expected["e"], e)
	}

	f := inner.Define("f")
	if f != expected["f"] {
		t.Errorf("f: want %+v, got %+v", expected["f"], f)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, sym := range expected {
		result, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("%s: not found", sym.Name)
		}
		if result != sym {
			t.Errorf("%s: want %+v, got %+v", sym.Name, sym, result)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "c", Scope: LocalScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		result, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("%s: not found", sym.Name)
		}
		if result != sym {
			t.Errorf("%s: want %+v, got %+v", sym.Name, sym, result)
		}
	}
}

func TestResolveNestedLocals(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	outer := NewEnclosedSymbolTable(global)
	outer.Define("c")
	outer.Define("d")

	inner := NewEnclosedSymbolTable(outer)
	inner.Define("e")
	inner.Define("f")

	tests := []struct {
		table    *SymbolTable
		expected []Symbol
	}{
		{outer,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			}},
		{inner,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			}},
	}

	for _, test := range tests {
		for _, sym := range test.expected {
			result, ok := test.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("%s: not found", sym.Name)
			}
			if result != sym {
				t.Errorf("%s: want %+v, got %+v", sym.Name, sym, result)
			}
		}
	}
}

func TestResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)

	expected := []Symbol{
		{Name: "a", Scope: BuiltinScope, Index: 0},
		{Name: "c", Scope: BuiltinScope, Index: 1},
		{Name: "e", Scope: BuiltinScope, Index: 2},
		{Name: "f", Scope: BuiltinScope, Index: 3},
	}

	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}

	for _, table := range []*SymbolTable{global, firstLocal, secondLocal} {
		for _, sym := range expected {
			result, ok := table.Resolve(sym.Name)
			if !ok {
				t.Errorf("%s: not found", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("%s: want %+v, got %+v", sym.Name, sym, result)
			}
		}
	}
}

func TestResolveFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table               *SymbolTable
		expectedSymbols     []Symbol
		expectedFreeSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
			[]Symbol{},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: FreeScope, Index: 0},
				{Name: "d", Scope: FreeScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
			[]Symbol{
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, test := range tests {
		for _, symbol := range test.expectedSymbols {
			result, ok := test.table.Resolve(symbol.Name)
			if !ok {
				t.Errorf("%s: not found", symbol.Name)
				continue
			}
			if result != symbol {
				t.Errorf("symbol %s: want %+v, got %+v", symbol.Name, symbol, result)
			}
		}

		if len(test.table.FreeSymbols) != len(test.expectedFreeSymbols) {
			t.Errorf("free symbols: want %d, got %d", len(test.expectedFreeSymbols), len(test.table.FreeSymbols))
			continue
		}

		for i, symbol := range test.expectedFreeSymbols {
			result := test.table.FreeSymbols[i]
			if result != symbol {
				t.Errorf("free symbol %s: want %+v, got %+v", symbol.Name, symbol, result)
			}
		}
	}
}

func TestUnresolved(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "c", Scope: FreeScope, Index: 0},
		{Name: "e", Scope: LocalScope, Index: 0},
		{Name: "f", Scope: LocalScope, Index: 1},
	}

	for _, symbol := range expected {
		result, ok := secondLocal.Resolve(symbol.Name)
		if !ok {
			t.Errorf("%s: not found", symbol.Name)
			continue
		}
		if result != symbol {
			t.Errorf("symbol %s: want %+v, got %+v", symbol.Name, symbol, result)
		}
	}

	for _, name := range []string{"b", "d"} {
		_, ok := secondLocal.Resolve(name)
		if ok {
			t.Errorf("%s: should not resolve", name)
		}
	}

}
