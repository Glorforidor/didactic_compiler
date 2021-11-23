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
		{
			input: `var x int
{
	var x int
}`,
			expectedToErr: true,
		},
		{
			input: `var x int
{
	var x1 float
}`,
			expectedToErr: false,
		},
		{
			input: `var x int
{
	var x int
	var x float
}`,
			expectedToErr: true,
		},
		{
			input: `
			type human struct{name string}
			`,
			expectedToErr: false,
		},
		{
			input: `
			type human struct{name string}
			`,
			expectedToErr: false,
		},
		{
			input: `
			type human struct{name string}
			`,
			expectedToErr: false,
		},
		{
			input: `
			if 2 < 3 {
				print 2
			}`,
			expectedToErr: false,
		},
		{
			input: `if 2 < 3 {
				var x int
			} else {
				var y int
			}`,
			expectedToErr: false,
		},
		{
			input: `if 2 < 3 {
				x = 2
			} else {
				y = 2.5
			}`,
			expectedToErr: true,
		},
		{
			input: `
			var x int
			var y float
			if 2 < 3 {
				x = 2
			} else {
				y = 2.5
			}`,
			expectedToErr: false,
		},
		{
			input: `
			for var i int = 0; i < 10; i = i + 1 {
				print i
			}
			`,
			expectedToErr: false,
		},
		{
			input: `
			var i int
			for i = 0; i < 10; i = i + 1 {
				print i
			}
			`,
			expectedToErr: false,
		},
		{
			input: `
			for var i int = 0; i < 10; i = i + 1 {
				print x
			}
			`,
			expectedToErr: true,
		},
		{
			input: `
			func test(x int);`,
			expectedToErr: false,
		},
		{
			input: `
			var x int
			func test(x int);
			func test(x int) {
				print x
			}`,
			expectedToErr: false,
		},
		{
			input: `
			func test(x int) {
				print x
			}`,
			expectedToErr: false,
		},
		{
			input: `
			func test(x int) {
				print x
			}
			func test(x string) {
				print x
			}`,
			expectedToErr: true,
		},
		{
			input: `
			var test int
			func test(x int) {
				print x
			}
			func test(x string) {
				print x
			}`,
			expectedToErr: true,
		},
		{
			input: `
			func test(x int) int {
				return x + 2
			}`,
			expectedToErr: false,
		},
		{
			input: `
			func test(x int) int {
				return x + 2
			}

			test(2)`,
			expectedToErr: false,
		},
		{
			input: `
			func test(x int) int {
				return x + 2
			}

			test(x)`,
			expectedToErr: true,
		},
		{
			input: `
			type human struct{name string}
			var x human
			x.name`,
			expectedToErr: false,
		},
	}

	for i, tt := range tests {
		program := parse(tt.input)
		symbolTable := symbol.NewTable()

		err := Resolve(program, symbolTable)

		if err == nil && tt.expectedToErr {
			t.Fatalf("expected to fail, but succeeded")
		}

		if err != nil && !tt.expectedToErr {
			t.Fatalf("test[%d]: expected not to fail, got: %s",i, err)
		}
	}
}
