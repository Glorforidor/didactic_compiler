package symbol

import (
	"testing"

	"github.com/Glorforidor/didactic_compiler/token"
)

func TestDefine(t *testing.T) {
	expected := map[string]*Symbol{
		"x": {Name: "x", Scope: GlobalScope, Type: token.IntType, which: 0},
		"y": {Name: "y", Scope: GlobalScope, Type: token.IntType, which: 1},
		"a": {Name: "a", Scope: LocalScope, Type: token.IntType, which: 0},
		"b": {Name: "b", Scope: LocalScope, Type: token.IntType, which: 1},
	}

	global := NewTable()

	x, _ := global.Define("x", token.IntType)
	if *x != *expected["x"] {
		t.Errorf("expected x=%+v, got=%+v", expected["x"], x)
	}

	y, _ := global.Define("y", token.IntType)
	if *y != *expected["y"] {
		t.Errorf("expected y=%+v, got=%+v", expected["y"], y)
	}

	local := NewEnclosedTable(global)

	a, _ := local.Define("a", token.IntType)
	if *a != *expected["a"] {
		t.Errorf("expected a=%+v, got=%+v", expected["a"], a)
	}

	b, _ := local.Define("b", token.IntType)
	if *b != *expected["b"] {
		t.Errorf("expected a=%+v, got=%+v", expected["b"], b)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewTable()
	global.Define("x", token.IntType)
	global.Define("y", token.IntType)

	expected := []*Symbol{
		{Name: "x", Scope: GlobalScope, Type: token.IntType, which: 0},
		{Name: "y", Scope: GlobalScope, Type: token.IntType, which: 1},
	}

	for _, sym := range expected {
		t.Run(sym.Name, func(t *testing.T) {
			result, ok := global.Resolve(sym.Name)
			if !ok {
				t.Fatalf("name %s not resolvable", sym.Name)
			}

			if *result != *sym {
				t.Fatalf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
			}
		})
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewTable()
	global.Define("a", token.IntType)
	global.Define("b", token.IntType)

	local := NewEnclosedTable(global)
	local.Define("c", token.IntType)
	local.Define("d", token.IntType)

	expected := []*Symbol{
		{Name: "a", Scope: GlobalScope, Type: token.IntType, which: 0},
		{Name: "b", Scope: GlobalScope, Type: token.IntType, which: 1},
		{Name: "c", Scope: LocalScope, Type: token.IntType, which: 0},
		{Name: "d", Scope: LocalScope, Type: token.IntType, which: 1},
	}

	for _, sym := range expected {
		result, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}

		if *result != *sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
		}
	}
}

func TestResolveMultipleLocal(t *testing.T) {
	global := NewTable()
	global.Define("a", token.IntType)
	global.Define("b", token.IntType)

	local := NewEnclosedTable(global)
	local.Define("c", token.IntType)
	local.Define("d", token.IntType)

	secondLocal := NewEnclosedTable(local)
	secondLocal.Define("e", token.IntType)
	secondLocal.Define("f", token.IntType)
	secondLocal.Define("g", token.IntType)

	expected := []*Symbol{
		{Name: "a", Scope: GlobalScope, Type: token.IntType, which: 0},
		{Name: "b", Scope: GlobalScope, Type: token.IntType, which: 1},
		{Name: "c", Scope: LocalScope, Type: token.IntType, which: 0},
		{Name: "d", Scope: LocalScope, Type: token.IntType, which: 1},
		{Name: "e", Scope: LocalScope, Type: token.IntType, which: 0},
		{Name: "f", Scope: LocalScope, Type: token.IntType, which: 1},
		{Name: "g", Scope: LocalScope, Type: token.IntType, which: 2},
	}

	for _, sym := range expected {
		result, ok := secondLocal.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}

		if *result != *sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
		}
	}
}

func TestCode(t *testing.T) {
	tests := []struct {
		input    Symbol
		expected interface{}
	}{
		{
			input:    Symbol{Name: "x", Scope: GlobalScope, Type: token.IntType},
			expected: "x",
		},
		{
			input:    Symbol{Name: "x", Scope: LocalScope, Type: token.IntType, stackPoint: 8},
			expected: 8,
		},
	}

	for _, tt := range tests {
		got := tt.input.Code()
		if tt.expected != got {
			t.Fatalf("symbol had wrong code. expected='%v', got='%v'", tt.expected, got)
		}
	}
}
