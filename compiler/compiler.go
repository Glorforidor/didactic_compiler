package compiler

import (
	"fmt"
	"strings"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/types"
)

type Compiler struct {
	code []string

	registerTable      registerTable
	registerFloatTable registerTable
	label              label
}

func New() *Compiler {
	rt := registerTable{
		{name: "a0"},
		{name: "a1"},
		{name: "a2"},
		{name: "a3"},
		{name: "a4"},
		{name: "a5"},
		{name: "a6"},
		{name: "a7"},
	}

	ft := registerTable{
		{name: "fa0"},
		{name: "fa1"},
		{name: "fa2"},
		{name: "fa3"},
		{name: "fa4"},
		{name: "fa5"},
		{name: "fa6"},
		{name: "fa7"},
	}

	return &Compiler{registerTable: rt, registerFloatTable: ft}
}

func (c *Compiler) Compile(node ast.Node) error {
	if err := types.Checker(node); err != nil {
		return err
	}

	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		if err := c.Compile(node.Expression); err != nil {
			return err
		}

		reg := node.Expression.Register()
		switch node.Expression.Type().Kind {
		case ast.Float:
			c.registerFloatTable.dealloc(reg)
		default:
			c.registerTable.dealloc(reg)
		}
	case *ast.PrintStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		var printType int
		switch node.Value.Type().Kind {
		case ast.Int:
			printType = 1
		case ast.Float:
			printType = 3
		case ast.String:
			printType = 4
		default:
			return fmt.Errorf("compile error: can not print type: %q", node.Value.Type().Kind)
		}

		code := []string{
			fmt.Sprintf("li a7, %d", printType),
			"ecall",
		}
		c.emit(code...)

		c.registerTable.dealloc(node.Value.Register())
	case *ast.InfixExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}

		if err := c.Compile(node.Right); err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			code := []string{
				fmt.Sprintf(
					"add %s, %s, %s",
					c.registerTable.name(node.Left.Register()),
					c.registerTable.name(node.Left.Register()),
					c.registerTable.name(node.Right.Register()),
				),
			}
			c.emit(code...)
		}
		c.registerTable.dealloc(node.Right.Register())
	case *ast.IntegerLiteral:
		reg, err := c.registerTable.alloc()
		if err != nil {
			return err
		}

		node.Reg = reg
		code := []string{
			fmt.Sprintf("li %s, %d", c.registerTable.name(node.Reg), node.Value),
		}
		c.emit(code...)
	case *ast.FloatLiteral:
		reg, err := c.registerFloatTable.alloc()
		if err != nil {
			return err
		}

		node.Reg = reg

		// register a temporay register for the fld instruction.
		reg, err = c.registerTable.alloc()
		if err != nil {
			return err
		}

		// create a label for the floating point
		// TODO: maybe it would be nice to have different names for labels
		// after their type.
		c.label.Create()

		code := []string{
			".data",
			fmt.Sprintf("%s:", c.label.Name()),
			fmt.Sprintf(".double %g", node.Value),
			".text",
			fmt.Sprintf(
				"fld %s, %s, %s",
				c.registerFloatTable.name(node.Reg),
				c.label.Name(),
				c.registerTable.name(reg),
			),
		}
		c.emit(code...)

		// dealloc the temporay register
		c.registerTable.dealloc(reg)
	case *ast.StringLiteral:
		reg, err := c.registerTable.alloc()
		if err != nil {
			return err
		}

		node.Reg = reg
		c.label.Create()

		code := []string{
			".data",
			fmt.Sprintf("%s:", c.label.Name()),
			fmt.Sprintf(".string %q", node.Value),
			".text",
			fmt.Sprintf("la %s, %s", c.registerTable.name(node.Reg), c.label.Name()),
		}

		c.emit(code...)
	default:
		return fmt.Errorf("unknown type: %#v", node)
	}

	return nil
}

func (c *Compiler) emit(code ...string) {
	c.code = append(c.code, code...)
}

func (c *Compiler) Asm() string {
	var sb strings.Builder

	for i, t := range c.code {
		sb.WriteString(t)
		if i != len(c.code)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
