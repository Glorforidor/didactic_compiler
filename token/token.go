// Package token defines the tokens that makes up the source language for the
// didactic compiler.
package token

type TokenType string

const (
	Illegal TokenType = "ILLEGAL"
	Eof     TokenType = "EOF"

	// Literals
	Int    TokenType = "INT"    // 0123456789
	String TokenType = "STRING" // "hello world"
	Float  TokenType = "FLOAT"  // 0.123456789

	// Keywords
	Print TokenType = "PRINT"
)

type Token struct {
	Type    TokenType
	Literal string
}
