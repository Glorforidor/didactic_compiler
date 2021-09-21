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

func checkProgramLength(t *testing.T, program *ast.Program) {
	if len(program.Statements) != 1 {
		t.Fatalf(
			"program.Statements does not contain 1 statement. got=%d",
			len(program.Statements),
		)
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

		checkProgramLength(t, program)

		stmt := program.Statements[0]
		if stmt.TokenLiteral() != "print" {
			t.Fatalf("stmt.TokenLiteral not %q. got=%q", "print", stmt.TokenLiteral())
		}

		printStmt, ok := stmt.(*ast.PrintStatement)
		if !ok {
			t.Fatalf("stmt not *ast.PrintStatement. got=%T", stmt)
		}

		if printStmt.TokenLiteral() != "print" {
			t.Fatalf("printStmt.TokenLiteral not %q, got=%q", "print", printStmt.TokenLiteral())
		}

		testLiteralExpression(t, printStmt.Value, tt.expectedValue)
	}
}

func TestVarStatement(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedType       ast.Type
		expectedValue      interface{}
	}{
		{"var x int", "x", ast.Type{ast.Int}, nil},
		{"var x float", "x", ast.Type{ast.Float}, nil},
		{"var x string", "x", ast.Type{ast.String}, nil},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserError(t, p)

		checkProgramLength(t, program)

		stmt := program.Statements[0]
		if stmt.TokenLiteral() != "var" {
			t.Fatalf("stmt.TokenLiteral not %q. got=%q", "var", stmt.TokenLiteral())
		}

		varStmt, ok := stmt.(*ast.VarStatement)
		if !ok {
			t.Fatalf("stmt not *ast.VarStatement. got=%T", stmt)
		}

		if varStmt.Name.Value != tt.expectedIdentifier {
			t.Fatalf("varStmt.Name.Value not %q, got=%q", tt.expectedIdentifier, varStmt.Name.Value)
		}

		if varStmt.Name.T.Kind != tt.expectedType.Kind {
			t.Fatalf("varStmt.Name.T.Kind is not %T, got=%T", tt.expectedType, varStmt.Name.T.Kind)
		}

		if tt.expectedValue != nil {
			val := varStmt.Value
			testLiteralExpression(t, val, tt.expectedValue)
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserError(t, p)

		checkProgramLength(t, program)

		exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0],
			)
		}

		testInfixExpression(t, exprStmt.Expression, tt.leftValue, tt.operator, tt.rightValue)
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "5 + 5 - 5",
			expected: "((5 + 5) - 5)",
		},
		{
			input:    "5 + 5 * 2",
			expected: "(5 + (5 * 2))",
		},
		{
			input:    "5 + 5 / 2 * 3 - 2",
			expected: "((5 + ((5 / 2) * 3)) - 2)",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserError(t, p)

		got := program.String()
		if got != tt.expected {
			t.Fatalf("expected=%q, got=%q", tt.expected, got)
		}
	}
}

func testInfixExpression(
	t *testing.T, expr ast.Expression,
	left interface{}, operator string, right interface{},
) {
	t.Helper()

	opExpr, ok := expr.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expr is not ast.InfixExpression. got=%T(%s)", expr, expr)
	}

	testLiteralExpression(t, opExpr.Left, left)

	if opExpr.Operator != operator {
		t.Fatalf("opExpr.Operator is not %q. got=%s", operator, opExpr.Operator)
	}

	testLiteralExpression(t, opExpr.Right, right)
}

func testLiteralExpression(
	t *testing.T,
	expr ast.Expression,
	expected interface{},
) {
	t.Helper()

	switch v := expected.(type) {
	case int:
		testIntegerLiteral(t, expr, int64(v))
	case int64:
		testIntegerLiteral(t, expr, v)
	case float64:
		testFloatLiteral(t, expr, v)
	case string:
		testStringLiteral(t, expr, v)
	default:
		t.Fatalf("type of exp not handled. got=%T", expr)
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
