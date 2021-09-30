package resolver

import (
	"fmt"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/symbol"
)

func Resolve(node ast.Node, symbolTable *symbol.Table) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := Resolve(s, symbolTable); err != nil {
				return err
			}
		}

		// When reaching here the symbol table should be the outer most, which
		// is the table with global definitions.
		node.SymbolTable = symbolTable
	case *ast.ExpressionStatement:
		if err := Resolve(node.Expression, symbolTable); err != nil {
			return err
		}
	case *ast.VarStatement:
		if _, err := symbolTable.Define(node.Name.Value, node.Name.T); err != nil {
			return err
		}
	case *ast.Identifier:
		_, ok := symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("resolver: identifier: %q is not defined", node.Value)
		}
	}

	return nil
}
