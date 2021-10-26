package parser

import (
	"fmt"
	"testing"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/lexer"
	"github.com/Glorforidor/didactic_compiler/types"
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
	t.Helper()

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
		expectedType       types.Type
		expectedValue      interface{}
	}{
		{"var x int", "x", types.Typ[types.Int], nil},
		{"var x float", "x", types.Typ[types.Float], nil},
		{"var x string", "x", types.Typ[types.String], nil},
		{"var x bool", "x", types.Typ[types.Bool], nil},
		{"var x int = 1", "x", types.Typ[types.Int], 1},
		{"var x float = 1.0", "x", types.Typ[types.Float], 1.0},
		{`var x string = "Hello World"`, "x", types.Typ[types.String], "Hello World"},
		{"var x bool = true", "x", types.Typ[types.Bool], true},
		{"var x bool = false", "x", types.Typ[types.Bool], false},
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

		testVarStatement(t, stmt, tt.expectedIdentifier, tt.expectedType, tt.expectedValue)
	}
}

func testVarStatement(t *testing.T, stmt ast.Statement, id string, varType types.Type, value interface{}) {
	t.Helper()

	varStmt, ok := stmt.(*ast.VarStatement)
	if !ok {
		t.Fatalf("stmt not *ast.VarStatement. got=%T", stmt)
	}

	if varStmt.Name.Value != id {
		t.Fatalf("varStmt.Name.Value not %q, got=%q", id, varStmt.Name.Value)
	}

	if varStmt.Name.T != varType {
		t.Fatalf("varStmt.Name.T.Kind is not %T, got=%T", varType, varStmt.Name.T)
	}

	if value != nil {
		val := varStmt.Value
		testLiteralExpression(t, val, value)
	}
}

type infix struct {
	lhs      interface{}
	operator string
	rhs      interface{}
}

func TestAssignStatement(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"x = 2", "x", 2},
		{"x = 3.0", "x", 3.0},
		{`x = "Hello world`, "x", "Hello world"},
		{"x = 2 + 2", "x", infix{2, "+", 2}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserError(t, p)

		checkProgramLength(t, program)

		stmt := program.Statements[0]
		if stmt.TokenLiteral() != "=" {
			t.Fatalf("stmt.TokenLiteral not %q. got=%q", "=", stmt.TokenLiteral())
		}

		testAssignStatement(t, stmt, tt.expectedIdentifier, tt.expectedValue)
	}
}

func testAssignStatement(t *testing.T, stmt ast.Statement, id string, value interface{}) {
	t.Helper()

	assignStmt, ok := stmt.(*ast.AssignStatement)
	if !ok {
		t.Fatalf("stmt not *ast.AssignStatement, got=%T", stmt)
	}

	if assignStmt.Name.Value != id {
		t.Fatalf(
			"assignStmt.Name.Value not %q, got=%q",
			id, assignStmt.Name.Value,
		)
	}

	if inf, ok := value.(infix); ok {
		testInfixExpression(t, assignStmt.Value, inf.lhs, inf.operator, inf.rhs)
	} else {
		testLiteralExpression(t, assignStmt.Value, value)
	}
}

func TestBlockStatement(t *testing.T) {
	tests := []struct {
		input        string
		expectedSize int
	}{
		{
			input:        "{var x int = 2}",
			expectedSize: 1,
		},
		{
			input: `{
var x int = 2
var y int = 2
print x + y
}`,
			expectedSize: 3,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserError(t, p)

		checkProgramLength(t, program)

		stmt := program.Statements[0]
		if stmt.TokenLiteral() != "{" {
			t.Fatalf("stmt.TokenLiteral not %q. got=%q", "=", stmt.TokenLiteral())
		}

		blockStmt, ok := stmt.(*ast.BlockStatement)
		if !ok {
			t.Fatalf("stmt not *ast.BlockStatement, got=%T", stmt)
		}

		if len(blockStmt.Statements) != tt.expectedSize {
			t.Fatalf("blockStmt.Statements had the wrong size. expected=%v, got=%v",
				tt.expectedSize, len(blockStmt.Statements))
		}
	}
}

func TestIfStatement(t *testing.T) {
	input := `if x < y { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserError(t, p)

	checkProgramLength(t, program)

	ifStmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an *ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if ifStmt.TokenLiteral() != "if" {
		t.Fatalf("ifStmt.TokenLiteral not %q. got=%q", "if", ifStmt.TokenLiteral())
	}

	testInfixExpression(t, ifStmt.Condition, "x", "<", "y")

	consequence, ok := ifStmt.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("ifstmt.Consequence.Statements[0] is not an *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	testIdentifier(t, consequence.Expression, "x")

	if ifStmt.Alternative != nil {
		t.Fatalf("ifstmt.Alternative was not nil. got=%+v", ifStmt.Alternative)
	}
}

func TestIfStatementWithElse(t *testing.T) {
	input := `if x < y { x } else { y }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserError(t, p)

	checkProgramLength(t, program)

	ifStmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an *ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if ifStmt.TokenLiteral() != "if" {
		t.Fatalf("ifStmt.TokenLiteral not %q. got=%q", "if", ifStmt.TokenLiteral())
	}

	testInfixExpression(t, ifStmt.Condition, "x", "<", "y")

	consequence, ok := ifStmt.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("ifstmt.Consequence.Statements[0] is not an *ast.ExpressionStatement. got=%T",
			ifStmt.Consequence.Statements[0])
	}

	testIdentifier(t, consequence.Expression, "x")

	alternative, ok := ifStmt.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("ifstmt.Consequence.Statements[0] is not an *ast.ExpressionStatement. got=%T",
			ifStmt.Alternative.Statements[0])
	}

	testIdentifier(t, alternative.Expression, "y")
}

