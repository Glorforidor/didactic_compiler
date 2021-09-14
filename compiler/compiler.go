package compiler

import (
	"fmt"
	"strings"

	"github.com/Glorforidor/didactic_compiler/ast"
)

type Compiler struct {
	code []string

	registerTable registerTable
	label         label
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

	return &Compiler{registerTable: rt}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
	case *ast.PrintStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		// FIXME: When type checking comes in, then we add a type to an
		// expression we can switch on.
		var printType int
		switch node.Value.(type) {
		case *ast.IntegerLiteral:
			printType = 1
		case *ast.FloatLiteral:
			printType = 3
		case *ast.StringLiteral:
			printType = 4
		}

		code := []string{
			fmt.Sprintf("li a7, %d", printType),
			"ecall",
		}
		c.emit(code...)

		c.registerTable.dealloc(node.Value.Register())
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
		reg, err := c.registerTable.alloc()
		if err != nil {
			return err
		}

		// TODO: Should probably have a allocate floating point register func.
		node.Reg = reg

		// create a label for the floating point
		// TODO: maybe it would be nice to have different names for labels
		// after their type.
		c.label.Create()

		code := []string{
			".data",
			fmt.Sprintf("%s:", c.label.Name()),
			fmt.Sprintf(".double %g", node.Value),
			".text",
			// TODO: Change the below to allocate a float register
			fmt.Sprintf("fld fa0, %s, %s", c.label.Name(), c.registerTable.name(reg)), 
		}
		c.emit(code...)
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
