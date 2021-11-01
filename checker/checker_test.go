package checker

import (
	"fmt"
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
			expectedType:  types.Typ[types.Int],
			expectedToErr: false,
		},
		{
			input:         "print 2 + 2",
			expectedType:  types.Typ[types.Int],
			expectedToErr: false,
		},
		{
			input:         "print 2.0 + 2",
			expectedType:  nil,
			expectedToErr: true,
		},
		{
			input:         `print "Hello World" + 2`,
			expectedType:  nil,
			expectedToErr: true,
		},
	}

	runCheckerTests(t, tests)
}

func TestVarStatement(t *testing.T) {
	tests := []checkerTest{
		{
			input:         "var x int",
			expectedType:  types.Typ[types.Int],
			expectedToErr: false,
		},
		{
			input:         "var x int = 1",
			expectedType:  types.Typ[types.Int],
			expectedToErr: false,
		},
		{
			input:         "var x float = 1.0",
			expectedType:  types.Typ[types.Float],
			expectedToErr: false,
		},
		{
			input:         `var x string = "Hello World"`,
			expectedType:  types.Typ[types.String],
			expectedToErr: false,
		},
		{
			input:         `var x bool = false`,
			expectedType:  types.Typ[types.Bool],
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
			expectedType:  types.Typ[types.Int],
			expectedToErr: false,
		},
		{
			input: `var x int
x = 2.5`,
			expectedType:  types.Typ[types.Int],
			expectedToErr: true,
		},
		{
			input: `var x bool
x = "True"`,
			expectedType:  types.Typ[types.Int],
			expectedToErr: true,
		},
	}

	runCheckerTests(t, tests)
}

func TestTypeStatement(t *testing.T) {
	tests := []checkerTest{
		{
			input: `type human struct{name string}`,
			expectedType: &types.Struct{
				Fields: []*types.Field{
					{
						Name: "name",
						Type: types.Typ[types.String],
					},
				},
			},
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
			expectedType:  types.Typ[types.Int],
			expectedToErr: false,
		},
		{
			input: `var x float
x`,
			expectedType:  types.Typ[types.Float],
			expectedToErr: false,
		},
		{
			input: `var x string
x`,
			expectedType:  types.Typ[types.String],
			expectedToErr: false,
		},
		{
			input: `var x bool
x`,
			expectedType:  types.Typ[types.Bool],
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
			expectedType:  types.Typ[types.Int],
			expectedToErr: false,
		},
		{
			input: `{
	var x float
}`,
			expectedType:  types.Typ[types.Float],
			expectedToErr: false,
		},
		{
			input: `{
	var x string
}`,
			expectedType:  types.Typ[types.String],
			expectedToErr: false,
		},
	}

	runCheckerTests(t, tests)
}

func TestIfStatement(t *testing.T) {
	tests := []checkerTest{
		{
			input: `
			if 2 < 2 {
				print 2
			}`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: false,
		},
		{
			input: `
			if 2 < 3 {
				print 20
			} else {
				print 30
			}`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: false,
		},
		{
			input: `
			if true {
				print 20
			} else {
				print 30
			}`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: false,
		},
		{
			input: `
			if 1 {
				print 20
			}`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: true,
		},
		{
			input: `
			if 1 + 1 {
				print 20
			}`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: true,
		},
	}

	runCheckerTests(t, tests)
}

func TestForStatement(t *testing.T) {
	tests := []checkerTest{
		{
			input: `
			for var i int = 0; i < 10; i = i + 1 {
				print i
			}`,
			expectedToErr: false,
		},
		{
			input: `
			var i float
			for i = 0; i < 10; i = i + 1 {
				print i
			}`,
			expectedToErr: true,
		},
		{
			input: `
			for var i int = 0; i + 10; i = i + 1 {
				print i
			}`,
			expectedToErr: true,
		},
	}

	runCheckerTests(t, tests)
}

func TestComparsion(t *testing.T) {
	tests := []checkerTest{
		{
			input:         `2 < 2`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: false,
		},
		{
			input:         `2 == 2`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: false,
		},
		{
			input:         `2 != 2`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: false,
		},
		{
			input:         `2 < false`,
			expectedType:  nil,
			expectedToErr: true,
		},
		{
			input:         `2 < 3 == true`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: false,
		},
		{
			input:         `2 < 3 == true`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: false,
		},
		{
			input:         `true < true`,
			expectedType:  types.Typ[types.Bool],
			expectedToErr: true,
		},
	}

	runCheckerTests(t, tests)
}