func TestForStatement(t *testing.T) {
	input := `for var i int = 0; i < 2; i = i + 1 { print i }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserError(t, p)
	checkProgramLength(t, program)

	forStmt, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an *ast.ForStatement. got=%T",
			program.Statements[0])
	}

	if forStmt.TokenLiteral() != "for" {
		t.Fatalf("forStmt.TokenLiteral not %q. got=%q", "for", forStmt.TokenLiteral())
	}

	testVarStatement(t, forStmt.Init, "i", types.Typ[types.Int], 0)

	testInfixExpression(t, forStmt.Condition, "i", "<", 2)

	testAssignStatement(t, forStmt.Next, "i", infix{"i", "+", 1})

	if len(forStmt.Body.Statements) != 1 {
		t.Fatalf("forStmt.Body.Statements had the wrong size. expected=%v, got=%v",
			1, len(forStmt.Body.Statements))
	}
}

func TestFuncStatement(t *testing.T) {
	tests := []struct {
		input             string
		expectedName      string
		expectedParam     string
		expectedParamType types.Type
		expectedResult    types.Type
	}{
		{
			input:             `func greet(s string) { }`,
			expectedName:      "greet",
			expectedParam:     "s",
			expectedParamType: types.Typ[types.String],
			expectedResult:    nil,
		},
		{
			input:             `func greet(s string) int { }`,
			expectedName:      "greet",
			expectedParam:     "s",
			expectedParamType: types.Typ[types.String],
			expectedResult:    types.Typ[types.Int],
		},
		{
			input:             `func greet(s string) float { }`,
			expectedName:      "greet",
			expectedParam:     "s",
			expectedParamType: types.Typ[types.String],
			expectedResult:    types.Typ[types.Float],
		},
		{
			input:             `func greet(s string) string { }`,
			expectedName:      "greet",
			expectedParam:     "s",
			expectedParamType: types.Typ[types.String],
			expectedResult:    types.Typ[types.String],
		},
		{
			input:             `func greet(s string) bool { }`,
			expectedName:      "greet",
			expectedParam:     "s",
			expectedParamType: types.Typ[types.String],
			expectedResult:    types.Typ[types.Bool],
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserError(t, p)
		checkProgramLength(t, program)

		funcStmt, ok := program.Statements[0].(*ast.FuncStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not an *ast.FuncStatement. got=%T",
				program.Statements[0])
		}

		if funcStmt.TokenLiteral() != "func" {
			t.Fatalf("funcStmt.TokenLiteral not %q. got=%q", "func", funcStmt.TokenLiteral())
		}

		if funcStmt.Name.Value != tt.expectedName {
			t.Fatalf("funcStmt.Name is not %s, got=%s", tt.expectedName, funcStmt.Name)
		}

		if funcStmt.Parameter.Value != tt.expectedParam {
			t.Fatalf("funcStmt.Parameter.Value is not %s, got=%s", tt.expectedParam, funcStmt.Parameter.Value)
		}

		if funcStmt.Parameter.T != tt.expectedParamType {
			t.Fatalf("funcStmt.Parameter.T is not %s, got=%s", tt.expectedParamType, funcStmt.Parameter.T)
		}

		if funcStmt.Result != nil && tt.expectedResult != nil && funcStmt.Result != tt.expectedResult {
			t.Fatalf("funcStmt.Result is not %v, got=%v", tt.expectedResult, funcStmt.Result)
		}

		if funcStmt.Result != nil && tt.expectedResult == nil {
			t.Fatalf("funcStmt.Result should be nil, got=%v", funcStmt.Result)
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
		{"x / 5", "x", "/", 5},
		{"2 < 2", 2, "<", 2},
		{"2 == 2", 2, "==", 2},
		{"2 != 2", 2, "!=", 2},
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
		{
			input:    "5 * (5 + 5) / 5",
			expected: "((5 * (5 + 5)) / 5)",
		},
		{
			input:    "1 < 2 == true",
			expected: "((1 < 2) == true)",
		},
		{
			input:    "1 < 1 == false",
			expected: "((1 < 1) == false)",
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
		if _, ok := expr.(*ast.Identifier); ok {
			testIdentifier(t, expr, v)
		} else {
			testStringLiteral(t, expr, v)
		}
	case bool:
		testBoolLiteral(t, expr, v)
	default:
		t.Fatalf("type of exp not handled. got=%T", expr)
	}
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) {
	t.Helper()

	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", exp)
	}

	if ident.Value != value {
		t.Fatalf("ident.Value is not %s. got=%s", value, ident.Value)
	}

	if ident.TokenLiteral() != value {
		t.Fatalf("ident.TokentLiteral is not %s. got=%s", value, ident.TokenLiteral())
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

	// Kinda dirty, but format %g will give the maximum decimals to represent
	// the float, the down side is that it strip trailing zeros, so 1.0 -> 1
	v := fmt.Sprintf("%g", value)
	if len(v) == 1 {
		v += ".0"
	}

	if float.TokenLiteral() != v {
		t.Fatalf("float.TokenLiteral is not %s. got=%s", v, float.TokenLiteral())
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

func testBoolLiteral(t *testing.T, bl ast.Expression, value bool) {
	t.Helper()

	bul, ok := bl.(*ast.BoolLiteral)
	if !ok {
		t.Fatalf("bl is not an *ast.BoolLiteral. got=%T", bl)
	}

	if bul.Value != value {
		t.Fatalf("bul.Value is not %v. got=%v", value, bul.Value)
	}

	if bul.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Fatalf("bul.TokenLiteral is not %t. got=%s", value, bul.TokenLiteral())
	}
}
