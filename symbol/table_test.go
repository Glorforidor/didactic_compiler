package symbol

import (
	"testing"

	types "github.com/Glorforidor/didactic_compiler/types"
)

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"x": Symbol{Name: "x", Scope: GlobalScope, Type: types.Type{Kind: types.Int}},
		"y": Symbol{Name: "y", Scope: GlobalScope, Type: types.Type{Kind: types.Int}},
	}

	global := NewTable()

	x := global.Define("x", types.Type{Kind: types.Int})
	if x != expected["x"] {
		t.Errorf("expected x=%+v, got=%+v", expected["x"], x)
	}

	y := global.Define("y", types.Type{Kind: types.Int})
	if y != expected["y"] {
		t.Errorf("expected y=%+v, got=%+v", expected["y"], y)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewTable()
	global.Define("x", types.Type{Kind: types.Int})
	global.Define("y", types.Type{Kind: types.Int})

	expected := []Symbol{
		Symbol{Name: "x", Scope: GlobalScope, Type: types.Type{Kind: types.Int}},
		Symbol{Name: "y", Scope: GlobalScope, Type: types.Type{Kind: types.Int}},
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
