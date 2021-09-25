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
			expected: `li t0, 42
mv a0, t0
li a7, 1
ecall`,
		},
		{
			input: `print 42
print 43`,
			expected: `li t0, 42
mv a0, t0
li a7, 1
ecall
li t0, 43
mv a0, t0
li a7, 1
ecall`,
		},
		{
			input: `print "Hello World"`,
			expected: `.data
.L1:
.string "Hello World"
.text
la t0, .L1
mv a0, t0
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
la t0, .L1
mv a0, t0
li a7, 4
ecall
.data
.L2:
.string "Hello Peeps"
.text
la t0, .L2
mv a0, t0
li a7, 4
ecall`,
		},
		{
			input: `print 42.1`,
			expected: `.data
.L1:
.double 42.1
.text
fld ft0, .L1, t0
fmv.d fa0, ft0
li a7, 3
ecall`,
		},
		{
			input: `print 2 + 2`,
			expected: `li t0, 2
li t1, 2
add t0, t0, t1
mv a0, t0
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
			expected: `li t0, 2
li t1, 2
add t0, t0, t1`,
		},
		{
			input: `2 + 2
3 + 3`,
			expected: `li t0, 2
li t1, 2
add t0, t0, t1
li t0, 3
li t1, 3
add t0, t0, t1`,
		},
		{
			input: "2 - 2",
			expected: `li t0, 2
li t1, 2
sub t0, t0, t1`,
		},
		{
			input: "2 * 2",
			expected: `li t0, 2
li t1, 2
mul t0, t0, t1`,
		},
		{
			input: "2 / 2",
			expected: `li t0, 2
li t1, 2
div t0, t0, t1`,
		},
		{
			input: "2.1 + 2.1",
			expected: `.data
.L1:
.double 2.1
.text
fld ft0, .L1, t0
.data
.L2:
.double 2.1
.text
fld ft1, .L2, t0
fadd.d ft0, ft0, ft1`,
		},
		{
			input: "2.1 - 2.1",
			expected: `.data
.L1:
.double 2.1
.text
fld ft0, .L1, t0
.data
.L2:
.double 2.1
.text
fld ft1, .L2, t0
fsub.d ft0, ft0, ft1`,
		},
		{
			input: "2.1 * 2.1",
			expected: `.data
.L1:
.double 2.1
.text
fld ft0, .L1, t0
.data
.L2:
.double 2.1
.text
fld ft1, .L2, t0
fmul.d ft0, ft0, ft1`,
		},
		{
			input: "2.1 / 2.1",
			expected: `.data
.L1:
.double 2.1
.text
fld ft0, .L1, t0
.data
.L2:
.double 2.1
.text
fld ft1, .L2, t0
fdiv.d ft0, ft0, ft1`,
		},
		{
			input: "2 + 2 - 2",
			expected: `li t0, 2
li t1, 2
add t0, t0, t1
li t1, 2
sub t0, t0, t1`,
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
