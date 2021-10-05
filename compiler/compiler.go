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
		t := node.Expression.Type().Kind
		_, ok := node.Expression.(*ast.Identifier)
		switch {
		case t == types.Float && !ok:
			c.registerFloatTable.dealloc(reg)
		default:
			c.registerTable.dealloc(reg)
		}
	case *ast.PrintStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		c.loadValue(node.Value)

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
				// print newline after
				// fmt.Sprint("li a0, 0xa"),
				// fmt.Sprint("li, a7, 11"),
				// "ecall",
			}
			c.registerFloatTable.dealloc(node.Value.Register())
		} else {
			code = []string{
				fmt.Sprintf("mv a0, %s", c.registerTable.name(reg)),
				fmt.Sprintf("li a7, %d", printType),
				"ecall",
				// print newline after
				// fmt.Sprint("li a0, 0xa"),
				// fmt.Sprint("li, a7, 11"),
				// "ecall",
			}
			c.registerTable.dealloc(node.Value.Register())
		}

		c.emit(code...)
	case *ast.VarStatement:
		la, err := createASMLabelIdentifier(node.Name.Value, node.Name.T)
		if err != nil {
			return err
		}
		// The label name is the name of the identifier.
		c.addConstant(la)

		// dirty hack!
		// This is like translating:
		// var x int = 2
		// -->
		// var x int
		// x = 2
		if node.Value != nil {
			if err := c.Compile(
				&ast.AssignStatement{
					Name:  node.Name,
					Value: node.Value,
				},
			); err != nil {
				return err
			}
		}
	case *ast.AssignStatement:
		if err := c.Compile(node.Name); err != nil {
			return err
		}

		if err := c.Compile(node.Value); err != nil {
			return err
		}

		c.loadValue(node.Value)

		switch node.Name.T.Kind {
		case types.Float:
			c.emit(fmt.Sprintf("fsd %s, 0(%v)", c.registerFloatTable.name(node.Value.Register()), c.registerTable.name(node.Name.Reg)))
			c.registerFloatTable.dealloc(node.Value.Register())
		default:
			c.emit(fmt.Sprintf("sd %s, 0(%v)", c.registerTable.name(node.Value.Register()), c.registerTable.name(node.Name.Reg)))
			c.registerTable.dealloc(node.Value.Register())
		}

		c.registerTable.dealloc(node.Name.Reg)
	case *ast.Identifier:
		symbol, _ := c.symbolTable.Resolve(node.Value)

		reg, err := c.registerTable.alloc()
		if err != nil {
			return err
		}

		node.Reg = reg

		regName := c.registerTable.name(reg)

		c.emit(fmt.Sprintf("la %s, %s", regName, symbol.Code()))
	case *ast.InfixExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		c.loadValue(node.Left)

		if err := c.Compile(node.Right); err != nil {
			return err
		}
		c.loadValue(node.Right)

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

		left := node.Left.Register()
		right := node.Right.Register()

		switch node.T.Kind {
		case types.Float:
			// float point operations starts with f and end with .d for double
			// precision.
			operator = fmt.Sprintf("f%s.d", operator)
			c.emit(
				fmt.Sprintf(
					"%s %s, %s, %s",
					operator,
					c.registerFloatTable.name(left),
					c.registerFloatTable.name(left),
					c.registerFloatTable.name(right),
				),
			)
			c.registerFloatTable.dealloc(right)
		default:
			c.emit(
				fmt.Sprintf(
					"%s %s, %s, %s",
					operator,
					c.registerTable.name(left),
					c.registerTable.name(left),
					c.registerTable.name(right),
				),
			)
			c.registerTable.dealloc(right)
		}

		node.Reg = node.Left.Register()
	case *ast.IntegerLiteral:
		reg, err := c.registerTable.alloc()
		if err != nil {
			return err
		}

		node.Reg = reg

		c.emit(
			fmt.Sprintf("li %s, %d", c.registerTable.name(node.Reg), node.Value),
		)
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

		la, err := createASMLabelLiteral(c.label.Name(), node.T, node.Value)
		if err != nil {
			return err
		}
		c.addConstant(la)

		c.emit(
			fmt.Sprintf(
				"fld %s, %s, %s",
				c.registerFloatTable.name(node.Reg),
				c.label.Name(),
				c.registerTable.name(reg),
			),
		)

		// dealloc the temporay register
		c.registerTable.dealloc(reg)
	case *ast.StringLiteral:
		reg, err := c.registerTable.alloc()
		if err != nil {
			return err
		}

		node.Reg = reg
		c.label.Create()

		la, err := createASMLabelLiteral(c.label.Name(), node.T, node.Value)
		if err != nil {
			return err
		}

		c.addConstant(la)

		c.emit(
			fmt.Sprintf("la %s, %s", c.registerTable.name(node.Reg), c.label.Name()),
		)
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

// loadValue emits the load instructions iff the given node is of type
// ast.Identifier.
func (c *Compiler) loadValue(node ast.Expression) {
	id, ok := node.(*ast.Identifier)
	if !ok {
		return
	}

	switch id.T.Kind {
	case types.Float:
		// The identifier will always have normal register allocated to each
		// since we fetch them by address, so when identifier is a float, we
		// need to allocate a float register for it.
		freg, err := c.registerFloatTable.alloc()
		if err != nil {
			panic(err)
		}
		c.emit(fmt.Sprintf("fld %s, 0(%s)", c.registerFloatTable.name(freg), c.registerTable.name(id.Reg)))

		// Deallocate the old normal register which held the address of the
		// label.
		c.registerTable.dealloc(id.Reg)

		id.Reg = freg
	default:
		c.emit(fmt.Sprintf("ld %s, 0(%s)", c.registerTable.name(id.Reg), c.registerTable.name(id.Reg)))
	}
}

// Asm returns the compiled assembly code.
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

// createASMLabelIdentifier creates an asm label for identifiers.
func createASMLabelIdentifier(name string, t types.Type) (string, error) {
	switch t.Kind {
	case types.Int, types.String:
		// string identifiers are treated as memory address of the actual
		// string.
		return fmt.Sprintf("%s: .dword 0", name), nil
	case types.Float:
		return fmt.Sprintf("%s: .double 0", name), nil
	default:
		return "", fmt.Errorf("compiler error: could create label: %s with type: %T", name, t)
	}
}

// createASMLabelLiteral creates an asm label for literal values.
func createASMLabelLiteral(name string, t types.Type, value interface{}) (string, error) {
	if value == nil {
		return "", fmt.Errorf("compiler error: can not define label with no value")
	}

	switch t.Kind {
	case types.Int:
		return fmt.Sprintf("%s: .dword %v", name, value), nil
	case types.Float:
		return fmt.Sprintf("%s: .double %v", name, value), nil
	case types.String:
		return fmt.Sprintf(`%s: .string "%v"`, name, value), nil
	default:
		return "", fmt.Errorf("compiler error: could create label: %s with type: %T", name, t)
	}
}

// zeroValue returns the zero value of any type.
func zeroValue(t types.Type) interface{} {
	switch t.Kind {
	case types.Float:
		return 0.0
	case types.String:
		return ""
	default:
		return 0
	}
}
