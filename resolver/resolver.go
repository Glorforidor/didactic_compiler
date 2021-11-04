package resolver

import (
	"fmt"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/symbol"
)

func Resolve(node ast.Node, symbolTable *symbol.Table) error {
	switch node := node.(type) {
	case *ast.Program:
		node.SymbolTable = symbolTable

		for _, s := range node.Statements {
			if err := Resolve(s, node.SymbolTable); err != nil {
				return err
			}
		}
	case *ast.BlockStatement:
		node.SymbolTable = symbol.NewEnclosedTable(symbolTable)

		for _, s := range node.Statements {
			if err := Resolve(s, node.SymbolTable); err != nil {
				return err
			}
		}
	case *ast.PrintStatement:
		if err := Resolve(node.Value, symbolTable); err != nil {
			return err
		}
	case *ast.ExpressionStatement:
		if err := Resolve(node.Expression, symbolTable); err != nil {
			return err
		}
	case *ast.VarStatement:
		if err := Resolve(node.Value, symbolTable); err != nil {
			return err
		}
		if _, err := symbolTable.Define(node.Name.Value, node.Name.Ttoken.Type); err != nil {
			return err
		}
	case *ast.TypeStatement:
		if _, err := symbolTable.Define(node.Name.Value, node.Type); err != nil {
			return err
		}
	case *ast.AssignStatement:
		if err := Resolve(node.Name, symbolTable); err != nil {
			return err
		}
	case *ast.IfStatement:
		if err := Resolve(node.Condition, symbolTable); err != nil {
			return err
		}
		if err := Resolve(node.Consequence, symbolTable); err != nil {
			return err
		}

		if node.Alternative != nil {
			if err := Resolve(node.Alternative, symbolTable); err != nil {
				return err
			}
		}
	case *ast.ForStatement:
		node.SymbolTable = symbol.NewEnclosedTable(symbolTable)
		if err := Resolve(node.Init, node.SymbolTable); err != nil {
			return err
		}

		if err := Resolve(node.Condition, node.SymbolTable); err != nil {
			return err
		}

		if err := Resolve(node.Next, node.SymbolTable); err != nil {
			return err
		}

		if err := Resolve(node.Body, node.SymbolTable); err != nil {
			return err
		}
	case *ast.FuncStatement:
		if _, err := symbolTable.Define(node.Name.Value, node.Name.Ttoken.Type); err != nil {
			return err
		}

		node.SymbolTable = symbol.NewEnclosedTable(symbolTable)

		if node.Parameter != nil {
			node.SymbolTable.Define(node.Parameter.Value, node.Parameter.Ttoken.Type)
		}

		if err := Resolve(node.Body, node.SymbolTable); err != nil {
			return err
		}
	case *ast.ReturnStatement:
		if err := Resolve(node.Value, symbolTable); err != nil {
			return err
		}
	case *ast.Identifier:
		_, ok := symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("resolver: identifier: %q is not defined", node.Value)
		}
	case *ast.InfixExpression:
		if err := Resolve(node.Left, symbolTable); err != nil {
			return err
		}
		if err := Resolve(node.Right, symbolTable); err != nil {
			return err
		}
	}

	return nil
}
