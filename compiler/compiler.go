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

const wordAllignment = 8

// NOTE: the reason code and fun are splittet is to generate all function
// instructions in the buttom of the outputed assembly file. This will ensure
// that function is not outputted before a function is called since a assembly
// is computed from top to buttom.

type Compiler struct {
	// constants contains the constant data that will reside in the data
	// segment.
	constants []string
	// code contains the assembly code for non functions that will reside in
	// the text segment.
	code []string
	// fun contains the assembly code for functions that will reside in the
	// text segment.
	fun []string

	symbolTable   *symbol.Table
	registerTable *registerTable
	label         label

	// funcCleanUpStack signalise that functions epilogue is going to clean up
	// the stack space for all blocks within in the function.
	funcCleanUpStack bool

	// stackSpace grow each time a block allocated stack space and shrink when
	// block deallocates.
	stackSpace int

	// isTest is used for testing purposes. This will skip the wrapping of the
	// program in __start and __end.
	isTest bool
}

func New() *Compiler {
	return &Compiler{
		registerTable: riscvTable(),
	}
}

// newTest is only used for testing.
func newTest() *Compiler {
	return &Compiler{
		isTest:        true,
		registerTable: riscvTable(),
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		c.symbolTable = node.SymbolTable
		c.symbolTable.ComputeStack()
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
	case *ast.BlockStatement:
		c.symbolTable = node.SymbolTable
		c.symbolTable.ComputeStack()
		s := c.symbolTable.StackSpace()
		c.stackSpace += s
		c.emitf("addi sp, sp, -%d", s)
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}

		// let the *ast.FuncStatement clean up the stack.
		if !c.funcCleanUpStack {
			c.stackSpace -= s
			c.emitf("addi sp, sp, %d", s)
		}

		c.symbolTable = node.SymbolTable.Outer
	case *ast.ExpressionStatement:
		if err := c.Compile(node.Expression); err != nil {
			return err
		}

		c.registerTable.dealloc(node.Expression.Register())
	case *ast.PrintStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		c.loadGlobalOrPtrValue(node.Value)

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

		c.print(printType, reg)

		c.registerTable.dealloc(node.Value.Register())
	case *ast.VarStatement:
		s, _ := c.symbolTable.Resolve(node.Name.Value)

		if s.Scope == symbol.GlobalScope {
			err := c.createASMLabelIdentifier(s.Name, node.Name.T)
			if err != nil {
				return err
			}
		} else {
			if v, ok := node.Name.Type().(*types.Struct); ok {
				c.heapAllocate(v.Size())
				c.emitf("sd a0, %d(sp)", s.Code().(int))
			}
		}

		if node.Value == nil {
			break
		}

		// dirty hack!
		// This is like translating:
		// var x int = 2
		// -->
		// var x int
		// x = 2
		if err := c.Compile(
			&ast.AssignStatement{
				Name:  node.Name,
				Value: node.Value,
			},
		); err != nil {
			return err
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
		c.loadGlobalOrPtrValue(node.Value)

		var s *symbol.Symbol
		var offset int
		var isStruct bool
		switch t := node.Name.(type) {
		case *ast.Identifier:
			s, _ = c.symbolTable.Resolve(t.Value)
		case *ast.SelectorExpression:
			switch tt := t.X.(type) {
			case *ast.Identifier:
				s, _ = c.symbolTable.Resolve(tt.Value)
				offset = t.Offset
				isStruct = true
			}
		default:
			// We should never end here, so panic if we do.
			panic("unhandled type in assign statement")
		}

		regVal := node.Value.Register()
		regName := node.Name.Register()

		if s.Scope == symbol.GlobalScope {
			switch node.Name.Type().Kind() {
			case types.Float:
				c.emitf("fsd %s, %d(%v)", regVal, offset, regName)
			default:
				c.emitf("sd %s, %d(%v)", regVal, offset, regName)
			}
		} else {
			switch node.Name.Type().Kind() {
			case types.Float:
				if isStruct {
					c.emitf("fsd %s, %d(%v)", regVal, offset, regName)
				} else {
					c.emitf("fsd %s, %d(sp)", regVal, offset+s.Code().(int))
				}
			default:
				if isStruct {
					c.emitf("sd %s, %d(%v)", regVal, offset, regName)
				} else {
					c.emitf("sd %s, %d(sp)", regVal, offset+s.Code().(int))
				}
			}
		}

		c.registerTable.dealloc(regVal)
		c.registerTable.dealloc(regName)
	case *ast.IfStatement:
		if err := c.Compile(node.Condition); err != nil {
			return err
		}

		falseLabel := c.label.create()
		doneLabel := c.label.create()
		condRes := node.Condition.Register()

		c.emitf("beqz %s, %s", condRes, falseLabel)

		// Deallocate the condition register as it is not needed anymore.
		c.registerTable.dealloc(condRes)

		if err := c.Compile(node.Consequence); err != nil {
			return err
		}

		c.emitf("b %s", doneLabel)
		c.emitf("%s:", falseLabel)

		if node.Alternative != nil {
			if err := c.Compile(node.Alternative); err != nil {
				return err
			}
		}

		c.emitf("%s:", doneLabel)
	case *ast.ForStatement:
		c.symbolTable = node.SymbolTable
		c.symbolTable.ComputeStack()

		s := c.symbolTable.StackSpace()

		c.stackSpace += s

		c.emitf("addi sp, sp, -%d", s)

		if err := c.Compile(node.Init); err != nil {
			return err
		}

		topLabel := c.label.create()
		c.emitf("%s:", topLabel)

		doneLabel := c.label.create()

		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		condRes := node.Condition.Register()

		c.emitf("beqz %s, %s", condRes, doneLabel)

		c.registerTable.dealloc(condRes)

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		if err := c.Compile(node.Next); err != nil {
			return err
		}

		c.emitf("b %s", topLabel)
		c.emitf("%s:", doneLabel)

		if !c.funcCleanUpStack {
			c.stackSpace -= s
			c.emitf("addi sp, sp, %d", s)
		}
		c.symbolTable = node.SymbolTable.Outer
	case *ast.TypeStatement:
		// type statements is only needed for the semantic analysis.
	case *ast.FuncStatement:
		c.funcCleanUpStack = true
		defer func() { c.funcCleanUpStack = false }()
		// If there is no body then the node is forward declaration of a
		// function. Therefore, no need to generate code.
		if node.Body == nil {
			break
		}

		c.symbolTable = node.SymbolTable
		c.symbolTable.ComputeStack()

		space := c.symbolTable.StackSpace()
		c.emitf("%s:", node.Name.Value)

		if node.Signature.Parameter != nil {
			c.emitf("addi sp, sp, -%d", space)
			switch node.Signature.Parameter.Type().Kind() {
			case types.Float:
				c.emitf("fsd fa0, %d(sp)", wordAllignment)
			case types.StructKind:
				c.emitf("sd a0, %d(sp)", wordAllignment)
			default:
				c.emitf("sd a0, %d(sp)", wordAllignment)
			}
		} else {
			space = 16
			c.emitf("addi sp, sp, -%d", space)
		}

		c.emitf("sd ra, %d(sp)", space)

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// Epilogue of the function
		c.emitf("%s.epilogue:", node.Name.Value)

		c.emitf("addi sp, sp, %d", c.stackSpace)

		c.emitf("ld ra, %d(sp)", space)
		c.emitf("addi sp, sp, %d", space)
		c.emitf("ret")

		c.stackSpace = 0
		c.symbolTable = c.symbolTable.Outer
	case *ast.ReturnStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		c.loadGlobalOrPtrValue(node.Value)

		switch node.Value.Type().Kind() {
		case types.Float:
			c.emitf("fmv.d fa0, %s", node.Value.Register())
		default:
			c.emitf("mv a0, %s", node.Value.Register())
		}

		// Unconditionally jump to the functions epilogue.
		c.emitf("j %s.epilogue", node.Function.Value)

		c.registerTable.dealloc(node.Value.Register())
	case *ast.CallExpression:
		if node.T.Kind() != types.Nil {
			switch node.T.Kind() {
			case types.Float:
				node.Reg = "fa0"
			default:
				node.Reg = "a0"
			}
		}

		if node.Argument == nil {
			c.emitf("call %s", node.Function.TokenLiteral())
			return nil
		}

		if err := c.Compile(node.Argument); err != nil {
			return err
		}

		// We need to load the value from a global variable otherwise we would
		// pass along the address of the variable in the data segment and not
		// the value it points to.
		c.loadGlobalOrPtrValue(node.Argument)

		switch node.Argument.Type().Kind() {
		case types.Float:
			c.emitf("fmv.d fa0, %s", node.Argument.Register())
		default:
			c.emitf("mv a0, %s", node.Argument.Register())
		}

		c.emitf("call %s", node.Function.TokenLiteral())
	case *ast.Identifier:
		reg, err := c.loadSymbol(node)
		if err != nil {
			return err
		}

		node.Reg = reg
	case *ast.SelectorExpression:
		switch v := node.X.(type) {
		case *ast.Identifier:
			s, _ := c.symbolTable.Resolve(v.Value)
			if s.Scope == symbol.GlobalScope {
				if err := c.Compile(v); err != nil {
					return err
				}

				c.emitf("ld %s, 0(%s)", v.Reg, v.Reg)
				node.Reg = v.Reg
			} else {
				if err := c.Compile(v); err != nil {
					return err
				}
				node.Reg = v.Reg
			}
		case *ast.CallExpression:
			if err := c.Compile(v); err != nil {
				return err
			}

			reg, err := c.registerTable.allocGeneral()
			if err != nil {
				return err
			}
			c.emitf("ld %s, %d(a0)", reg, node.Offset)
			node.Reg = reg
		default:
			return fmt.Errorf("compiler error: SelectorExpression.X of type: %T is not implemented", v)
		}
	case *ast.InfixExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		c.loadGlobalOrPtrValue(node.Left)

		if err := c.Compile(node.Right); err != nil {
			return err
		}
		c.loadGlobalOrPtrValue(node.Right)

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

		c.emitf("li %s, %d", node.Reg, node.Value)
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

		floatLabel := c.label.create()

		la, err := c.createASMLabelLiteral(floatLabel, node.T, node.Value)
		if err != nil {
			return err
		}
		c.addConstant(la)

		c.emitf("fld %s, %s, %s", node.Reg, floatLabel, reg)

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

		c.emitf("la %s, %s", node.Reg, stringLabel)
	case *ast.BoolLiteral:
		reg, err := c.registerTable.allocGeneral()
		if err != nil {
			return err
		}

		node.Reg = reg
		if node.Value {
			c.emitf("li %s, 1", reg)
		} else {
			c.emitf("li %s, 0", reg)
		}
	default:
		return fmt.Errorf("unknown type: %#v", node)
	}

	return nil
}

