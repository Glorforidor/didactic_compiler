package resolver

import (
	"testing"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/lexer"
	"github.com/Glorforidor/didactic_compiler/parser"
	"github.com/Glorforidor/didactic_compiler/symbol"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func TestResolv(t *testing.T) {
	tests := []struct {
		input         string
		expectedToErr bool
	}{
		{
			input:         "var x int",
			expectedToErr: false,
		},
		{
			input: `var x int
var x float`,
			expectedToErr: true,
		},
		{
			input:         "x",
			expectedToErr: true,
		},
		{
			input: `var x int
x`,
			expectedToErr: false,
		},
		{
			input: `x
var x int`,
			expectedToErr: true,
		},
		{
			input:         `x = 2`,
			expectedToErr: true,
		},
		{
			input: `var x int
x = 2`,
			expectedToErr: false,
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		symbolTable := symbol.NewTable()

		err := Resolve(program, symbolTable)

		if err == nil && tt.expectedToErr {
			t.Fatalf("expected to fail, but succeeded")
		}

		if err != nil && !tt.expectedToErr {
			t.Fatalf("expected not to fail, got: %s", err)
		}
	}
}
