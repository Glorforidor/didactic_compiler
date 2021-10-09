package checker

import (
	"testing"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/lexer"
	"github.com/Glorforidor/didactic_compiler/parser"
	"github.com/Glorforidor/didactic_compiler/resolver"
	"github.com/Glorforidor/didactic_compiler/symbol"
	"github.com/Glorforidor/didactic_compiler/types"
)

type checkerTest struct {
	input         string
	expectedType  types.Type
	expectedToErr bool
}

func TestPrintStatement(t *testing.T) {
	tests := []checkerTest{
		{
			input:         "print 2",
			expectedType:  types.Type{Kind: types.Int},
			expectedToErr: false,
		},
		{
			input:         "print 2 + 2",
			expectedType:  types.Type{Kind: types.Int},
			expectedToErr: false,
		},
		{
			input:         "print 2.0 + 2",
			expectedType:  types.Type{},
			expectedToErr: true,
		},
		{
			input:         `print "Hello World" + 2`,
			expectedType:  types.Type{},
			expectedToErr: true,
		},
	}

	runCheckerTests(t, tests)
}

func TestVarStatement(t *testing.T) {
	tests := []checkerTest{
		{
			input:         "var x int",
			expectedType:  types.Type{Kind: types.Int},
			expectedToErr: false,
		},
		{
			input:         "var x int = 1",
			expectedType:  types.Type{Kind: types.Int},
			expectedToErr: false,
		},
		{
			input:         "var x float = 1.0",
			expectedType:  types.Type{Kind: types.Float},
			expectedToErr: false,
		},
		{
			input:         `var x string = "Hello World"`,
			expectedType:  types.Type{Kind: types.String},
			expectedToErr: false,
		},
	}

	runCheckerTests(t, tests)
}

func TestAssignStatement(t *testing.T) {
	tests := []checkerTest{
		{
			input: `var x int
x = 2`,
			expectedType:  types.Type{Kind: types.Int},
			expectedToErr: false,
		},
		{
			input: `var x int
x = 2.5`,
			expectedType:  types.Type{Kind: types.Int},
			expectedToErr: true,
		},
	}

	runCheckerTests(t, tests)
}

func TestIdentifier(t *testing.T) {
	tests := []checkerTest{
		{
			input: `var x int
x`,
			expectedType:  types.Type{Kind: types.Int},
			expectedToErr: false,
		},
		{
			input: `var x float
x`,
			expectedType:  types.Type{Kind: types.Float},
			expectedToErr: false,
		},
		{
			input: `var x string
x`,
			expectedType:  types.Type{Kind: types.String},
			expectedToErr: false,
		},
	}

	runCheckerTests(t, tests)
}

func TestBlockStatement(t *testing.T) {
	// TODO: Write better tests for Block statements so testing that a variable
	// with the same name in different blocks can have different types.
	tests := []checkerTest{
		{
			input: `{
var x int
}`,
			expectedType:  types.Type{Kind: types.Int},
			expectedToErr: false,
		},
		{
			input: `{
	var x float
}`,
			expectedType:  types.Type{Kind: types.Float},
			expectedToErr: false,
		},
		{
			input: `{
	var x string
}`,
			expectedType:  types.Type{Kind: types.String},
			expectedToErr: false,
		},
	}

	runCheckerTests(t, tests)
}

func TestArithmetic(t *testing.T) {
	tests := []checkerTest{
		{
			input:         `2 + 2`,
			expectedType:  types.Type{Kind: types.Int},
			expectedToErr: false,
		},
		{
			input:         `2.0 + 2`,
			expectedType:  types.Type{},
			expectedToErr: true,
		},
		{
			input:         `"Hello" + "World"`,
			expectedType:  types.Type{},
			expectedToErr: true,
		},
	}

	runCheckerTests(t, tests)
}

func runCheckerTests(t *testing.T, tests []checkerTest) {
	t.Helper()
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		resolver.Resolve(program, symbol.NewTable())

		t.Logf("Program: %v", program.String())

		if err := Check(program); err != nil {
			if !tt.expectedToErr {
				t.Fatalf("checker had errors which was not expected. got=%s", err)
			}

			continue
		}

		t.Logf("Typed Program: %v", program.String())

		var testing func(node ast.Node)
		testing = func(node ast.Node) {
			switch node := node.(type) {
			case *ast.Program:
				for _, s := range node.Statements {
					testing(s)
				}
			case *ast.BlockStatement:
				for _, s := range node.Statements {
					testing(s)
				}
			case *ast.PrintStatement:
				if node.Value.Type() != tt.expectedType {
					t.Fatalf("added wrong type: expected=%s, got=%s", tt.expectedType, node.Value.Type())
				}
			case *ast.VarStatement:
				if node.Name.T != tt.expectedType {
					t.Fatalf(
						"variable was defined with the wrong type. expected=%s, got=%s",
						tt.expectedType, node.Name.T.Kind)
				}

				if node.Value != nil {
					if node.Name.Type() != node.Value.Type() {
						t.Fatalf("allowed to add type: %s to an identifier with type: %s", node.Value.Type(), node.Name.Type())
					}
				}
			case *ast.AssignStatement:
				if node.Name.T != tt.expectedType {
					t.Fatalf(
						"variable in assignment have unexpected type. expected=%s, got=%s",
						tt.expectedType, node.Name.T.Kind)
				}
			case *ast.ExpressionStatement:
				testing(node.Expression)
			case *ast.Identifier:
				if node.T != tt.expectedType {
					t.Fatalf("identified wrong type.")
				}
			}
		}

		testing(program)
	}
}
