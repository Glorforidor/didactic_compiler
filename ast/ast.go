// Package ast declares the types that represent the syntax tree for the source
// language of the didactic compiler.
package ast

import "github.com/Glorforidor/didactic_compiler/token"

type Node interface {
	// TokenLiteral is used for debugging and testing purpose.
	TokenLiteral() string
}

type Statement interface {
	Node

	// statementNode is not stricly needed, but will guide the Go compiler to
	// error if a statement is used as an expression.
	statementNode()
}

type Expression interface {
	Node

	// expressionNode is not stricly needed, but will guide the Go compiler to
	// error if a expression is used as an statement.
	expressionNode()
}

// Program is the root of the source language for the didactic compiler.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}

	return ""
}

type PrintStatement struct {
	Token token.Token // The token.Print token.
	Value Expression
}

func (ps *PrintStatement) statementNode()       {}
func (ps *PrintStatement) TokenLiteral() string { return ps.Token.Literal }
