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

func parse(t *testing.T, input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if err := resolver.Resolve(program, symbol.NewTable()); err != nil {
		t.Fatalf("Resolver failed: %v", err)
	}

	if err := checker.Check(program); err != nil {
		t.Fatalf("Type checker failed: %v", err)
	}
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
			input: "print true",
			expected: `
			.data
			.text
			li t0, 1
			mv a0, t0
			li a7, 1
			ecall`,
		},
		{
			input: "print false",
			expected: `
			.data
			.text
			li t0, 0
			mv a0, t0
			li a7, 1
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
			input: "var x float",
			expected: `
			.data
			x: .double 0
			.text`,
		},
		{
			input: "var x bool",
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
			input: "var x func() int",
			expected: `
			.data
			x: .dword 0
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
			la s1, x`,
		},
		{
			input: "var x int = 2",
			expected: `
			.data
			x: .dword 0
			.text
			la s1, x
			li t0, 2
			sd t0, 0(s1)`,
		},
		{
			input: `
			type human struct{name string}
			var x human
			`,
			expected: `
			.data
			x: .dword 0
			.text
			li a0, 8
			li a7, 9
			ecall
			la t0, x
			sd a0, 0(t0)
			`,
		},
		{
			input: `
			type human struct{
				name string
				age int
			}
			var x human
			`,
			expected: `
			.data
			x: .dword 0
			.text
			li a0, 16
			li a7, 9
			ecall
			la t0, x
			sd a0, 0(t0)
			`,
		},
		{
			input: `
			func incrementer(x int) int {
				return x + 1
			}

			var x func(int) int
			x = incrementer
			`,
			expected: `
			.data
			x: .dword 0
			.text
			la s1, x
			la s10, incrementer
			sd s10, 0(s1)
			incrementer:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			li t1, 1
			add t0, t0, t1
			mv a0, t0
			addi sp, sp, 0
			j incrementer.epilogue
			addi sp, sp, 0
			incrementer.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret`,
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
			la s1, x
			li t0, 2
			sd t0, 0(s1)`,
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
			la s1, x
			fld ft0, .L1, t0
			fsd ft0, 0(s1)`,
		},
		{
			input: `
			var x bool
			x = true
			`,
			expected: `
			.data
			x: .dword 0
			.text
			la s1, x
			li t0, 1
			sd t0, 0(s1)
			`,
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
			la s1, x
			la t0, .L1
			sd t0, 0(s1)`,
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

func TestTypeStatement(t *testing.T) {
	tests := []compilerTest{
		{
			input: `
			type human struct{name string}
			`,
			expected: `
			.data
			.text`,
		},
	}

	runCompilerTests(t, tests)
}

func TestSelectorExpression(t *testing.T) {
	tests := []compilerTest{
		{
			input: `
			type human struct{name string}
			var x human
			print x.name
			`,
			expected: `
			.data
			x: .dword 0
			.text
			li a0, 8
			li a7, 9
			ecall
			la t0, x
			sd a0, 0(t0)
			la s1, x
			ld s1, 0(s1)
			ld s1, 0(s1)
			mv a0, s1
			li a7, 4
			ecall`,
		},
		{
			input: `
			type human struct{name string}

			var x human

			x.name = "Mads"
			`,
			expected: `
			.data
			x: .dword 0
			.L1: .string "Mads"
			.text
			li a0, 8
			li a7, 9
			ecall
			la t0, x
			sd a0, 0(t0)
			la s1, x
			ld s1, 0(s1)
			la t0, .L1
			sd t0, 0(s1)`,
		},
	}

	runCompilerTests(t, tests)
}

func TestFuncStatement(t *testing.T) {
	tests := []compilerTest{
		{
			input: `func greeter() {
				print true
			}`,
			expected: `
			.data
			.text
			greeter:
			addi sp, sp, -16
			sd ra, 16(sp)
			addi sp, sp, -0
			li t0, 1
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 0
			greeter.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			`,
		},
		{
			input: `func greeter(x int) {
				print x
			}`,
			expected: `
			.data
			.text
			greeter:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 0
			greeter.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			`,
		},
		{
			input: `func greeter(x string) {
				print x
			}`,
			expected: `
			.data
			.text
			greeter:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			mv a0, t0
			li a7, 4
			ecall
			addi sp, sp, 0
			greeter.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			`,
		},
		{
			input: `func greeter(x string) string {
				return x
			}`,
			expected: `
			.data
			.text
			greeter:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			mv a0, t0
			addi sp, sp, 0
			j greeter.epilogue
			addi sp, sp, 0
			greeter.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			`,
		},
		{
			input: `
			type human struct{name string; age int}
			func greeter() human {
				var x human
				return x
			}`,
			expected: `
			.data
			.text
			greeter:
			addi sp, sp, -16
			sd ra, 16(sp)
			addi sp, sp, -16
			li a0, 16
			li a7, 9
			ecall
			sd a0, 8(sp)
			ld t0, 8(sp)
			mv a0, t0
			addi sp, sp, 16
			j greeter.epilogue
			addi sp, sp, 16
			greeter.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			`,
		},
		{
			input: `
			type human struct{name string; age int}
			func greeter(x human) int {
				return x.age
			}`,
			expected: `
			.data
			.text
			greeter:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			ld t0, 8(t0)
			mv a0, t0
			addi sp, sp, 0
			j greeter.epilogue
			addi sp, sp, 0
			greeter.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			`,
		},
		{
			input: `
			func incrementer(x int) int {
				return x + 1
			}

			func decrementer(x int) int {
				return x - 1
			}

			func getFunc(x int) func(int) int {
				if x == 1 {
					return incrementer
				}
				return decrementer
			}
			`,
			expected: `
			.data
			.text
			incrementer:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			li t1, 1
			add t0, t0, t1
			mv a0, t0
			addi sp, sp, 0
			j incrementer.epilogue
			addi sp, sp, 0
			incrementer.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			decrementer:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			li t1, 1
			sub t0, t0, t1
			mv a0, t0
			addi sp, sp, 0
			j decrementer.epilogue
			addi sp, sp, 0
			decrementer.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			getFunc:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			li t1, 1
			beq t0, t1, .L1
			li t0, 0
			b .L2
			.L1:
			li t0, 1
			.L2:
			beqz t0, .L3
			addi sp, sp, -0
			la t0, incrementer
			mv a0, t0
			addi sp, sp, 0
			j getFunc.epilogue
			addi sp, sp, 0
			b .L4
			.L3:
			.L4:
			la t0, decrementer
			mv a0, t0
			addi sp, sp, 0
			j getFunc.epilogue
			addi sp, sp, 0
			getFunc.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret`,
		},
	}

	runCompilerTests(t, tests)
}

func TestCallExpression(t *testing.T) {
	tests := []compilerTest{
		{
			input: `
			func greeter(x string) {
				print x
			}

			greeter("Hello Compiler World")`,
			expected: `
			.data
			.L1: .string "Hello Compiler World"
			.text
			la t0, .L1
			mv a0, t0
			call greeter
			greeter:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			mv a0, t0
			li a7, 4
			ecall
			addi sp, sp, 0
			greeter.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			`,
		},
		{
			input: `
			type human struct{name string; age int}
			func greeter(x human) {
				print x.age
			}

			var h human
			h.name = "Didac"
			h.age = 0

			greeter(h)`,
			expected: `
			.data
			h: .dword 0
			.L1: .string "Didac"
			.text
			li a0, 16
			li a7, 9
			ecall
			la t0, h
			sd a0, 0(t0)
			la s1, h
			ld s1, 0(s1)
			la t0, .L1
			sd t0, 0(s1)
			la s1, h
			ld s1, 0(s1)
			li t0, 0
			sd t0, 8(s1)
			la s1, h
			ld s1, 0(s1)
			mv a0, s1
			call greeter
			greeter:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			ld t0, 8(t0)
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 0
			greeter.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			`,
		},
		{
			input: `
			func incrementer(x int) int {
				return x + 1
			}

			var x func(int) int
			x = incrementer
			`,
			expected: `
			.data
			x: .dword 0
			.text
			la s1, x
			la s10, incrementer
			sd s10, 0(s1)
			incrementer:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			li t1, 1
			add t0, t0, t1
			mv a0, t0
			addi sp, sp, 0
			j incrementer.epilogue
			addi sp, sp, 0
			incrementer.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret`,
		},
		{
			input: `
			var x func() func(int)
			var y func(int)
			func test() func(int)
			func test2(int)

			func test() func(int) {
				return test2
			}

			func test2(x int) {
				print x
			}

			x = test
			y = x()
			y(10)
			`,
			expected: `
			.data
			x: .dword 0
			y: .dword 0
			.text
			la s1, x
			la s10, test
			sd s10, 0(s1)
			la s1, y
			la t0, x
			ld t0, 0(t0)
			jalr t0
			sd a0, 0(s1)
			li t0, 10
			mv a0, t0
			la t0, y
			ld t0, 0(t0)
			jalr t0
			test:
			addi sp, sp, -16
			sd ra, 16(sp)
			addi sp, sp, -0
			la t0, test2
			mv a0, t0
			addi sp, sp, 0
			j test.epilogue
			addi sp, sp, 0
			test.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			test2:
			addi sp, sp, -16
			sd a0, 8(sp)
			sd ra, 16(sp)
			addi sp, sp, -0
			ld t0, 8(sp)
			mv a0, t0
			li a7, 1
			ecall
			addi sp, sp, 0
			test2.epilogue:
			ld ra, 16(sp)
			addi sp, sp, 16
			ret
			`,
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTest) {
	t.Helper()

	// replace all those pesky tabs and newlines from the raw strings.
	replacer := strings.NewReplacer("\t", "", "\n", "")

	for _, tt := range tests {
		program := parse(t, tt.input)

		comp := newTest()

		if err := comp.Compile(program); err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		asm := comp.Asm()
		if replacer.Replace(asm) != replacer.Replace(tt.expected) {
			t.Errorf(
				"wrong assembly emitted.\nexpected=\n%s\ngot=\n%s",
				strings.Replace(tt.expected, "\t", "", -1),
				strings.Replace(asm, "\t", "", -1),
			)
		}
	}

}
