// Package parser implements a Top Down Operator Precedence parser also known
// as Pratt parser for the source language of the didactic compiler.
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
	p.registerPrefixFunc(token.True, p.parseBoolLiteral)
	p.registerPrefixFunc(token.False, p.parseBoolLiteral)

	// register identifier
	p.registerPrefixFunc(token.Ident, p.parseIdentifier)

	// register grouping
	p.registerPrefixFunc(token.Lparen, p.parseGroupedExpression)

	// register operators
	p.registerInfixFunc(token.Plus, p.parseInfixExpression)
	p.registerInfixFunc(token.Minus, p.parseInfixExpression)
	p.registerInfixFunc(token.Asterisk, p.parseInfixExpression)
	p.registerInfixFunc(token.Slash, p.parseInfixExpression)
	p.registerInfixFunc(token.Equal, p.parseInfixExpression)
	p.registerInfixFunc(token.NotEqual, p.parseInfixExpression)
	p.registerInfixFunc(token.LessThan, p.parseInfixExpression)

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

func (p *Parser) peekTokenIs(ts ...token.TokenType) bool {
	for _, t := range ts {
		if p.peekToken.Type == t {
			return true
		}
	}

	return false
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
		p.errorf("expected next token to be %q, got: %q", ts, p.peekToken)
	} else {
		p.errorf("expected next token to be one of %q, got: %q", ts, p.peekToken)
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
		v := p.parseVarStatement()
		if !p.expectPeek(token.Semicolon) {
			return nil
		}

		return v
	case token.Type:
		return p.parseTypeStatement()
	case token.Lbrace:
		block := p.parseBlockStatement()
		if !p.expectPeek(token.Semicolon) {
			return nil
		}
		return block
	case token.If:
		return p.parseIfStatement()
	case token.For:
		return p.parseForStatement()
	case token.Func:
		return p.parseFuncStatement()
	case token.Return:
		return p.parseReturnStatement()
	default:
		if p.curToken.Type == token.Ident && p.peekTokenIs(token.Assign) {
			assign := p.parseAssignStatement()
			if !p.expectPeek(token.Semicolon) {
				return nil
			}

			return assign
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

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseVarStatement() *ast.VarStatement {
	stmt := &ast.VarStatement{Token: p.curToken}

	if !p.expectPeek(token.Ident) {
		return nil
	}

	id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.IntType, token.FloatType, token.StringType, token.BoolType, token.Ident) {
		return nil
	}

	id.Ttoken = p.curToken

	stmt.Name = id

	if p.peekTokenIs(token.Assign) {
		p.nextToken() // "="
		p.nextToken() // the expression

		stmt.Value = p.parseExpression(Lowest)
	}

	return stmt
}

func (p *Parser) parseTypeStatement() *ast.TypeStatement {
	stmt := &ast.TypeStatement{Token: p.curToken}

	if !p.expectPeek(token.Ident) {
		return nil
	}

	id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.Struct) {
		return nil
	}

	id.Ttoken = p.curToken

	stmt.Name = id

	stmt.Type = p.parseStructType()

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseStructType() ast.TypeNode {
	st := &ast.StructType{
		Token:  p.curToken,
		Fields: []*ast.Identifier{},
	}

	p.nextToken() // advance to '{'

	if p.peekTokenIs(token.Rbrace) {
		p.nextToken() // advance to '}'
		return st
	}

	if !p.expectPeek(token.Ident) {
		return nil
	}

	id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.IntType, token.FloatType, token.StringType, token.BoolType, token.Ident) {
		return nil
	}

	id.Ttoken = p.curToken

	st.Fields = append(st.Fields, id)
	for p.peekTokenIs(token.Semicolon, token.Ident) {
		if p.peekTokenIs(token.Semicolon) {
			p.nextToken() // the semicolon

			if p.peekTokenIs(token.Rbrace) {
				break
			}
		}
		p.nextToken() // ident

		id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(token.IntType, token.FloatType, token.StringType, token.BoolType, token.Ident) {
			return nil
		}

		id.Ttoken = p.curToken

		st.Fields = append(st.Fields, id)
	}

	if !p.expectPeek(token.Rbrace) {
		return nil
	}

	return st
}

