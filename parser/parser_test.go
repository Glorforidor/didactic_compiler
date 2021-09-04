package parser

import (
	"testing"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/lexer"
)

func TestPrintStatement(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`print 42`},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if stmt.TokenLiteral() != "print" {
			t.Fatalf("stmt.TokenLiteral not 'print'. got=%q", stmt.TokenLiteral())
		}

		printStmt, ok := stmt.(*ast.PrintStatement)
		if !ok {
			t.Fatalf("stmt not *ast.PrintStatement. got=%T", stmt)
		}

		if printStmt.TokenLiteral() != "print" {
			t.Fatalf("printStmt.TokenLiteral not 'print', got=%q", printStmt.TokenLiteral())
		}
	}
}
