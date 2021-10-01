// Package parser implements a Top Down Operator Precedence parser also known
// as Pratt parser for the source language of the didactic compiler.
package parser

import (
	"fmt"
	"strconv"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/lexer"
	"github.com/Glorforidor/didactic_compiler/token"
	"github.com/Glorforidor/didactic_compiler/types"
)

type (
	prefixParseFunc func() ast.Expression
	infixParseFunc  func(ast.Expression) ast.Expression
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
		infixParseFuncs:  make(map[token.TokenType]infixParseFunc),
	}

	// register literals
	p.registerPrefixFunc(token.Int, p.parseIntegerLiteral)
	p.registerPrefixFunc(token.Float, p.parseFloatLiteral)
	p.registerPrefixFunc(token.String, p.parseStringLiteral)

	// register identifier
	p.registerPrefixFunc(token.Ident, p.parseIdentifier)

	// register operators
	p.registerInfixFunc(token.Plus, p.parseInfixExpression)
	p.registerInfixFunc(token.Minus, p.parseInfixExpression)
	p.registerInfixFunc(token.Asterisk, p.parseInfixExpression)
	p.registerInfixFunc(token.Slash, p.parseInfixExpression)

	// Prime the parser, so curToken and peekToken are in the right positions.
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek advances to next token if the next token is expected.
func (p *Parser) expectPeek(ts ...token.TokenType) bool {
	for _, t := range ts {
		if p.peekTokenIs(t) {
			p.nextToken()
			return true
		}
	}

	if len(ts) == 1 {
		p.errorf("expected next token to be %s, got: %s", ts, p.peekToken.Type)
	} else {
		p.errorf("expected next token to be on of %v, got: %s", ts, p.peekToken.Type)
	}
	return false
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

func (p *Parser) errorf(format string, a ...interface{}) {
	p.errors = append(p.errors, fmt.Sprintf(format, a...))
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
	case token.Var:
		return p.parseVarStatement()
	default:
		if p.curToken.Type == token.Ident && p.peekTokenIs(token.Assign) {
			return p.parseAssignStatement()
		}

		return p.parseExpressionStatement()
	}
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	stmt := &ast.PrintStatement{Token: p.curToken} // p.curtoken = "print"

	p.nextToken() // advance to the literal

	if p.curTokenIs(token.Eof) {
		p.errorf("invalid form for print statement - must have an argument")
		return nil
	}

	stmt.Value = p.parseExpression(Lowest)

	return stmt
}

func (p *Parser) parseVarStatement() *ast.VarStatement {
	stmt := &ast.VarStatement{Token: p.curToken}

	if !p.expectPeek(token.Ident) {
		return nil
	}

	id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.IntType, token.FloatType, token.StringType) {
		return nil
	}

	switch p.curToken.Type {
	case token.IntType:
		id.T = types.Type{Kind: types.Int}
	case token.FloatType:
		id.T = types.Type{Kind: types.Float}
	case token.StringType:
		id.T = types.Type{Kind: types.String}
	}

	stmt.Name = id

	if p.peekTokenIs(token.Assign) {
		p.nextToken() // "="
		p.nextToken() // the expression

		stmt.Value = p.parseExpression(Lowest)
	}

	return stmt
}

func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	stmt := &ast.AssignStatement{}
	// current token is on the identifier
	stmt.Name = p.parseExpression(Lowest).(*ast.Identifier)

	p.nextToken() // advance to the "="
	stmt.Token = p.curToken

	p.nextToken() // advance to the value
	stmt.Value = p.parseExpression(Lowest)

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(Lowest)

	return stmt
}

// parseExpression is the heart of the Pratt parsing.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFuncs[p.curToken.Type]
	if prefix == nil {
		p.errorf(
			"no prefix parse function attached to token %s found",
			p.curToken.Type,
		)
		return nil
	}

	leftExp := prefix()

	// TODO: maybe stop when reaching newline as all statements a newline
	// delimited and not semicolon.
	for precedence < p.peekPrecedence() {
		infix := p.infixParseFuncs[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		// Advance to next token so infix does not parse an already parsed
		// token.
		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

// Precendence table
const (
	_ int = iota
	Lowest
	Sum     // +
	Product // *
)

var precedences = map[token.TokenType]int{
	token.Plus:     Sum,
	token.Minus:    Sum,
	token.Asterisk: Product,
	token.Slash:    Product,
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return Lowest
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return Lowest
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.errorf("could not parse %q as an integer", p.curToken.Literal)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as a float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}
