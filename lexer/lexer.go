// Package lexer implements a lexer (scanner) for the source language for the
// didactic compiler.
package lexer

import (
	"github.com/Glorforidor/didactic_compiler/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// readChar reads next character and advance positions accordingly.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = eof
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition += 1
}

const eof byte = 0

func (l *Lexer) NextToken() token.Token {
	l.skipWhiteSpace()

	var tok token.Token
	switch l.ch {
	case '"':
		tok.Type = token.String
		tok.Literal = l.readString()
	case eof:
		tok.Literal = ""
		tok.Type = token.Eof
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.Print

			// readIdentifier advances read positions, and therefore we must
			// return early so we do not advance further.
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.Int

			// readNumber advances read positions, and therefore we must
			// return early so we do not advance further.
			return tok
		} else {
			tok = token.Token{Type: token.Illegal, Literal: string(l.ch)}
		}
	}

	// Advance this pointers so l.ch is updated for the next invocation of
	// NextToken.
	l.readChar()

	return tok
}

// isLetter check whether ch is a letter. It includes "_" as a letter.
func isLetter(ch byte) bool {
	// Checking whether ch is a letter is done by checking if the byte is
	// within the ascii values of 'a'...'z' and 'A'...'Z'.
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit check whether ch is a number.
func isDigit(ch byte) bool {
	// Checking whether ch is a number is done by checking if the byte is
	// within the ascii values of '0'...'9'.
	return '0' <= ch && ch <= '9'
}

// readIdentifier reads an identifier from l.input.
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

// readNumber reads an number from l.input.
func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

// readString reads a string from l.input.
func (l *Lexer) readString() string {
	// readPosition is right after the '"' so inside the string.
	position := l.readPosition

	for {
		l.readChar()

		// TODO: consider if eof is reached the string is malformed.
		if l.ch == '"' || l.ch == eof {
			break
		}
	}

	return l.input[position:l.position]
}

func (l *Lexer) skipWhiteSpace() {
	// keep advancing the input positions until hitting a non whitespace
	// character.
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}
