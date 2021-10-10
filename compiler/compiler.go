package compiler

import (
	"fmt"
	"math"
	"strings"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/symbol"
	"github.com/Glorforidor/didactic_compiler/types"
)

type Compiler struct {
	constants []string
	code      []string

	symbolTable   *symbol.Table
	registerTable *registerTable
	label         label
}

func New() *Compiler {
	return &Compiler{
		registerTable: riscvTable(),
	}
}

func stackSpace(defs int) int {
	x := math.Round(float64(defs) / 2)
	y := x * 16
	return int(y)
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
	case *ast.BlockStatement:
		c.symbolTable = node.SymbolTable // enter scope
		s := stackSpace(c.symbolTable.NumDefinitions)

		// Begin of block
		c.emit(fmt.Sprintf("addi sp, sp, -%d", s))
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
		// end of block
		c.emit(fmt.Sprintf("addi sp, sp, %d", s))

		c.symbolTable = node.SymbolTable.Outer // leave scope
	case *ast.ExpressionStatement:
		if err := c.Compile(node.Expression); err != nil {
			return err
		}

		reg := node.Expression.Register()
		c.registerTable.dealloc(reg)
	case *ast.PrintStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		c.loadGlobalValue(node.Value)

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
				fmt.Sprintf("fmv.d fa0, %s", reg),
				fmt.Sprintf("li a7, %d", printType),
				"ecall",
				// print newline after
				// fmt.Sprint("li a0, 0xa"),
				// fmt.Sprint("li, a7, 11"),
				// "ecall",
			}
		} else {
			code = []string{
				fmt.Sprintf("mv a0, %s", reg),
				fmt.Sprintf("li a7, %d", printType),
				"ecall",
				// print newline after
				// fmt.Sprint("li a0, 0xa"),
				// fmt.Sprint("li, a7, 11"),
				// "ecall",
			}
		}

		c.registerTable.dealloc(node.Value.Register())
		c.emit(code...)
	case *ast.VarStatement:
		sym, _ := c.symbolTable.Resolve(node.Name.Value)

		if sym.Scope == symbol.GlobalScope {
			la, err := createASMLabelIdentifier(sym.Name, sym.Type)
			if err != nil {
				return err
			}
			// The label name is the name of the identifier.
			c.addConstant(la)
		}

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
		c.loadGlobalValue(node.Value)

		s, _ := c.symbolTable.Resolve(node.Name.Value)

		if s.Scope == symbol.GlobalScope {
			switch node.Name.T.Kind {
			case types.Float:
				c.emit(fmt.Sprintf("fsd %s, 0(%v)", node.Value.Register(), node.Name.Reg))
			default:
				c.emit(fmt.Sprintf("sd %s, 0(%v)", node.Value.Register(), node.Name.Reg))
			}
		} else {
			switch node.Name.T.Kind {
			case types.Float:
				c.emit(fmt.Sprintf("fsd %s, %v", node.Value.Register(), s.Code()))
			default:
				c.emit(fmt.Sprintf("sd %s, %v", node.Value.Register(), s.Code()))
			}
		}

		c.registerTable.dealloc(node.Value.Register())
		c.registerTable.dealloc(node.Name.Reg)
	case *ast.Identifier:
		s, _ := c.symbolTable.Resolve(node.Value)

		reg, err := c.loadSymbol(s)
		if err != nil {
			return err
		}

		node.Reg = reg
	case *ast.InfixExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		c.loadGlobalValue(node.Left)

		if err := c.Compile(node.Right); err != nil {
			return err
		}
		c.loadGlobalValue(node.Right)

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
					left,
					left,
					right,
				),
			)
		default:
			c.emit(
				fmt.Sprintf(
					"%s %s, %s, %s",
					operator,
					left,
					left,
					right,
				),
			)
		}

		c.registerTable.dealloc(right)
		node.Reg = node.Left.Register()
	case *ast.IntegerLiteral:
		reg, err := c.registerTable.allocGeneral()
		if err != nil {
			return err
		}

		node.Reg = reg

		c.emit(
			fmt.Sprintf("li %s, %d", node.Reg, node.Value),
		)
	case *ast.FloatLiteral:
		reg, err := c.registerTable.allocFloating()
		if err != nil {
			return err
		}

		node.Reg = reg

		// register a temporay register for the fld instruction.
		reg, err = c.registerTable.allocGeneral()
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
				node.Reg,
				c.label.Name(),
				reg,
			),
		)

		// dealloc the temporay register
		c.registerTable.dealloc(reg)
	case *ast.StringLiteral:
		reg, err := c.registerTable.allocGeneral()
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
			fmt.Sprintf("la %s, %s", node.Reg, c.label.Name()),
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

func (c *Compiler) loadSymbol(s symbol.Symbol) (string, error) {
	switch {
	case s.Scope == symbol.GlobalScope:
		reg, err := c.registerTable.allocGeneral()
		if err != nil {
			return "", err
		}
		c.emit(fmt.Sprintf("la %s, %s", reg, s.Code()))

		return reg, nil
	default:
		switch s.Type.Kind {
		case types.Int, types.String:
			reg, err := c.registerTable.allocGeneral()
			if err != nil {
				return "", err
			}
			c.emit(fmt.Sprintf("ld %s, %s", reg, s.Code()))

			return reg, nil
		case types.Float:
			reg, err := c.registerTable.allocFloating()
			if err != nil {
				return "", err
			}

			c.emit(fmt.Sprintf("fld %s, %s", reg, s.Code()))

			return reg, nil
		}
	}

	return "", fmt.Errorf("compile error: can not load symbol: %s of type: %s", s.Name, s.Type.Kind)
}

// loadGlobalValue emits the load instructions iff the given node is of type
// ast.Identifier.
func (c *Compiler) loadGlobalValue(node ast.Expression) {
	id, ok := node.(*ast.Identifier)
	if !ok {
		return
	}

	s, _ := c.symbolTable.Resolve(id.Value)
	if s.Scope == symbol.GlobalScope {
		switch s.Type.Kind {
		case types.Float:
			// The identifier will always have normal register allocated to each
			// since we fetch them by address, so when identifier is a float, we
			// need to allocate a float register for it.
			reg, err := c.registerTable.allocFloating()
			if err != nil {
				panic(err)
			}
			c.emit(fmt.Sprintf("fld %s, 0(%s)", reg, id.Reg))

			// Deallocate the old normal register which held the address of the
			// label.
			c.registerTable.dealloc(id.Reg)

			id.Reg = reg
		default:
			c.emit(fmt.Sprintf("ld %s, 0(%s)", id.Reg, id.Reg))
		}
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
