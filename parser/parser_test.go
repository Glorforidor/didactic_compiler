package parser

import (
	"fmt"
	"testing"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/lexer"
)

func checkParserError(t *testing.T, p *Parser) {
	t.Helper()
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	if len(errors) == 1 {
		t.Errorf("parser has %d error", len(errors))
	} else {
		t.Errorf("parser has %d errors", len(errors))
	}

	for _, msg := range errors {
		t.Fatalf("parser error: %q", msg)
	}
}

func TestPrintStatement(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{`print 42`, 42},
		{`print "hello world"`, "hello world"},
		{`print 0.123456789`, 0.123456789},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserError(t, p)

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

		testLiteralExpression(t, printStmt.Value, tt.expectedValue)
	}
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) {
	t.Helper()

	switch v := expected.(type) {
	case int:
		testIntegerLiteral(t, exp, int64(v))
	case int64:
		testIntegerLiteral(t, exp, v)
	case float64:
		testFloatLiteral(t, exp, v)
	case string:
		testStringLiteral(t, exp, v)
	default:
		t.Fatalf("type of exp not handled. got=%T", exp)
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) {
	t.Helper()

	integer, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("il is not an *ast.IntegerLiteral. got=%T", il)
	}

	if integer.Value != value {
		t.Fatalf("integer.Value is not %d. got=%d", value, integer.Value)
	}

	if integer.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Fatalf("integer.TokenLiteral is not %d. got=%s", value, integer.TokenLiteral())
	}
}

func testFloatLiteral(t *testing.T, fl ast.Expression, value float64) {
	t.Helper()

	float, ok := fl.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("fl is not an *ast.FloatLiteral. got=%f", fl)
	}

	if float.Value != value {
		t.Fatalf("float.Value is not %f. got=%f", value, float.Value)
	}

	// XXX: the float formatting needs to match exactly number digits in the
	// fractional.
	if float.TokenLiteral() != fmt.Sprintf("%.9f", value) {
		t.Fatalf("float.TokenLiteral is not %.9f. got=%s", value, float.TokenLiteral())
	}
}

func testStringLiteral(t *testing.T, sl ast.Expression, value string) {
	t.Helper()

	str, ok := sl.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("sl is not an *ast.StringLiteral. got=%T", sl)
	}

	if str.Value != value {
		t.Fatalf("str.Value is not %s. got=%s", value, str.Value)
	}

	if str.TokenLiteral() != fmt.Sprintf("%s", value) {
		t.Fatalf("str.TokenLiteral is not %s. got=%s", value, str.TokenLiteral())
	}
}