func (c *Compiler) emitf(format string, a ...interface{}) {
	if c.funcCleanUpStack {
		c.fun = append(c.fun, fmt.Sprintf(format, a...))
	} else {
		c.code = append(c.code, fmt.Sprintf(format, a...))
	}
}

func (c *Compiler) addConstantf(format string, a ...interface{}) {
	c.constants = append(c.constants, fmt.Sprintf(format, a...))
}
func (c *Compiler) addConstant(data ...string) {
	c.constants = append(c.constants, data...)
}

func (c *Compiler) loadSymbol(node *ast.Identifier) (string, error) {
	s, _ := c.symbolTable.Resolve(node.Value)

	switch s.Scope {
	case symbol.GlobalScope:
		reg, err := c.registerTable.allocGeneralSaved()
		if err != nil {
			return "", err
		}

		c.emitf("la %s, %s", reg, s.Code())

		return reg, nil
	case symbol.LocalScope:
		reg, err := c.allocateRegByType(node.T)
		if err != nil {
			return "", err
		}

		ld, err := loadASM(node.T)
		if err != nil {
			return "", err
		}

		c.emitf(ld, reg, s.Code())

		return reg, nil
	default:
		return "", fmt.Errorf("compile error: can not load symbol: %s of type: %s", s.Name, s.Type)
	}
}

