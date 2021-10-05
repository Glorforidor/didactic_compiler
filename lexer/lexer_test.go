package lexer

import (
	"testing"

	"github.com/Glorforidor/didactic_compiler/token"
)

func TestNextToken(t *testing.T) {
	input := `
	print 42
	print "hello world"
	print 0.42
	-/*+
	var x int	
	var x2 int
	var y int = 2
	x = 2
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.Print, "print"},
		{token.Int, "42"},
		{token.Print, "print"},
		{token.String, "hello world"},
		{token.Print, "print"},
		{token.Float, "0.42"},
		{token.Minus, "-"},
		{token.Slash, "/"},
		{token.Asterisk, "*"},
		{token.Plus, "+"},
		{token.Var, "var"},
		{token.Ident, "x"},
		{token.IntType, "int"},
		{token.Var, "var"},
		{token.Ident, "x2"},
		{token.IntType, "int"},
		{token.Var, "var"},
		{token.Ident, "y"},
		{token.IntType, "int"},
		{token.Assign, "="},
		{token.Int, "2"},
		{token.Ident, "x"},
		{token.Assign, "="},
		{token.Int, "2"},
		{token.Eof, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i,
				tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i,
				tt.expectedLiteral, tok.Literal)
		}
	}
}
