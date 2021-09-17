// Package ast declares the types that represent the syntax tree for the source
// language of the didactic compiler.
package ast

import (
	"strings"

	"github.com/Glorforidor/didactic_compiler/token"
)

type Node interface {
	// TokenLiteral is used for debugging and testing purpose.
	TokenLiteral() string

	// String stringifies a nodes structure.
	String() string
}

type Statement interface {
	Node

	// statementNode is not stricly needed, but will guide the Go compiler to
	// error if a statement is used as an expression.
	statementNode()
}

type Expression interface {
	Node

	// Register returns a register number.
	Register() int

	// Type returns the expressions type.
	Type() Type

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

func (p *Program) String() string {
	var sb strings.Builder

	for _, s := range p.Statements {
		sb.WriteString(s.String())
	}

	return sb.String()
}

type ExpressionStatement struct {
	Token      token.Token // The first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}

type PrintStatement struct {
	Token token.Token // The token.Print token.
	Value Expression
}

func (ps *PrintStatement) statementNode()       {}
func (ps *PrintStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PrintStatement) String() string {
	var sb strings.Builder

	sb.WriteString(ps.TokenLiteral())
	sb.WriteString(" ")
	sb.WriteString(ps.Value.String())

	return sb.String()
}

type IntegerLiteral struct {
	Token token.Token // The token.Int token.
	Value int64
	Reg   int
	T     Type
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) Register() int        { return il.Reg }
func (il *IntegerLiteral) Type() Type           { return il.T }
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token // The token.Float token.
	Value float64
	Reg   int
	T     Type
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) Register() int        { return fl.Reg }
func (fl *FloatLiteral) Type() Type           { return fl.T }
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type StringLiteral struct {
	Token token.Token // The token.String token.
	Value string
	Reg   int
	T     Type
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) Register() int        { return sl.Reg }
func (sl *StringLiteral) Type() Type           { return sl.T }
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

type InfixExpression struct {
	Token    token.Token // The operator token (+, -, /, *)
	Left     Expression
	Operator string
	Right    Expression
	Reg      int
	T        Type
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) Register() int        { return ie.Reg }
func (ie *InfixExpression) Type() Type           { return ie.T }
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var sb strings.Builder

	sb.WriteString("(")
	sb.WriteString(ie.Left.String())
	sb.WriteString(" " + ie.Operator + " ")
	sb.WriteString(ie.Right.String())
	sb.WriteString(")")

	return sb.String()
}