func TestFuncStatement(t *testing.T) {
	tests := []struct {
		input             string
		expectedFuncType  types.Type
		expectedParamType types.Type
		expectedToErr     bool
	}{
		{
			input: `func greeter(x string) {
				print x
			}`,
			expectedFuncType: &types.Signature{
				Parameter: types.Typ[types.String],
				Result:    nil,
			},
			expectedParamType: types.Typ[types.String],
			expectedToErr:     false,
		},
		{
			input: `func greeter(x string) string {
				print x
			}`,
			expectedFuncType: &types.Signature{
				Parameter: types.Typ[types.String],
				Result:    types.Typ[types.String],
			},
			expectedParamType: types.Typ[types.String],
			expectedToErr:     false,
		},
		{
			input: `func greeter(x int) {
		print x + "Hello"
		}`,
			expectedFuncType:  &types.Signature{},
			expectedParamType: types.Typ[types.Int],
			expectedToErr:     true,
		},
		{
			input: `func greeter(x int) string {
		print x + "Hello"
		}`,
			expectedFuncType:  &types.Signature{},
			expectedParamType: types.Typ[types.Int],
			expectedToErr:     true,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		resolver.Resolve(program, symbol.NewTable())

		t.Logf("Program: %v", program.String())

		err := Check(program)
		if err != nil && tt.expectedToErr {
			continue
		}

		if err != nil && !tt.expectedToErr {
			t.Fatalf("checker had errors which was not expected. got=%s", err)
		}

		if err == nil && tt.expectedToErr {
			t.Fatalf("checker was assumed to fail, but it did not.")
		}

		funcStmt, _ := program.Statements[0].(*ast.FuncStatement)

		if funcStmt.Name.T.String() != tt.expectedFuncType.String() {
			t.Fatalf("funcStmt.Name.T is not %q. got=%q", tt.expectedFuncType.String(), funcStmt.Name.T.String())
		}

		if funcStmt.Parameter.T != tt.expectedParamType {
			t.Fatalf("funcStmt.Parameter.T is not %s. got=%s", tt.expectedParamType, funcStmt.Parameter.T)
		}
	}
}

func TestLiteral(t *testing.T) {
	tests := []checkerTest{
		{
			input:        "2",
			expectedType: types.Typ[types.Int],
		},
		{
			input:        "2.0",
			expectedType: types.Typ[types.Float],
		},
		{
			input:        `"Hello Compiler World`,
			expectedType: types.Typ[types.String],
		},
		{
			input:        "true",
			expectedType: types.Typ[types.Bool],
		},
		{
			input:        "false",
			expectedType: types.Typ[types.Bool],
		},
	}

	runCheckerTests(t, tests)
}

func TestArithmetic(t *testing.T) {
	tests := []checkerTest{
		{
			input:         `2 + 2`,
			expectedType:  types.Typ[types.Int],
			expectedToErr: false,
		},
		{
			input:         `2.0 + 2`,
			expectedType:  nil,
			expectedToErr: true,
		},
		{
			input:         `"Hello" + "World"`,
			expectedType:  nil,
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

		err := Check(program)
		if err != nil && tt.expectedToErr {
			continue
		}

		if err != nil && !tt.expectedToErr {
			t.Fatalf("checker had errors which was not expected. got=%s", err)
		}

		if err == nil && tt.expectedToErr {
			t.Fatalf("checker was assumed to fail, but it did not.")
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
						"variable: %s was defined with the wrong type. expected=%s, got=%s",
						node.Name.TokenLiteral(), tt.expectedType, node.Name.T)
				}

				if node.Value != nil {
					if node.Name.Type() != node.Value.Type() {
						t.Fatalf("allowed to add type: %s to an identifier with type: %s", node.Value.Type(), node.Name.Type())
					}
				}
			case *ast.TypeStatement:
				n, _ := node.Name.T.(*types.Struct)
				vv, _ := tt.expectedType.(*types.Struct)
				for i, f := range n.Fields {
					if f.Type != vv.Fields[i].Type {
						t.Fatalf("identifier: %q was defined with the wrong type. expected=%s, got=%s", node.Name.TokenLiteral(), tt.expectedType, node.Name.T)
					}
				}
				// if node.Name.T != tt.expectedType {
				// 	t.Fatalf("identifier: %q was defined with the wrong type. expected=%s, got=%s", node.Name.TokenLiteral(), tt.expectedType, node.Name.T)
				// }
			case *ast.AssignStatement:
				if node.Name.T != tt.expectedType {
					t.Fatalf(
						"variable in assignment have unexpected type. expected=%s, got=%s",
						tt.expectedType, node.Name.T)
				}
			case *ast.ExpressionStatement:
				testing(node.Expression)
			case *ast.FuncStatement:
				testing(node.Name)
			case *ast.IfStatement:
				testing(node.Condition)
			case *ast.ForStatement:
				// TODO: Should test that the initialise, condition, and next
				// all have the correct type.
			case *ast.Identifier:
				if node.T != tt.expectedType {
					t.Fatalf("identifier has the wrong type. expected=%s, got=%s", tt.expectedType, node.T)
				}
			case *ast.InfixExpression:
				if node.T != tt.expectedType {
					t.Fatalf("infx expression have unexpected type. expected=%s, got=%s",
						tt.expectedType, node.T)
				}
			case *ast.IntegerLiteral:
				if node.T != tt.expectedType {
					t.Fatalf("int literal have unexpected type. expected=%s, got=%s",
						tt.expectedType, node.T)
				}
			case *ast.FloatLiteral:
				if node.T != tt.expectedType {
					t.Fatalf("float literal have unexpected type. expected=%s, got=%s",
						tt.expectedType, node.T)
				}
			case *ast.StringLiteral:
				if node.T != tt.expectedType {
					t.Fatalf("string literal have unexpected type. expected=%s, got=%s",
						tt.expectedType, node.T)
				}
			case *ast.BoolLiteral:
				if node.T != tt.expectedType {
					t.Fatalf("bool literal have unexpected type. expected=%s, got=%s",
						tt.expectedType, node.T)
				}
			default:
				panic(fmt.Sprintf("unhandled type: %T", node))
			}
		}

		testing(program)
	}
}
