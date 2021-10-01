// Package ast declares the types that represent the syntax tree for the source
// language of the didactic compiler.
package ast

import (
	"strings"

	"github.com/Glorforidor/didactic_compiler/symbol"
	"github.com/Glorforidor/didactic_compiler/token"
	"github.com/Glorforidor/didactic_compiler/types"
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
	Type() types.Type

	// expressionNode is not stricly needed, but will guide the Go compiler to
	// error if a expression is used as an statement.
	expressionNode()
}

// Program is the root of the source language for the didactic compiler.
type Program struct {
	Statements  []Statement
	SymbolTable *symbol.Table
}

func (p *Program) TokenLiteral() string {
	if 0 < len(p.Statements) {
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

type VarStatement struct {
	Token token.Token // The token.Var token.
	Name  *Identifier
	Value Expression
}

func (vs *VarStatement) statementNode()       {}
func (vs *VarStatement) TokenLiteral() string { return vs.Token.Literal }
func (vs *VarStatement) String() string {
	var sb strings.Builder

	sb.WriteString(vs.TokenLiteral())
	sb.WriteString(" ")
	sb.WriteString(vs.Name.String())

	if vs.Value != nil {
		sb.WriteString(" = ")
		sb.WriteString(vs.Value.String())
	}

	return sb.String()
}

type AssignStatement struct {
	Token token.Token // The token.Assign token.
	Name  *Identifier
	Value Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignStatement) String() string {
	var sb strings.Builder

	sb.WriteString(as.Name.String())
	sb.WriteString(" ")
	sb.WriteString(as.TokenLiteral())
	sb.WriteString(" ")
	sb.WriteString(as.Value.String())

	return sb.String()
}

type Identifier struct {
	Token token.Token // The token.Ident token.
	Value string      // e.g. foo, bar, foobar
	Reg   int
	T     types.Type
}

func (id *Identifier) expressionNode()      {}
func (id *Identifier) Register() int        { return id.Reg }
func (id *Identifier) Type() types.Type     { return id.T }
func (id *Identifier) TokenLiteral() string { return id.Token.Literal }
func (id *Identifier) String() string {
	var sb strings.Builder

	sb.WriteString(id.Value)
	sb.WriteString(" ")
	sb.WriteString(id.T.Kind.String())

	return sb.String()
}

type IntegerLiteral struct {
	Token token.Token // The token.Int token.
	Value int64
	Reg   int
	T     types.Type
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) Register() int        { return il.Reg }
func (il *IntegerLiteral) Type() types.Type     { return il.T }
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token // The token.Float token.
	Value float64
	Reg   int
	T     types.Type
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) Register() int        { return fl.Reg }
func (fl *FloatLiteral) Type() types.Type     { return fl.T }
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type StringLiteral struct {
	Token token.Token // The token.String token.
	Value string
	Reg   int
	T     types.Type
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) Register() int        { return sl.Reg }
func (sl *StringLiteral) Type() types.Type     { return sl.T }
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

type InfixExpression struct {
	Token    token.Token // The operator token (+, -, /, *)
	Left     Expression
	Operator string
	Right    Expression
	Reg      int
	T        types.Type
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) Register() int        { return ie.Reg }
func (ie *InfixExpression) Type() types.Type     { return ie.T }
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
