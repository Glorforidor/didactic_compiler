package types

import (
	"fmt"

	"github.com/Glorforidor/didactic_compiler/ast"
)

func Checker(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := Checker(s); err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		if err := Checker(node.Expression); err != nil {
			return err
		}
	case *ast.PrintStatement:
		if err := Checker(node.Value); err != nil {
			return err
		}
	case *ast.InfixExpression:
		if err := Checker(node.Left); err != nil {
			return err
		}

		if err := Checker(node.Right); err != nil {
			return err
		}

		lt := node.Left.Type()
		rt := node.Right.Type()

		if lt.Kind != rt.Kind {
			return fmt.Errorf("type error: mismatch of types %s and %s", lt.Kind, rt.Kind)
		}

		node.T = lt
	case *ast.IntegerLiteral:
		node.T = ast.Type{Kind: ast.Int}
	case *ast.FloatLiteral:
		node.T = ast.Type{Kind: ast.Float}
	case *ast.StringLiteral:
		node.T = ast.Type{Kind: ast.String}
	}

	return nil
}