func loadASM(t types.Type) (string, error) {
	switch t.Kind() {
	case types.Bool, types.Int, types.String:
		return "ld %s, %d(sp)", nil
	case types.Float:
		return "fld %s, %d(sp)", nil
	case types.StructKind:
		return "ld %s, %d(sp)", nil
	default:
		return "", fmt.Errorf("compile error: loading value of type: %s is not supported", t)
	}
}

// loadSelectorValue emits the load instruction iff the SelectorExpression
// selects from a global identifier. Otherwise emits nothing.
func (c *Compiler) loadSelectorValue(sel *ast.SelectorExpression) {
	_, ok := sel.X.(*ast.Identifier)
	if !ok {
		return
	}

	offset := sel.Offset

	switch sel.T.Kind() {
	case types.Float:
		// The identifier will always have normal register allocated to each
		// since we fetch them by address, so when identifier is a float, we
		// need to allocate a float register for it.
		reg, err := c.registerTable.allocFloating()
		if err != nil {
			// This should probably not happen as there are many floating
			// registers to allocate, but if there is a error then panic.
			panic(err)
		}

		c.emitf("fld %s, %d(%s)", reg, offset, sel.Reg)

		// Deallocate the old normal register which held the address of the
		// label.
		c.registerTable.dealloc(sel.Reg)

		// Update the SelectorExpressions register.
		sel.Reg = reg
	default:
		c.emitf("ld %s, %d(%s)", sel.Register(), offset, sel.Register())
	}
}

