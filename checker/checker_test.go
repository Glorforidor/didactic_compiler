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
			input:         "var x int = 1",
			expectedType:  types.Type{Kind: types.Int},
			expectedToErr: false,
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
	}

	runCheckerTests(t, tests)
}

func runCheckerTests(t *testing.T, tests []checkerTest) {
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		resolver.Resolve(program, symbol.NewTable())

		if err := Check(program); err != nil {
			if !tt.expectedToErr {
				t.Fatalf("checker had errors which was not expected. got=%s", err)
			}

			continue
		}

		t.Logf("Program: %s", program.String())

		for _, s := range program.Statements {
			switch s := s.(type) {
			case *ast.PrintStatement:
				if s.Value.Type() != tt.expectedType {
					t.Fatalf("added wrong type: expected=%s, got=%s", tt.expectedType, s.Value.Type())
				}
			case *ast.VarStatement:
				if s.Value != nil {
					if s.Value.Type() != tt.expectedType {
						t.Fatalf("added wrong type: expected=%s, got=%s", tt.expectedType, s.Value.Type())
					}
					if s.Name.Type() != s.Value.Type() {
						t.Fatalf("allowed to add type: %s to an identifier with type: %s", s.Value.Type(), s.Name.Type())
					}
				}
			case *ast.ExpressionStatement:
				switch e := s.Expression.(type) {
				case *ast.Identifier:
					if e.T != tt.expectedType {
						t.Fatalf("identified wrong type.")
					}
				}
			}
		}
	}
}
