// Package token defines the tokens that makes up the source language for the
// didactic compiler.
package token

type TokenType string

const (
	// Misc
	Illegal TokenType = "ILLEGAL"
	Eof     TokenType = "EOF"
	Blank   TokenType = "_"

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

	// Selector
	Period TokenType = "."

	// Delimiters
	Semicolon TokenType = ";"

	// Comparison operators
	Equal    TokenType = "=="
	NotEqual TokenType = "!="
	LessThan TokenType = "<"

	// Keywords
	Print      TokenType = "PRINT"
	Var        TokenType = "VAR"
	Type       TokenType = "TYPE"
	For        TokenType = "FOR"
	If         TokenType = "IF"
	Else       TokenType = "ELSE"
	Func       TokenType = "FUNC"
	Return     TokenType = "RETURN"
	Struct     TokenType = "STRUCT"
	IntType    TokenType = "INT_TYPE"
	FloatType  TokenType = "FLOAT_TYPE"
	StringType TokenType = "STRING_TYPE"
	BoolType   TokenType = "BOOL_TYPE"
	True       TokenType = "TRUE"
	False      TokenType = "FALSE"
)

// NOTE: could add pos and end to token so error messages later could add
// information about where in the source file the error occured.

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"print":  Print,
	"var":    Var,
	"type":   Type,
	"for":    For,
	"if":     If,
	"else":   Else,
	"func":   Func,
	"return": Return,
	"struct": Struct,
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
