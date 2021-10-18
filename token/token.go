// Package token defines the tokens that makes up the source language for the
// didactic compiler.
package token

type TokenType string

const (
	// Misc
	Illegal TokenType = "ILLEGAL"
	Eof     TokenType = "EOF"

	// Identifier
	Ident TokenType = "IDENT" // foo, bar, foobar, x, y, z, ...

	// Literals
	Int    TokenType = "INT"    // 0123456789
	Float  TokenType = "FLOAT"  // 0.123456789
	String TokenType = "STRING" // "hello world"

	// Operators
	Plus     TokenType = "+"
	Minus    TokenType = "-"
	Asterisk TokenType = "*"
	Slash    TokenType = "/"
	Assign   TokenType = "="

	// Grouping
	Lparen TokenType = "("
	Rparen TokenType = ")"
	Lbrace TokenType = "{"
	Rbrace TokenType = "}"

	// Comparison operators
	Equal    TokenType = "=="
	NotEqual TokenType = "!="
	LessThan TokenType = "<"

	// Keywords
	Print      TokenType = "PRINT"
	Var        TokenType = "VAR"
	If         TokenType = "IF"
	Else       TokenType = "ELSE"
	IntType    TokenType = "INT_TYPE"
	FloatType  TokenType = "FLOAT_TYPE"
	StringType TokenType = "STRING_TYPE"
	BoolType   TokenType = "BOOL_TYPE"
	True       TokenType = "TRUE"
	False      TokenType = "FALSE"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"print":  Print,
	"var":    Var,
	"if":     If,
	"else":   Else,
	"int":    IntType,
	"float":  FloatType,
	"string": StringType,
	"bool":   BoolType,
	"true":   True,
	"false":  False,
}

// LookupIdentifier checks if the identifier is a keyword, and if so returns
// that keyword TokenType. Otherwise returns Ident TokenType.
func LookupIdentifier(id string) TokenType {
	if v, ok := keywords[id]; ok {
		return v
	}

	return Ident
}
