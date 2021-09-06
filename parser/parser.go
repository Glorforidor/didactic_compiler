package parser

import (
	"fmt"
	"strconv"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/lexer"
	"github.com/Glorforidor/didactic_compiler/token"
)

type (
	prefixParseFunc func() ast.Expression
	infixParseFunc  func(ast.Expression) ast.Expression
)

// Precendence table
const (
	_ int = iota
	Lowest
)

// Parser holds the parser's internal state.
type Parser struct {
	l *lexer.Lexer

	errors []string

	curToken  token.Token
	peekToken token.Token

	// Pratt parsing maps token types with parsing functions.
	prefixParseFuncs map[token.TokenType]prefixParseFunc
	infixParseFuncs  map[token.TokenType]infixParseFunc
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:                l,
		prefixParseFuncs: make(map[token.TokenType]prefixParseFunc),
	}

	p.registerPrefixFunc(token.Int, p.parseIntegerLiteral)
	p.registerPrefixFunc(token.String, p.parseStringLiteral)

	// Prime the parser, so curToken and peekToken are in the right positions.
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) registerPrefixFunc(tt token.TokenType, f prefixParseFunc) {
	p.prefixParseFuncs[tt] = f
}

func (p *Parser) registerInfixFunc(tt token.TokenType, f infixParseFunc) {
	p.infixParseFuncs[tt] = f
}

func (p *Parser) Errors() []string {
	return p.errors
}

// ParseProgram parses the source language for the didactic compiler into an
// AST.
func (p *Parser) ParseProgram() *ast.Program {
	var program ast.Program

	for p.curToken.Type != token.Eof {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return &program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.Print:
		return p.parsePrintStatement()
	default:
		return nil
	}
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	stmt := &ast.PrintStatement{Token: p.curToken} // p.curtoken = "print"

	p.nextToken() // advance to the literal

	stmt.Value = p.parseExpression(Lowest)

	return stmt
}

// parseExpression is the heart of the Pratt parsing.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFuncs[p.curToken.Type]
	if prefix == nil {
		msg := fmt.Sprintf(
			"no prefix parse function attached to token %s found",
			p.curToken.Type,
		)
		p.errors = append(p.errors, msg)
		return nil
	}
	leftExp := prefix()

	return leftExp
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as an integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}
