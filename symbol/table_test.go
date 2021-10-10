package symbol

import (
	"testing"

	types "github.com/Glorforidor/didactic_compiler/types"
)

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"x": Symbol{Name: "x", Scope: GlobalScope, Type: types.Type{Kind: types.Int}, Which: 0},
		"y": Symbol{Name: "y", Scope: GlobalScope, Type: types.Type{Kind: types.Int}, Which: 1},
		"a": Symbol{Name: "a", Scope: LocalScope, Type: types.Type{Kind: types.Int}, Which: 0},
		"b": Symbol{Name: "b", Scope: LocalScope, Type: types.Type{Kind: types.Int}, Which: 1},
	}

	global := NewTable()

	x, _ := global.Define("x", types.Type{Kind: types.Int})
	if x != expected["x"] {
		t.Errorf("expected x=%+v, got=%+v", expected["x"], x)
	}

	y, _ := global.Define("y", types.Type{Kind: types.Int})
	if y != expected["y"] {
		t.Errorf("expected y=%+v, got=%+v", expected["y"], y)
	}

	local := NewEnclosedTable(global)

	a, _ := local.Define("a", types.Type{Kind: types.Int})
	if a != expected["a"] {
		t.Errorf("expected a=%+v, got=%+v", expected["a"], a)
	}

	b, _ := local.Define("b", types.Type{Kind: types.Int})
	if b != expected["b"] {
		t.Errorf("expected a=%+v, got=%+v", expected["b"], b)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewTable()
	global.Define("x", types.Type{Kind: types.Int})
	global.Define("y", types.Type{Kind: types.Int})

	expected := []Symbol{
		Symbol{Name: "x", Scope: GlobalScope, Type: types.Type{Kind: types.Int}, Which: 0},
		Symbol{Name: "y", Scope: GlobalScope, Type: types.Type{Kind: types.Int}, Which: 1},
	}

	for _, sym := range expected {
		t.Run(sym.Name, func(t *testing.T) {
			result, ok := global.Resolve(sym.Name)
			if !ok {
				t.Fatalf("name %s not resolvable", sym.Name)
			}

			if result != sym {
				t.Fatalf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
			}
		})
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewTable()
	global.Define("a", types.Type{Kind: types.Int})
	global.Define("b", types.Type{Kind: types.Int})

	local := NewEnclosedTable(global)
	local.Define("c", types.Type{Kind: types.Int})
	local.Define("d", types.Type{Kind: types.Int})

	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Type: types.Type{Kind: types.Int}, Which: 0},
		Symbol{Name: "b", Scope: GlobalScope, Type: types.Type{Kind: types.Int}, Which: 1},
		Symbol{Name: "c", Scope: LocalScope, Type: types.Type{Kind: types.Int}, Which: 0},
		Symbol{Name: "d", Scope: LocalScope, Type: types.Type{Kind: types.Int}, Which: 1},
	}

	for _, sym := range expected {
		result, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}

		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
		}
	}
}

func TestCode(t *testing.T) {
	tests := []struct {
		input    Symbol
		expected string
	}{
		{
			input:    Symbol{Name: "x", Scope: GlobalScope, Type: types.Type{Kind: types.Int}},
			expected: "x",
		},
	}

	for _, tt := range tests {
		got := tt.input.Code()
		if tt.expected != got {
			t.Fatalf("symbol had wrong code. expected=%q, got=%q", tt.expected, got)
		}
	}
}
