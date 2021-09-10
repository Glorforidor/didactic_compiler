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
		{
			input: `print "Hello World"`,
			expected: `.data
.L1:
.string "Hello World"
.text
la a0, .L1
li a7, 4
ecall`,
		},
		{
			input: `print "Hello World"
print "Hello Peeps"`,
			expected: `.data
.L1:
.string "Hello World"
.text
la a0, .L1
li a7, 4
ecall
.data
.L2:
.string "Hello Peeps"
.text
la a0, .L2
li a7, 4
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
			t.Fatalf("wrong assembly emitted.\nexpected=\n%q\ngot=\n%q", tt.expected, asm)
		}
	}
}
