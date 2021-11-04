package compiler

import (
	"fmt"
	"strings"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/symbol"
	"github.com/Glorforidor/didactic_compiler/types"
)

const (
	cTrue  = 1
	cFalse = 0
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

		s := c.symbolTable.StackSpace()

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
		switch node.Value.Type().Kind() {
		case types.Int, types.Bool:
			printType = 1
		case types.Float:
			printType = 3
		case types.String:
			printType = 4
		default:
			return fmt.Errorf("compile error: can not print type: %q", node.Value.Type())
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
		s, _ := c.symbolTable.Resolve(node.Name.Value)

		if s.Scope == symbol.GlobalScope {
			la, err := c.createASMLabelIdentifier(s.Name, node.Name.T)
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
		// BUG: There is a bug with var statement in blocks, if they are not
		// initialise with their zero value, the value from the previous
		// interation is kept, even though it is declared again. I need to
		// explicit set a zero value if there is non assigned.
	case *ast.AssignStatement:
		// TODO: maybe move this to into the global scope check?
		// Otherwise we will emit an unnecessary load instruction for locals.
		if err := c.Compile(node.Name); err != nil {
			return err
		}

		if err := c.Compile(node.Value); err != nil {
			return err
		}
		c.loadGlobalValue(node.Value)

		s, _ := c.symbolTable.Resolve(node.Name.Value)

		if s.Scope == symbol.GlobalScope {
			switch node.Name.T.Kind() {
			case types.Float:
				c.emit(fmt.Sprintf("fsd %s, 0(%v)", node.Value.Register(), node.Name.Reg))
			default:
				c.emit(fmt.Sprintf("sd %s, 0(%v)", node.Value.Register(), node.Name.Reg))
			}
		} else {
			switch node.Name.T.Kind() {
			case types.Float:
				c.emit(fmt.Sprintf("fsd %s, %v", node.Value.Register(), s.Code()))
			default:
				c.emit(fmt.Sprintf("sd %s, %v", node.Value.Register(), s.Code()))
			}
		}

		c.registerTable.dealloc(node.Value.Register())
		c.registerTable.dealloc(node.Name.Reg)
	case *ast.IfStatement:
		if err := c.Compile(node.Condition); err != nil {
			return err
		}

		falseLabel := c.label.create()
		doneLabel := c.label.create()
		condRes := node.Condition.Register()

		c.emit(fmt.Sprintf("beqz %s, %s", condRes, falseLabel))

		// Deallocate the condition register as it is not needed anymore.
		c.registerTable.dealloc(condRes)

		if err := c.Compile(node.Consequence); err != nil {
			return err
		}

		c.emit(fmt.Sprintf("b %s", doneLabel))
		c.emit(fmt.Sprintf("%s:", falseLabel))

		if node.Alternative != nil {
			if err := c.Compile(node.Alternative); err != nil {
				return err
			}
		}

		c.emit(fmt.Sprintf("%s:", doneLabel))
	case *ast.ForStatement:
		c.symbolTable = node.SymbolTable

		s := c.symbolTable.StackSpace()

		// Begin of block
		c.emit(fmt.Sprintf("addi sp, sp, -%d", s))

		if err := c.Compile(node.Init); err != nil {
			return err
		}

		topLabel := c.label.create()
		c.emit(fmt.Sprintf("%s:", topLabel))

		doneLabel := c.label.create()

		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		condRes := node.Condition.Register()

		c.emit(fmt.Sprintf("beqz %s, %s", condRes, doneLabel))

		c.registerTable.dealloc(condRes)

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		if err := c.Compile(node.Next); err != nil {
			return err
		}

		c.emit(fmt.Sprintf("b %s", topLabel))

		c.emit(fmt.Sprintf("%s:", doneLabel))

		// End of block
		c.emit(fmt.Sprintf("addi sp, sp, %d", s))
		c.symbolTable = node.SymbolTable.Outer
	case *ast.TypeStatement:
		// type statements is only needed for the semantic analysis.
	case *ast.FuncStatement:
		c.emit(fmt.Sprintf("%s:", node.Name.Value))
		// always allocate 16 on the stack since we need to store the return address
		c.emit("addi sp, sp, -16")

		if node.Parameter != nil {
			c.emit("sd a0, 8(sp)")
		}

		c.emit(fmt.Sprintf("sd ra, 16(sp)"))

		c.Compile(node.Body)

		c.emit(fmt.Sprintf("ld ra, 16(sp)"))
		c.emit("addi sp, sp, 16")
		c.emit("ret")
	case *ast.ReturnStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		c.emit(fmt.Sprintf("mv a0, %s", node.Value.Register()))
		c.registerTable.dealloc(node.Value.Register())
	case *ast.Identifier:
		reg, err := c.loadSymbol(node)
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

		if err := c.infix(node); err != nil {
			return err
		}

		c.registerTable.dealloc(node.Right.Register())
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

		floatLabel := c.label.create()

		la, err := c.createASMLabelLiteral(floatLabel, node.T, node.Value)
		if err != nil {
			return err
		}
		c.addConstant(la)

		c.emit(
			fmt.Sprintf(
				"fld %s, %s, %s",
				node.Reg,
				floatLabel,
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

		stringLabel := c.label.create()
		la, err := c.createASMLabelLiteral(stringLabel, node.T, node.Value)
		if err != nil {
			return err
		}

		c.addConstant(la)

		c.emit(
			fmt.Sprintf("la %s, %s", node.Reg, stringLabel),
		)
	case *ast.BoolLiteral:
		reg, err := c.registerTable.allocGeneral()
		if err != nil {
			return err
		}

		node.Reg = reg
		if node.Value {
			c.emit(fmt.Sprintf("li %s, 1", reg))
		} else {
			c.emit(fmt.Sprintf("li %s, 0", reg))
		}
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

func (c *Compiler) loadSymbol(node *ast.Identifier) (string, error) {
	s, _ := c.symbolTable.Resolve(node.Value)

	switch {
	case s.Scope == symbol.GlobalScope:
		reg, err := c.registerTable.allocGeneral()
		if err != nil {
			return "", err
		}
		c.emit(fmt.Sprintf("la %s, %s", reg, s.Code()))

		return reg, nil
	default:
		switch node.T.Kind() {
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

	return "", fmt.Errorf("compile error: can not load symbol: %s of type: %s", s.Name, s.Type)
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
		switch node.Type().Kind() {
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

func (c *Compiler) infix(inf *ast.InfixExpression) error {
	left := inf.Left.Register()
	right := inf.Right.Register()
	switch inf.Operator {
	case "+":
		c.arithmetic("add", left, right, inf.T)
	case "-":
		c.arithmetic("sub", left, right, inf.T)
	case "*":
		c.arithmetic("mul", left, right, inf.T)
	case "/":
		c.arithmetic("div", left, right, inf.T)
	case "<":
		c.compare("blt", left, right, inf.T)
	case "==":
		c.compare("beq", left, right, inf.T)
	case "!=":
		c.compare("bne", left, right, inf.T)
	default:
		return fmt.Errorf("unknown operator: %s", inf.Operator)
	}

	return nil
}

func (c *Compiler) arithmetic(operator, left, right string, t types.Type) {
	switch t.Kind() {
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
}

func (c *Compiler) compare(operator, left, right string, t types.Type) {
	trueLabel := c.label.create()
	doneLabel := c.label.create()
	switch t.Kind() {
	default:
		c.emit(
			fmt.Sprintf(
				"%s %s, %s, %s",
				operator,
				left,
				right,
				trueLabel,
			),
			fmt.Sprintf("li %s, %d", left, cFalse),
			fmt.Sprintf("b %s", doneLabel),
			fmt.Sprintf("%s:", trueLabel),
			fmt.Sprintf("li %s, %d", left, cTrue),
			fmt.Sprintf("%s:", doneLabel),
		)
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
func (c *Compiler) createASMLabelIdentifier(name string, t types.Type) (string, error) {
	switch t.Kind() {
	case types.Int, types.String, types.Bool:
		// string identifiers are treated as memory address of the actual
		// string.
		return fmt.Sprintf("%s: .dword 0", name), nil
	case types.Float:
		return fmt.Sprintf("%s: .double 0", name), nil
	case types.StructKind:
		str, _ := t.(*types.Struct)

		// A string field in a struct is stored as an address to a label with
		// the string value.
		var stringLits strings.Builder
		var structBuilder strings.Builder
		fmt.Fprintf(&structBuilder, "%s:\n", name)
		for _, f := range str.Fields {
			switch f.Type.Kind() {
			case types.Int, types.Bool:
				structBuilder.WriteString(".dword 0")
			case types.Float:
				structBuilder.WriteString(".double 0.0")
			case types.String:
				s := c.label.create()
				strLabel, err := c.createASMLabelLiteral(s, f.Type, "")
				if err != nil {
					return "", err
				}

				stringLits.WriteString(strLabel)
				fmt.Fprintf(&structBuilder, ".dword %s", s)
			}
		}

		return stringLits.String() + "\n" + structBuilder.String(), nil
	default:
		return "", fmt.Errorf("compiler error: could not create label: %s with type: %s", name, t)
	}
}

// createASMLabelLiteral creates an asm label for literal values.
func (c *Compiler) createASMLabelLiteral(name string, t types.Type, value interface{}) (string, error) {
	if value == nil {
		return "", fmt.Errorf("compiler error: can not define label with no value")
	}

	switch t.Kind() {
	case types.Float:
		return fmt.Sprintf("%s: .double %v", name, value), nil
	case types.String:
		return fmt.Sprintf(`%s: .string "%v"`, name, value), nil
	default:
		return "", fmt.Errorf("compiler error: could not create label: %s with type: %T", name, t)
	}
}