func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	var stmt ast.AssignStatement
	// current token is on the identifier
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken() // advance to the "="
	stmt.Token = p.curToken

	p.nextToken() // advance to the value
	stmt.Value = p.parseExpression(Lowest)

	return &stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}

	p.nextToken() // advance beyond "{"

	for !p.curTokenIs(token.Rbrace) && !p.curTokenIs(token.Eof) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()
	}

	if p.curTokenIs(token.Eof) {
		p.errorf("parser error: block statement was never closed")
		return nil
	}

	return block
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	p.nextToken()

	stmt.Condition = p.parseExpression(Lowest)

	if !p.expectPeek(token.Lbrace) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.Else) {
		p.nextToken()

		if !p.expectPeek(token.Lbrace) {
			return nil
		}

		stmt.Alternative = p.parseBlockStatement()
	}

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	p.nextToken()

	if p.curTokenIs(token.Var) {
		stmt.Init = p.parseVarStatement()
	} else if p.curTokenIs(token.Ident) && p.peekTokenIs(token.Assign) {
		stmt.Init = p.parseAssignStatement()
	} else {
		// TODO: maybe allow that the init phase can be omitted.
		p.errorf("The initialisation in a for statement can either be a var statement or an assign statement.")
		return nil
	}

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	p.nextToken() // advance beyond ;

	stmt.Condition = p.parseExpression(Lowest)

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	p.nextToken() // advance beyond ;

	if p.curTokenIs(token.Ident) && p.peekTokenIs(token.Assign) {
		stmt.Next = p.parseAssignStatement()
	} else {
		// TODO: maybe allow that the next phase can be omitted.
		p.errorf("The next in a for statement can only be an assign statement.")
		return nil
	}

	if !p.expectPeek(token.Lbrace) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseFuncStatement() *ast.FuncStatement {
	stmt := &ast.FuncStatement{Token: p.curToken}

	if !p.expectPeek(token.Ident) {
		return nil
	}

	stmt.Name = &ast.Identifier{
		Token:  p.curToken,
		Value:  p.curToken.Literal,
		Ttoken: stmt.Token,
	}

	if !p.expectPeek(token.Lparen) {
		return nil
	}

	stmt.Parameter = p.parseFuncParameter()

	if p.peekTokenIs(token.IntType, token.FloatType, token.StringType, token.BoolType, token.Ident) {
		p.nextToken() // advance to type

		stmt.Result = p.curToken
	}

	if !p.expectPeek(token.Lbrace) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseFuncParameter() *ast.Identifier {
	// Early return if there is no parameter.
	if p.peekTokenIs(token.Rparen) {
		p.nextToken()
		return nil
	}

	p.nextToken() // The identifier.

	id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.IntType, token.FloatType, token.StringType, token.BoolType, token.Ident) {
		return nil
	}

	id.Ttoken = p.curToken

	if !p.expectPeek(token.Rparen) {
		return nil
	}

	return id
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	if p.curTokenIs(token.Semicolon) {
		return stmt
	}

	stmt.Value = p.parseExpression(Lowest)

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(Lowest)

	if p.peekTokenIs(token.Semicolon) {
		p.nextToken()
	}

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

	for !p.peekTokenIs(token.Semicolon) && precedence < p.peekPrecedence() {
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
	Equals  // ==
	Less    // <
	Sum     // +
	Product // *
)

var precedences = map[token.TokenType]int{
	token.Equal:    Equals,
	token.NotEqual: Equals,
	token.LessThan: Less,
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

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // advance beyound the "("

	exp := p.parseExpression(Lowest)

	if !p.expectPeek(token.Rparen) {
		return nil
	}

	return exp
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

func (p *Parser) parseBoolLiteral() ast.Expression {
	return &ast.BoolLiteral{Token: p.curToken, Value: p.curTokenIs(token.True)}
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}
