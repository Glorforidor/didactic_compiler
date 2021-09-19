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

type compilerTest struct {
	input, expected string
}

func TestPrintStatement(t *testing.T) {
	tests := []compilerTest{
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
		{
			input: `print 42.1`,
			expected: `.data
.L1:
.double 42.1
.text
fld fa0, .L1, a0
li a7, 3
ecall`,
		},
		{
			input: `print 2 + 2`,
			expected: `li a0, 2
li a1, 2
add a0, a0, a1
li a7, 1
ecall`,
		},
	}

	runCompilerTests(t, tests)
}

func TestArithmetic(t *testing.T) {
	tests := []compilerTest{
		{
			input: `2 + 2`,
			expected: `li a0, 2
li a1, 2
add a0, a0, a1`,
		},
		{
			input: `2 + 2
3 + 3`,
			expected: `li a0, 2
li a1, 2
add a0, a0, a1
li a0, 3
li a1, 3
add a0, a0, a1`,
		},
		{
			input: "2 - 2",
			expected: `li a0, 2
li a1, 2
sub a0, a0, a1`,
		},
		{
			input: "2 * 2",
			expected: `li a0, 2
li a1, 2
mul a0, a0, a1`,
		},
		{
			input: "2 / 2",
			expected: `li a0, 2
li a1, 2
div a0, a0, a1`,
		},
		{
			input: "2.1 + 2.1",
			expected: `.data
.L1:
.double 2.1
.text
fld fa0, .L1, a0
.data
.L2:
.double 2.1
.text
fld fa1, .L2, a0
fadd.d fa0, fa0, fa1`,
		},
		{
			input: "2.1 - 2.1",
			expected: `.data
.L1:
.double 2.1
.text
fld fa0, .L1, a0
.data
.L2:
.double 2.1
.text
fld fa1, .L2, a0
fsub.d fa0, fa0, fa1`,
		},
		{
			input: "2.1 * 2.1",
			expected: `.data
.L1:
.double 2.1
.text
fld fa0, .L1, a0
.data
.L2:
.double 2.1
.text
fld fa1, .L2, a0
fmul.d fa0, fa0, fa1`,
		},
		{
			input: "2.1 / 2.1",
			expected: `.data
.L1:
.double 2.1
.text
fld fa0, .L1, a0
.data
.L2:
.double 2.1
.text
fld fa1, .L2, a0
fdiv.d fa0, fa0, fa1`,
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTest) {
	t.Helper()

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
