package resolver

import (
	"fmt"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/symbol"
)

// funcPrototypes hold information about a function being a protype.
var funcPrototypes = map[string]bool{}

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
		if _, err := symbolTable.Define(node.Name.Value, node.Name.Tnode); err != nil {
			return err
		}
	case *ast.TypeStatement:
		if _, err := symbolTable.DefineType(node.Name.Value, node.Type); err != nil {
			return err
		}
	case *ast.AssignStatement:
		if err := Resolve(node.Name, symbolTable); err != nil {
			return err
		}
		if err := Resolve(node.Value, symbolTable); err != nil {
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
		_, ok := symbolTable.Resolve(node.Name.Value)
		if !ok {
			if _, err := symbolTable.DefineFunc(node.Name.Value, node.Signature); err != nil {
				return err
			}
		} else if !funcPrototypes[node.Name.Value] {
			return fmt.Errorf("resolver: function: %q already defined", node.Name.Value)
		} else if node.Body == nil && funcPrototypes[node.Name.Value] {
			return fmt.Errorf("resolver: function: %q already prototyped", node.Name.Value)
		}

		funcPrototypes[node.Name.Value] = node.Body == nil

		node.SymbolTable = symbol.NewEnclosedTable(symbolTable)

		if node.Signature.Parameter != nil {
			s, ok := node.SymbolTable.Resolve(node.Signature.Parameter.Value)
			if ok && s.Scope == symbol.TypeScope {
				node.SymbolTable.DefineFuncParameter(
					node.Signature.Parameter.Value,
					s.Type,
				)
			} else {
				// Allow the parameter to over shadow a global variable of same
				// name.
				node.SymbolTable.DefineFuncParameter(
					node.Signature.Parameter.Value,
					node.Signature.Parameter.Tnode,
				)
			}
		}

		if node.Body != nil {
			if err := Resolve(node.Body, node.SymbolTable); err != nil {
				return err
			}
		}
	case *ast.ReturnStatement:
		if err := Resolve(node.Value, symbolTable); err != nil {
			return err
		}
	case *ast.CallExpression:
		if err := Resolve(node.Function, symbolTable); err != nil {
			return err
		}

		if err := Resolve(node.Argument, symbolTable); err != nil {
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
	case *ast.SelectorExpression:
		if err := Resolve(node.X, symbolTable); err != nil {
			return err
		}
	}

	return nil
}
