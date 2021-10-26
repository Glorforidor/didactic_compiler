package compiler

import (
	"strings"
	"testing"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/checker"
	"github.com/Glorforidor/didactic_compiler/lexer"
	"github.com/Glorforidor/didactic_compiler/parser"
	"github.com/Glorforidor/didactic_compiler/resolver"
	"github.com/Glorforidor/didactic_compiler/symbol"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	resolver.Resolve(program, symbol.NewTable())
	checker.Check(program)
	return program
}

type compilerTest struct {
	input, expected string
}

func TestArithmetic(t *testing.T) {
	tests := []compilerTest{
		{
			input: "2 + 2",
			expected: `
			.data
			.text
			li t0, 2
			li t1, 2
			add t0, t0, t1`,
		},
		{
			input: `2 + 2
					3 + 3`,
			expected: `
			.data
			.text
			li t0, 2
			li t1, 2
			add t0, t0, t1
			li t0, 3
			li t1, 3
			add t0, t0, t1`,
		},
		{
			input: "2 - 2",
			expected: `
			.data
			.text
			li t0, 2
			li t1, 2
			sub t0, t0, t1`,
		},
		{
			input: "2 * 2",
			expected: `
			.data
			.text
			li t0, 2
			li t1, 2
			mul t0, t0, t1`,
		},
		{
			input: "2 / 2",
			expected: `
			.data
			.text
			li t0, 2
			li t1, 2
			div t0, t0, t1`,
		},
		{
			input: "2.1 + 2.1",
			expected: `
			.data
			.L1: .double 2.1
			.L2: .double 2.1
			.text
			fld ft0, .L1, t0
			fld ft1, .L2, t0
			fadd.d ft0, ft0, ft1`,
		},
		{
			input: "2.1 - 2.1",
			expected: `
			.data
			.L1: .double 2.1
			.L2: .double 2.1
			.text
			fld ft0, .L1, t0
			fld ft1, .L2, t0
			fsub.d ft0, ft0, ft1`,
		},
		{
			input: "2.1 * 2.1",
			expected: `
			.data
			.L1: .double 2.1
			.L2: .double 2.1
			.text
			fld ft0, .L1, t0
			fld ft1, .L2, t0
			fmul.d ft0, ft0, ft1`,
		},
		{
			input: "2.1 / 2.1",
			expected: `
			.data
			.L1: .double 2.1
			.L2: .double 2.1
			.text
			fld ft0, .L1, t0
			fld ft1, .L2, t0
			fdiv.d ft0, ft0, ft1`,
		},
		{
			input: "2 + 2 - 2",
			expected: `
				.data
				.text
				li t0, 2
				li t1, 2
				add t0, t0, t1
				li t1, 2
				sub t0, t0, t1
			`,
		},
		{
			input: "(5 + 5) / 5",
			expected: `
			.data
			.text
			li t0, 5
			li t1, 5
			add t0, t0, t1
			li t1, 5
			div t0, t0, t1`,
		},
	}

	runCompilerTests(t, tests)
}

