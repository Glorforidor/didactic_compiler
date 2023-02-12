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

	dontInsertSemi bool // used to skip semicolons in test
	insertSemi     bool // insert semicolon on new line

	line   int
	column int
}

// TODO: copy Go's way of inserting semicolons on newline (which is basically
// just create a token.Semicolon with the literal value of '\n')

func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

func newTest(input string) *Lexer {
	l := &Lexer{input: input, dontInsertSemi: true}
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

	if l.ch == '\n' {
		l.column = 0
		l.line++
	}

	l.column++
	l.position = l.readPosition
	l.readPosition++
}

const (
	eof     byte = 0
	newline byte = '\n'
)

func newToken(tokenType token.TokenType, ch byte, position token.Position) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Position: position}
}

func (l *Lexer) makeTwoCharToken(t token.TokenType) token.Token {
	pos := token.Position{Row: l.line, Col: l.column}
	ch := l.ch
	l.readChar()
	literal := string(ch) + string(l.ch)

	return token.Token{Type: t, Literal: literal, Position: pos}
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhiteSpace()

	for l.ch == '/' && l.peek() == '/' {
		l.skipComment()
	}

	var insertSemi bool

	position := token.Position{Row: l.line, Col: l.column}
	var tok token.Token
	switch l.ch {
	case newline:
		// We reach here, if l.insertSemi was true so l.skipWhiteSpace did not
		// skip the newline.
		l.insertSemi = false
		tok = newToken(token.Semicolon, l.ch, position)
	case '+':
		tok = newToken(token.Plus, l.ch, position)
	case '-':
		tok = newToken(token.Minus, l.ch, position)
	case '*':
		tok = newToken(token.Asterisk, l.ch, position)
	case '/':
		tok = newToken(token.Slash, l.ch, position)
	case '!':
		if l.peek() == '=' {
			tok = l.makeTwoCharToken(token.NotEqual)
		}
	case '=':
		if l.peek() == '=' {
			tok = l.makeTwoCharToken(token.Equal)
		} else {
			tok = newToken(token.Assign, l.ch, position)
		}
	case '<':
		tok = newToken(token.LessThan, l.ch, position)
	case '(':
		tok = newToken(token.Lparen, l.ch, position)
	case ')':
		tok = newToken(token.Rparen, l.ch, position)
		insertSemi = true
	case '{':
		tok = newToken(token.Lbrace, l.ch, position)
	case '}':
		tok = newToken(token.Rbrace, l.ch, position)
		insertSemi = true
	case ';':
		tok = newToken(token.Semicolon, l.ch, position)
	case '"':
		tok.Type = token.String
		tok.Literal = l.readString()
		insertSemi = true
	case '.':
		tok = newToken(token.Period, l.ch, position)
	case eof:
		if l.insertSemi {
			l.insertSemi = false
			return newToken(token.Semicolon, '\n', position)
		}

		tok.Type = token.Eof
		tok.Literal = ""
	default:
		if isLetter(l.ch) {
			tok.Position = position
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdentifier(tok.Literal)

			insertSemi = true

			if !l.dontInsertSemi {
				l.insertSemi = insertSemi
			}
			// readIdentifier advances read positions, and therefore we must
			// return early so we do not advance further.
			return tok
		} else if isDigit(l.ch) {
			tok.Position = position
			tok = l.readNumber()

			insertSemi = true

			if !l.dontInsertSemi {
				l.insertSemi = insertSemi
			}
			// readNumber advances read positions, and therefore we must
			// return early so we do not advance further.
			return tok
		} else {
			tok = token.Token{Type: token.Illegal, Literal: string(l.ch), Position: position}
		}
	}

	if !l.dontInsertSemi {
		l.insertSemi = insertSemi
	}

	// Advance this pointers so l.ch is updated for the next invocation of
	// NextToken.

	l.readChar()

	return tok
}

// peek peeks one character ahead in l.input.
func (l *Lexer) peek() byte {
	if l.readPosition >= len(l.input) {
		return eof
	}

	return l.input[l.readPosition]
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
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

// readNumber reads an number from l.input.
func (l *Lexer) readNumber() token.Token {
	var tok token.Token
	tok.Type = token.Int
	tok.Position = token.Position{Row: l.line, Col: l.column}

	position := l.position

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peek()) {
		tok.Type = token.Float

		l.readChar() // advance beyond the "."

		for isDigit(l.ch) {
			l.readChar()
		}
	}

	tok.Literal = l.input[position:l.position]

	return tok
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

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != '\r' {
		l.readChar()
	}

	// remove whitespace after the comment, otherwise it will treat whitespace
	// as tokens that needs to be lexed, which results in illegal tokens.
	l.skipWhiteSpace()
}

func (l *Lexer) skipWhiteSpace() {
	// keep advancing the input positions until hitting a non whitespace
	// character.
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' && !l.insertSemi || l.ch == '\r' {
		l.readChar()
	}
}
