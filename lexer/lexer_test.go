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
	(2 + 2)
	var x int
	var x2 int
	var y int = 2
	x = 2
	{}
	== !=
	<
	if 2 < 2 { } else { }
	true
	false
	var x bool
	for var i int = 0; i < 1; i = i + 1 { }
	func greet(x int) { }
	type human struct{name string}
	return 2
	x.name
	greet(2)
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
		{token.Lparen, "("},
		{token.Int, "2"},
		{token.Plus, "+"},
		{token.Int, "2"},
		{token.Rparen, ")"},
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
		{token.Lbrace, "{"},
		{token.Rbrace, "}"},
		{token.Equal, "=="},
		{token.NotEqual, "!="},
		{token.LessThan, "<"},
		{token.If, "if"},
		{token.Int, "2"},
		{token.LessThan, "<"},
		{token.Int, "2"},
		{token.Lbrace, "{"},
		{token.Rbrace, "}"},
		{token.Else, "else"},
		{token.Lbrace, "{"},
		{token.Rbrace, "}"},
		{token.True, "true"},
		{token.False, "false"},
		{token.Var, "var"},
		{token.Ident, "x"},
		{token.BoolType, "bool"},
		{token.For, "for"},
		{token.Var, "var"},
		{token.Ident, "i"},
		{token.IntType, "int"},
		{token.Assign, "="},
		{token.Int, "0"},
		{token.Semicolon, ";"},
		{token.Ident, "i"},
		{token.LessThan, "<"},
		{token.Int, "1"},
		{token.Semicolon, ";"},
		{token.Ident, "i"},
		{token.Assign, "="},
		{token.Ident, "i"},
		{token.Plus, "+"},
		{token.Int, "1"},
		{token.Lbrace, "{"},
		{token.Rbrace, "}"},
		{token.Func, "func"},
		{token.Ident, "greet"},
		{token.Lparen, "("},
		{token.Ident, "x"},
		{token.IntType, "int"},
		{token.Rparen, ")"},
		{token.Lbrace, "{"},
		{token.Rbrace, "}"},
		{token.Type, "type"},
		{token.Ident, "human"},
		{token.Struct, "struct"},
		{token.Lbrace, "{"},
		{token.Ident, "name"},
		{token.StringType, "string"},
		{token.Rbrace, "}"},
		{token.Return, "return"},
		{token.Int, "2"},
		{token.Ident, "x"},
		{token.Period, "."},
		{token.Ident, "name"},
		{token.Ident, "greet"},
		{token.Lparen, "("},
		{token.Int, "2"},
		{token.Rparen, ")"},
		{token.Eof, ""},
	}

	l := newTest(input)
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
