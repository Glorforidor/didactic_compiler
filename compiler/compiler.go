package compiler

import (
	"fmt"
	"strings"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/symbol"
	"github.com/Glorforidor/didactic_compiler/types"
)

type Compiler struct {
	constants []string
	code      []string

	symbolTable        *symbol.Table
	registerTable      registerTable
	registerFloatTable registerTable
	label              label
}

func New() *Compiler {
	rt := registerTable{
		{name: "t0"},
		{name: "t1"},
		{name: "t2"},
		{name: "t3"},
		{name: "t4"},
		{name: "t5"},
		{name: "t6"},
		{name: "t7"},
	}

	ft := registerTable{
		{name: "ft0"},
		{name: "ft1"},
		{name: "ft2"},
		{name: "ft3"},
		{name: "ft4"},
		{name: "ft5"},
		{name: "ft6"},
		{name: "ft7"},
	}

	return &Compiler{
		registerTable:      rt,
		registerFloatTable: ft,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		c.symbolTable = node.SymbolTable
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
		case types.Float:
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
		case types.Int:
			printType = 1
		case types.Float:
			printType = 3
		case types.String:
			printType = 4
		default:
			return fmt.Errorf("compile error: can not print type: %q", node.Value.Type().Kind)
		}

		reg := node.Value.Register()

		var code []string
		if printType == 3 {
			code = []string{
				fmt.Sprintf("fmv.d fa0, %s", c.registerFloatTable.name(reg)),
				fmt.Sprintf("li a7, %d", printType),
				"ecall",
			}
		} else {
			code = []string{
				fmt.Sprintf("mv a0, %s", c.registerTable.name(reg)),
				fmt.Sprintf("li a7, %d", printType),
				"ecall",
			}
		}

		c.emit(code...)

		c.registerTable.dealloc(node.Value.Register())
	case *ast.VarStatement:
		var val string
		if node.Value == nil {
			switch node.Name.T.Kind {
			case types.Int:
				val = "0"
			}
		} else {
			val = node.Value.TokenLiteral()
		}

		// The label name is the name of the identifier.
		data := []string{
			fmt.Sprintf("%s: .dword %s", node.Name.Value, val),
		}
		c.addConstant(data...)
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("compiler error: undefined variable %s", node.Value)
		}

		reg, err := c.registerTable.alloc()
		if err != nil {
			return err
		}

		node.Reg = reg

		regName := c.registerTable.name(reg)

		code := []string{
			fmt.Sprintf("la %s, %s", regName, symbol.Name),
			fmt.Sprintf("ld %s, 0(%s)", regName, regName),
		}

		c.emit(code...)
	case *ast.InfixExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}

		if err := c.Compile(node.Right); err != nil {
			return err
		}

		var operator string
		switch node.Operator {
		case "+":
			operator = "add"
		case "-":
			operator = "sub"
		case "*":
			operator = "mul"
		case "/":
			operator = "div"
		default:
			return fmt.Errorf("unknown operator: %s", node.Operator)
		}

		var code []string

		switch node.T.Kind {
		case types.Float:
			// float point operations starts with f and end with .d for double
			// precision.
			operator = fmt.Sprintf("f%s.d", operator)
			code = []string{
				fmt.Sprintf(
					"%s %s, %s, %s",
					operator,
					c.registerFloatTable.name(node.Left.Register()),
					c.registerFloatTable.name(node.Left.Register()),
					c.registerFloatTable.name(node.Right.Register()),
				),
			}
		default:
			code = []string{
				fmt.Sprintf(
					"%s %s, %s, %s",
					operator,
					c.registerTable.name(node.Left.Register()),
					c.registerTable.name(node.Left.Register()),
					c.registerTable.name(node.Right.Register()),
				),
			}
		}

		c.emit(code...)
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

		data := []string{
			fmt.Sprintf("%s: .double %g", c.label.Name(), node.Value),
		}

		code := []string{
			fmt.Sprintf(
				"fld %s, %s, %s",
				c.registerFloatTable.name(node.Reg),
				c.label.Name(),
				c.registerTable.name(reg),
			),
		}

		c.addConstant(data...)
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

		data := []string{
			fmt.Sprintf("%s: .string %q", c.label.Name(), node.Value),
		}

		code := []string{
			fmt.Sprintf("la %s, %s", c.registerTable.name(node.Reg), c.label.Name()),
		}

		c.addConstant(data...)
		c.emit(code...)
	default:
		return fmt.Errorf("unknown type: %#v", node)
	}

	return nil
}

func (c *Compiler) emit(code ...string) {
	c.code = append(c.code, code...)
}

func (c *Compiler) addConstant(data ...string) {
	c.constants = append(c.constants, data...)
}

func (c *Compiler) Asm() string {
	var sb strings.Builder

	sb.WriteString(".data")
	sb.WriteString("\n")
	for _, d := range c.constants {
		sb.WriteString(d)
		sb.WriteString("\n")
	}

	sb.WriteString(".text")
	for _, t := range c.code {
		sb.WriteString("\n")
		sb.WriteString(t)
	}

	return sb.String()
}