func TestPrintStatement(t *testing.T) {
	tests := []compilerTest{
		{
			input: "print 42",
			expected: `
			.data
			.text
			li t0, 42
			mv a0, t0
			li a7, 1
			ecall`,
		},
		{
			input: `
			print 42
			print 43`,
			expected: `
			.data
			.text
			li t0, 42
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
			expected: `
			.data
			.L1: .string "Hello World"
			.text
			la t0, .L1
			mv a0, t0
			li a7, 4
			ecall`,
		},
		{
			input: `
			print "Hello World"
			print "Hello Peeps"`,
			expected: `
			.data
			.L1: .string "Hello World"
			.L2: .string "Hello Peeps"
			.text
			la t0, .L1
			mv a0, t0
			li a7, 4
			ecall
			la t0, .L2
			mv a0, t0
			li a7, 4
			ecall`,
		},
		{
			input: `print 42.1`,
			expected: `
			.data
			.L1: .double 42.1
			.text
			fld ft0, .L1, t0
			fmv.d fa0, ft0
			li a7, 3
			ecall`,
		},
		{
			input: "print 2 + 2",
			expected: `
			.data
			.text
			li t0, 2
			li t1, 2
			add t0, t0, t1
			mv a0, t0
			li a7, 1
			ecall`,
		},
	}
	runCompilerTests(t, tests)
}

func TestVarStatement(t *testing.T) {
	tests := []compilerTest{
		{
			input: "var x int",
			expected: `
			.data
			x: .dword 0
			.text`,
		},
		{
			input: "var x string",
			expected: `
			.data
			x: .dword 0
			.text`,
		},
		{
			input: "var x float",
			expected: `
			.data
			x: .double 0
			.text`,
		},
		{
			input: `
			var x int
			x`,
			expected: `
			.data
			x: .dword 0
			.text
			la t0, x`,
		},
		{
			input: "var x int = 2",
			expected: `
			.data
			x: .dword 0
			.text
			la t0, x
			li t1, 2
			sd t1, 0(t0)`,
		},
	}

	runCompilerTests(t, tests)
}

func TestAssignStatement(t *testing.T) {
	tests := []compilerTest{
		{
			input: `
			var x int
			x = 2
			`,
			expected: `
			.data
			x: .dword 0
			.text
			la t0, x
			li t1, 2
			sd t1, 0(t0)`,
		},
		{
			input: `
			var x float
			x = 2.0
			`,
			expected: `
			.data
			x: .double 0
			.L1: .double 2
			.text
			la t0, x
			fld ft0, .L1, t1
			fsd ft0, 0(t0)`,
		},
		{
			input: `
			var x string
			x = "Hello Compiler World"
			`,
			expected: `
			.data
			x: .dword 0
			.L1: .string "Hello Compiler World"
			.text
			la t0, x
			la t1, .L1
			sd t1, 0(t0)`,
		},
	}

	runCompilerTests(t, tests)
}

func TestBlockStatement(t *testing.T) {
	tests := []compilerTest{
		{
			input: `
			{
				var x int
			}`,
			expected: `
			.data
			.text
			addi sp, sp, -16
			addi sp, sp, 16`,
		},
		{
			input: `
			{
				var x float
			}`,
			expected: `
			.data
			.text
			addi sp, sp, -16
			addi sp, sp, 16`,
		},
		{
			input: `
			{
				var x int = 2
			}`,
			expected: `
			.data
			.text
			addi sp, sp, -16
			ld t0, 8(sp)
			li t1, 2
			sd t1, 8(sp)
			addi sp, sp, 16`,
		},
		{
			input: `
			{
				var x string = "Hello Block statement"
			}`,
			expected: `
			.data
			.L1: .string "Hello Block statement"
			.text
			addi sp, sp, -16
			ld t0, 8(sp)
			la t1, .L1
			sd t1, 8(sp)
			addi sp, sp, 16`,
		},
		{
			input: `
			{
				var x int
				var y float = 32.0
				x = 2
			}`,
			expected: `
			.data
			.L1: .double 32
			.text
			addi sp, sp, -16
			fld ft0, 16(sp)
			fld ft1, .L1, t0
			fsd ft1, 16(sp)
			ld t0, 8(sp)
			li t1, 2
			sd t1, 8(sp)
			addi sp, sp, 16`,
		},
	}

	runCompilerTests(t, tests)
}

func TestConditional(t *testing.T) {
	tests := []compilerTest{
		{
			input: "2 < 3",
			expected: `
			.data
			.text
			li t0, 2
			li t1, 3
			blt t0, t1, .L1
			li t0, 0
			b .L2
			.L1:
			li t0, 1
			.L2:`,
		},
	}

	runCompilerTests(t, tests)
}

func TestIfStatement(t *testing.T) {
	tests := []compilerTest{
		{
			input: `
			if 2 < 3 {
				print 2
			}`,
			expected: `
			.data
			.text
			li t0, 2
			li t1, 3
			blt t0, t1, .L1
			li t0, 0
			b .L2
			.L1:
			li t0, 1
			.L2:
			beqz t0, .L3
			addi sp, sp, -0
			li t0, 2
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 0
			b .L4
			.L3:
			.L4:
			`,
		},
		{
			input: `
			if 2 < 3 {
				print 2
			} else {
				print 3
			}`,
			expected: `
			.data
			.text
			li t0, 2
			li t1, 3
			blt t0, t1, .L1
			li t0, 0
			b .L2
			.L1:
			li t0, 1
			.L2:
			beqz t0, .L3
			addi sp, sp, -0
			li t0, 2
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 0
			b .L4
			.L3:
			addi sp, sp, -0
			li t0, 3
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 0
			.L4:
			`,
		},
		{
			input: `
			if 2 < 3 {
				var x int = 2
				print x
			} else {
				var x int = 3
				print x
			}`,
			expected: `
			.data
			.text
			li t0, 2
			li t1, 3
			blt t0, t1, .L1
			li t0, 0
			b .L2
			.L1:
			li t0, 1
			.L2:
			beqz t0, .L3
			addi sp, sp, -16
			ld t0, 8(sp)
			li t1, 2
			sd t1, 8(sp)
			ld t0, 8(sp)
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 16
			b .L4
			.L3:
			addi sp, sp, -16
			ld t0, 8(sp)
			li t1, 3
			sd t1, 8(sp)
			ld t0, 8(sp)
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 16
			.L4:
			`,
		},
		{
			input: `
			if 2 + 2 < 3 + 3 {
				print 2
			}`,
			expected: `
			.data
			.text
			li t0, 2
			li t1, 2
			add t0, t0, t1
			li t1, 3
			li t2, 3
			add t1, t1, t2
			blt t0, t1, .L1
			li t0, 0
			b .L2
			.L1:
			li t0, 1
			.L2:
			beqz t0, .L3
			addi sp, sp, -0
			li t0, 2
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 0
			b .L4
			.L3:
			.L4:
			`,
		},
	}

	runCompilerTests(t, tests)
}

func TestForStatement(t *testing.T) {
	tests := []compilerTest{
		{
			input: `
			for var i int = 0; i < 10; i = i + 1 {
				print i
			}`,
			expected: `
			.data
			.text
			addi sp, sp, -16
			ld t0, 8(sp)
			li t1, 0
			sd t1, 8(sp)
			.L1:
			ld t0, 8(sp)
			li t1, 10
			blt t0, t1, .L3
			li t0, 0
			b .L4
			.L3:
			li t0, 1
			.L4:
			beqz t0, .L2
			addi sp, sp, -0
			ld t0, 8(sp)
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 0
			ld t0, 8(sp)
			ld t1, 8(sp)
			li t2, 1
			add t1, t1, t2
			sd t1, 8(sp)
			b .L1
			.L2:
			addi sp, sp, 16`,
		},
	}

	runCompilerTests(t, tests)
}

	tests := []compilerTest{
		{
			expected: `
			.data
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTest) {
	t.Helper()

	// replace all those pesky tabs and newlines from the raw strings.
	replacer := strings.NewReplacer("\t", "", "\n", "")

	for _, tt := range tests {
		program := parse(tt.input)

		comp := New()

		if err := comp.Compile(program); err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		asm := comp.Asm()
		if replacer.Replace(asm) != replacer.Replace(tt.expected) {
			t.Fatalf("wrong assembly emitted.\nexpected=\n%q\ngot=\n%q", replacer.Replace(tt.expected), replacer.Replace(asm))
		}
	}

}
