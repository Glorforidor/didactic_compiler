package checker

import (
	"fmt"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/symbol"
	"github.com/Glorforidor/didactic_compiler/types"
)

func Check(program *ast.Program) error {
	return check(program, program.SymbolTable)
}

func check(node ast.Node, symbolTable *symbol.Table) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := check(s, node.SymbolTable); err != nil {
				return err
			}
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			if err := check(s, node.SymbolTable); err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		if err := check(node.Expression, symbolTable); err != nil {
			return err
		}
	case *ast.PrintStatement:
		if err := check(node.Value, symbolTable); err != nil {
			return err
		}
	case *ast.VarStatement:
		if err := check(node.Name, symbolTable); err != nil {
			return err
		}

		if node.Value == nil {
			break
		}

		if err := check(node.Value, symbolTable); err != nil {
			return err
		}

		if node.Name.T != node.Value.Type() {
			return fmt.Errorf(
				"type error: identifier: %q of type: %s is assigned the wrong type: %s",
				node.Name.Value,
				node.Name.T.Kind,
				node.Value.Type().Kind,
			)
		}
	case *ast.AssignStatement:
		if err := check(node.Name, symbolTable); err != nil {
			return err
		}

		if err := check(node.Value, symbolTable); err != nil {
			return err
		}

		if node.Name.T != node.Value.Type() {
			return fmt.Errorf(
				"type error: identifier: %q of type: %s is assigned the wrong type: %s",
				node.Name.Value,
				node.Name.T.Kind,
				node.Value.Type().Kind,
			)
		}
	case *ast.Identifier:
		sym, _ := symbolTable.Resolve(node.Value)
		if sym.Type.Kind == types.Unknown {
			return fmt.Errorf("type error: identifier: %s is unknown", node.Value)
		}

		node.T = sym.Type
	case *ast.InfixExpression:
		if err := check(node.Left, symbolTable); err != nil {
			return err
		}

		if err := check(node.Right, symbolTable); err != nil {
			return err
		}

		lt := node.Left.Type()
		rt := node.Right.Type()

		if lt != rt {
			return fmt.Errorf("type error: mismatch of types %s and %s", lt.Kind, rt.Kind)
		}

		if lt.Kind == types.String {
			return fmt.Errorf("type error: operator: %v does not support type: %v", node.Operator, lt)
		}

		node.T = lt
	case *ast.IntegerLiteral:
		node.T = types.Type{Kind: types.Int}
	case *ast.FloatLiteral:
		node.T = types.Type{Kind: types.Float}
	case *ast.StringLiteral:
		node.T = types.Type{Kind: types.String}
	}

	return nil
}
