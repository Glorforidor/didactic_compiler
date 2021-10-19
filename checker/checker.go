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
	case *ast.IfStatement:
		if err := check(node.Condition, symbolTable); err != nil {
			return err
		}

		if node.Condition.Type().Kind != types.Bool {
			return fmt.Errorf(
				"type error: non-bool %s (type %s) used as if condition",
				node.Condition.String(),
				node.Condition.Type().Kind,
			)
		}

		if err := check(node.Consequence, symbolTable); err != nil {
			return err
		}

		if node.Alternative != nil {
			if err := check(node.Alternative, symbolTable); err != nil {
				return err
			}
		}
	case *ast.ForStatement:
		if err := check(node.Init, node.SymbolTable); err != nil {
			return err
		}

		if err := check(node.Condition, node.SymbolTable); err != nil {
			return err
		}

		if node.Condition.Type().Kind != types.Bool {
			return fmt.Errorf(
				"type error: non-bool %s (type %s) used as for condition",
				node.Condition.String(),
				node.Condition.Type().Kind,
			)
		}

		if err := check(node.Next, node.SymbolTable); err != nil {
			return err
		}

		if err := check(node.Body, node.SymbolTable); err != nil {
			return err
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
			return fmt.Errorf("type error: operator: %v does not support type: %v", node.Operator, lt.Kind)
		}

		switch node.Operator {
		case "<":
			if lt.Kind == types.Bool {
				return fmt.Errorf("type error: operator: %v does not support type: %v", node.Operator, lt.Kind)
			}

			node.T = types.Type{Kind: types.Bool}
		case "==", "!=":
			node.T = types.Type{Kind: types.Bool}
		default:
			node.T = lt
		}
	case *ast.IntegerLiteral:
		node.T = types.Type{Kind: types.Int}
	case *ast.FloatLiteral:
		node.T = types.Type{Kind: types.Float}
	case *ast.StringLiteral:
		node.T = types.Type{Kind: types.String}
	case *ast.BoolLiteral:
		node.T = types.Type{Kind: types.Bool}
	}

	return nil
}
