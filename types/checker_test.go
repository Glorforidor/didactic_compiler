package types

import (
	"testing"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/lexer"
	"github.com/Glorforidor/didactic_compiler/parser"
)

func TestChecker(t *testing.T) {
	tests := []struct {
		input         string
		expectedType  ast.Type
		expectedError bool
	}{
		{
			input:         "print 2",
			expectedType:  ast.Type{Kind: ast.Int},
			expectedError: false,
		},
		{
			input:         "print 2 + 2",
			expectedType:  ast.Type{Kind: ast.Int},
			expectedError: false,
		},
		{
			input:         "print 2.0 + 2",
			expectedType:  ast.Type{},
			expectedError: true,
		},
		{
			input:         "print 2.0 + 2",
			expectedType:  ast.Type{Kind: ast.Int},
			expectedError: true,
		},
		{
			input:         `print "Hello World" + 2`,
			expectedType:  ast.Type{Kind: ast.Int},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if err := Checker(program); err != nil {
			if !tt.expectedError {
				t.Fatalf("checker had errors which was not expected. got=%s", err)
			}
			t.Log(tt.expectedType)

			return
		}

		for _, s := range program.Statements {
			switch s := s.(type) {
			case *ast.PrintStatement:
				if s.Value.Type() != tt.expectedType {
					t.Fatalf("checker added wrong type: expected=%s, got=%s", tt.expectedType, s.Value.Type())
				}
			}
		}
	}
}
