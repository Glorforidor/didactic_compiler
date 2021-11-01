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
	Register() string

	// Type returns the expressions type.
	Type() types.Type

	// expressionNode is not stricly needed, but will guide the Go compiler to
	// error if a expression is used as an statement.
	expressionNode()
}

type TypeNode interface {
	Node

	typeNode()
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

type TypeStatement struct {
	Token token.Token // The token.Type token.
	Name  *Identifier
	Type  TypeNode
}

func (ts *TypeStatement) statementNode()       {}
func (ts *TypeStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TypeStatement) String() string {
	var sb strings.Builder

	sb.WriteString(ts.Token.Literal)
	sb.WriteString(" ")
	sb.WriteString(ts.Name.String())
	sb.WriteString(" ")
	if ts.Name.T == nil {
		sb.WriteString(ts.Type.String())
	}

	return sb.String()
}

type BlockStatement struct {
	Token       token.Token // The token.Lcurly token.
	Statements  []Statement
	SymbolTable *symbol.Table
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var sb strings.Builder

	sb.WriteString("{")
	for _, s := range bs.Statements {
		sb.WriteString(s.String())
	}
	sb.WriteString("}")

	return sb.String()
}

type IfStatement struct {
	Token       token.Token // The token.If token.
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ifs *IfStatement) statementNode()       {}
func (ifs *IfStatement) TokenLiteral() string { return ifs.Token.Literal }
func (ifs *IfStatement) String() string {
	var sb strings.Builder

	sb.WriteString("if")
	sb.WriteString(ifs.Condition.String())
	sb.WriteString(" ")
	sb.WriteString(ifs.Consequence.String())

	if ifs.Alternative != nil {
		sb.WriteString("else ")
		sb.WriteString(ifs.Alternative.String())
	}

	return sb.String()
}

type ForStatement struct {
	Token     token.Token // The token.For token.
	Init      Statement
	Condition Expression
	Next      Statement
	Body      *BlockStatement

	SymbolTable *symbol.Table
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var sb strings.Builder

	sb.WriteString("for")
	sb.WriteString(" ")
	sb.WriteString(fs.Init.String())
	sb.WriteString("; ")
	sb.WriteString(fs.Condition.String())
	sb.WriteString("; ")
	sb.WriteString(fs.Next.String())
	sb.WriteString(fs.Body.String())

	return sb.String()
}

type FuncStatement struct {
	Token     token.Token // The token.Func token.
	Name      *Identifier
	Parameter *Identifier
	Result    token.Token // Type token: token.IntType, token.FloatType, token.StringType, token.BoolType, token.Ident
	Body      *BlockStatement

	SymbolTable *symbol.Table
}

func (fs *FuncStatement) statementNode()       {}
func (fs *FuncStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FuncStatement) String() string {
	var sb strings.Builder

	sb.WriteString("func")
	sb.WriteString(" ")
	sb.WriteString(fs.Name.String())
	sb.WriteString("(")
	sb.WriteString(fs.Parameter.String())
	sb.WriteString(")")
	sb.WriteString(" ")
	if fs.Result.Literal != "" {
		sb.WriteString(fs.Result.Literal)
		sb.WriteString(" ")
	}
	sb.WriteString(fs.Body.String())

	return sb.String()
}

type ReturnStatement struct {
	Token token.Token // The token.Return token.
	Value Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var sb strings.Builder

	sb.WriteString(rs.Token.Literal)
	sb.WriteString(" ")

	if rs.Value != nil {
		sb.WriteString(rs.Value.String())
	}

	return sb.String()
}

type Identifier struct {
	Token  token.Token // The token.Ident token.
	Value  string      // e.g. foo, bar, foobar
	Ttoken token.Token // Type token: token.IntType, token.FloatType, token.StringType, token.BoolType, token.Ident

	Reg string
	T   types.Type
}

func (id *Identifier) expressionNode()      {}
func (id *Identifier) Register() string     { return id.Reg }
func (id *Identifier) Type() types.Type     { return id.T }
func (id *Identifier) TokenLiteral() string { return id.Token.Literal }
func (id *Identifier) String() string {
	var sb strings.Builder

	sb.WriteString(id.Value)
	if id.T != nil {
		sb.WriteString(" ")
		sb.WriteString(id.T.String())
	}
	return sb.String()
}

type IntegerLiteral struct {
	Token token.Token // The token.Int token.
	Value int64
	Reg   string
	T     types.Type
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) Register() string     { return il.Reg }
func (il *IntegerLiteral) Type() types.Type     { return il.T }
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token // The token.Float token.
	Value float64
	Reg   string
	T     types.Type
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) Register() string     { return fl.Reg }
func (fl *FloatLiteral) Type() types.Type     { return fl.T }
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type StringLiteral struct {
	Token token.Token // The token.String token.
	Value string
	Reg   string
	T     types.Type
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) Register() string     { return sl.Reg }
func (sl *StringLiteral) Type() types.Type     { return sl.T }
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

type BoolLiteral struct {
	Token token.Token // The token.Bool token.
	Value bool
	Reg   string
	T     types.Type
}

func (bl *BoolLiteral) expressionNode()      {}
func (bl *BoolLiteral) Register() string     { return bl.Reg }
func (bl *BoolLiteral) Type() types.Type     { return bl.T }
func (bl *BoolLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BoolLiteral) String() string       { return bl.Token.Literal }

type InfixExpression struct {
	Token    token.Token // The operator token (+, -, /, *)
	Left     Expression
	Operator string
	Right    Expression
	Reg      string
	T        types.Type
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) Register() string     { return ie.Reg }
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

type StructType struct {
	Token  token.Token // The token.Struct token.
	Fields []*Identifier
}

func (st *StructType) typeNode()            {}
func (st *StructType) TokenLiteral() string { return st.Token.Literal }
func (st *StructType) String() string {
	var sb strings.Builder

	sb.WriteString(st.Token.Literal)
	sb.WriteString("{")
	for i, field := range st.Fields {
		sb.WriteString(field.String())

		if i == len(st.Fields)-1 {
			break
		}

		sb.WriteString(";")
	}
	sb.WriteString("}")

	return sb.String()
}