// loadIdentifier emits the load instruction iff the identifier is global.
// Otherwise emits nothing.
func (c *Compiler) loadIdentifier(id *ast.Identifier) {
	s, _ := c.symbolTable.Resolve(id.Value)

	if s.Scope != symbol.GlobalScope {
		return
	}

	switch id.T.Kind() {
	case types.Float:
		// The identifier will always have normal register allocated to each
		// since we fetch them by address, so when identifier is a float, we
		// need to allocate a float register for it.
		reg, err := c.registerTable.allocFloating()
		if err != nil {
			// This should probably not happen as there are many floating
			// registers to allocate, but if there is a error then panic.
			panic(err)
		}
		c.emitf("fld %s, 0(%s)", reg, id.Reg)

		// Deallocate the old normal register which held the address of the
		// label.
		c.registerTable.dealloc(id.Reg)

		id.Reg = reg
	default:
		c.emitf("ld %s, 0(%s)", id.Reg, id.Reg)
	}
}

// loadGlobalOrPtrValue emits the load instructions iff the given node is of type
// *ast.Identifier or *ast.SelectorExpression.
func (c *Compiler) loadGlobalOrPtrValue(node ast.Expression) {
	switch node := node.(type) {
	case *ast.Identifier:
		c.loadIdentifier(node)
	case *ast.SelectorExpression:
		c.loadSelectorValue(node)
	default:
		// Ignore every other ast node as they do not have a global value to be
		// loaded.
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
		c.emitf("%s %s, %s, %s", operator, left, left, right)
	default:
		c.emitf("%s %s, %s, %s", operator, left, left, right)
	}
}

func (c *Compiler) compare(operator, left, right string, t types.Type) {
	trueLabel := c.label.create()
	doneLabel := c.label.create()
	switch t.Kind() {
	case types.Float:
		// TODO: Maybe panicing over floating comparison is a bit excessive.
		panic("compile error: floating point comparison is not implemented yet")
	default:
		c.emitf("%s %s, %s, %s", operator, left, right, trueLabel)
		c.emitf("li %s, %d", left, cFalse)
		c.emitf("b %s", doneLabel)
		c.emitf("%s:", trueLabel)
		c.emitf("li %s, %d", left, cTrue)
		c.emitf("%s:", doneLabel)
	}
}

func (c *Compiler) allocateRegByType(t types.Type) (string, error) {
	switch t.Kind() {
	case types.Float:
		return c.registerTable.allocFloating()
	default:
		return c.registerTable.allocGeneral()
	}
}

// createASMLabelIdentifier creates an asm label for identifiers and if the
// identifier is a struct, then it also emits the instructions to allocate
// space on the heap for it.
func (c *Compiler) createASMLabelIdentifier(name string, t types.Type) error {
	// TODO: This method could easily just emit the code instead of returning
	// the string.
	switch t.Kind() {
	case types.Int, types.String, types.Bool:
		// string identifiers are treated as memory address of the actual
		// string.
		c.addConstantf("%s: .dword 0", name)
	case types.Float:
		c.addConstantf("%s: .double 0", name)
	case types.StructKind:
		str := t.(*types.Struct)
		c.heapAllocate(str.Size())
		reg, err := c.registerTable.allocGeneral()
		if err != nil {
			return err
		}
		c.emitf("la %s, %s", reg, name)
		c.emitf("sd a0, 0(%s)", reg)

		c.addConstantf("%s: .dword 0", name)

		c.registerTable.dealloc(reg)
	default:
		return fmt.Errorf("compiler error: could not create label: %s with type: %s", name, t)
	}

	return nil
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

func (c *Compiler) print(printType int, reg string) {
	if printType == 3 {
		c.emitf("fmv.d fa0, %s", reg)
		c.emitf("li a7, %d", printType)
		c.emitf("ecall")
	} else {
		c.emitf("mv a0, %s", reg)
		c.emitf("li a7, %d", printType)
		c.emitf("ecall")
	}
}

func (c *Compiler) heapAllocate(size int) {
	c.emitf("li a0, %d", size)
	c.emitf("li a7, 9")
	c.emitf("ecall")
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
	if !c.isTest {
		sb.WriteString("\n")
		sb.WriteString("__start:")
	}
	for _, cc := range c.code {
		sb.WriteString("\n")
		sb.WriteString(cc)
	}
	if !c.isTest {
		sb.WriteString("\n")
		sb.WriteString("j __end")
	}
	for _, f := range c.fun {
		sb.WriteString("\n")
		sb.WriteString(f)
	}

	if !c.isTest {
		sb.WriteString("\n")
		sb.WriteString("__end:\n")
		sb.WriteString("li a7, 10\n")
		sb.WriteString("ecall")
	}

	return sb.String()
}
