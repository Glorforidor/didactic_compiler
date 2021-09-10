package compiler

import (
	"testing"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/lexer"
	"github.com/Glorforidor/didactic_compiler/parser"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func TestPrintStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input: "print 42",
			expected: `li a0, 42
li a7, 1
ecall`,
		},
		{
			input: `print 42
print 43`,
			expected: `li a0, 42
li a7, 1
ecall
li a0, 43
li a7, 1
ecall`,
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)

		comp := New()

		if err := comp.Compile(program); err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		asm := comp.Asm()
		if asm != tt.expected {
			t.Fatalf("wrong assembly emitted.\nexpected=%q\ngot=%q", tt.expected, asm)
		}
	}
}
