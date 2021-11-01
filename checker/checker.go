package checker

import (
	"fmt"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/symbol"
	"github.com/Glorforidor/didactic_compiler/token"
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
				node.Name.T,
				node.Value.Type(),
			)
		}
	case *ast.TypeStatement:
		// type human struct{name string}
		//

		if err := check(node.Name, symbolTable); err != nil {
			return err
		}

		return nil
	case *ast.StructType:
		for _, f := range node.Fields {
			if f.Ttoken.Type == token.Ident {
				if err := check(f, symbolTable); err != nil {
					return err
				}
			}

			f.T = tokenToType(f.Ttoken)
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
				node.Name.T,
				node.Value.Type(),
			)
		}
	case *ast.IfStatement:
		if err := check(node.Condition, symbolTable); err != nil {
			return err
		}

		if node.Condition.Type() != types.Typ[types.Bool] {
			return fmt.Errorf(
				"type error: non-bool %s (type %s) used as if condition",
				node.Condition.String(),
				node.Condition.Type(),
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

		if node.Condition.Type() != types.Typ[types.Bool] {
			return fmt.Errorf(
				"type error: non-bool %s (type %s) used as for condition",
				node.Condition.String(),
				node.Condition.Type(),
			)
		}

		if err := check(node.Next, node.SymbolTable); err != nil {
			return err
		}

		if err := check(node.Body, node.SymbolTable); err != nil {
			return err
		}
	case *ast.FuncStatement:
		if err := check(node.Name, symbolTable); err != nil {
			return err
		}

		if err := check(node.Parameter, node.SymbolTable); err != nil {
			return err
		}

		var result types.Type
		if node.Result.Literal != "" {
			result = tokenToType(node.Result)
		}

		// TODO: check result type is the same as return statements in body.
		// This will mean we need to loop over the body to check each statement
		// for a return statement and check if their type corrospond to the
		// Result type. Also the last statement in the body has to be a return
		// statement otherwise there would be unreachable code.

		if err := check(node.Body, node.SymbolTable); err != nil {
			return err
		}
		node.Name.T = &types.Signature{
			Parameter: node.Parameter.T,
			Result:    result,
		}
	case *ast.Identifier:
		sym, _ := symbolTable.Resolve(node.Value)
		switch sym.Type {
		case token.IntType:
			node.T = types.Typ[types.Int]
		case token.FloatType:
			node.T = types.Typ[types.Float]
		case token.StringType:
			node.T = types.Typ[types.String]
		case token.BoolType:
			node.T = types.Typ[types.Bool]
		case token.Ident:
			s, _ := symbolTable.Resolve(node.Ttoken.Literal)
			switch s.Type {
			case token.Func, token.Struct:
			default:
				panic(fmt.Sprintf("The underlying type of Ident: %s, is %s", node.Value, s.Type))
			}
			// if the identifier has another identifier as type, then that
			// identifier must be a struct or func type.
		case token.Struct:
			// if the identifier is a struct then the identifiers has been gone
			// through a TypeStatement which means the Identifier should have
			// the the correct type attached. Therefore, only check that the
			// node.T is not nil
			if node.T == nil {
				return fmt.Errorf("type error: identifier: %q, was not correctly typed as struct", node.Value)
			}
		case token.Func:
			// if the identifier is a func then it should have gone through
			// FuncStatement.
		default:
			if v, ok := sym.Type.(*ast.StructType); ok {
				if err := check(v, symbolTable); err != nil {
					return err
				}
				var fields []*types.Field
				for _, f := range v.Fields {
					fields = append(fields, &types.Field{
						Name: f.Value,
						Type: f.T,
					})
				}

				node.T = &types.Struct{fields}
			} else {
				return fmt.Errorf("type error: identifier: %q has the unknown type: %q", node.Value, node.Ttoken)
			}
		}
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
			return fmt.Errorf("type error: mismatch of types %s and %s", lt, rt)
		}

		if lt == types.Typ[types.String] {
			return fmt.Errorf("type error: operator: %v does not support type: %v", node.Operator, lt)
		}

		switch node.Operator {
		case "<":
			if lt == types.Typ[types.Bool] {
				return fmt.Errorf("type error: operator: %v does not support type: %v", node.Operator, lt)
			}

			node.T = types.Typ[types.Bool]
		case "==", "!=":
			node.T = types.Typ[types.Bool]
		default:
			node.T = lt
		}
	case *ast.IntegerLiteral:
		node.T = types.Typ[types.Int]
	case *ast.FloatLiteral:
		node.T = types.Typ[types.Float]
	case *ast.StringLiteral:
		node.T = types.Typ[types.String]
	case *ast.BoolLiteral:
		node.T = types.Typ[types.Bool]
	default:
		return fmt.Errorf("checker: ast node not handled: %T", node)
	}

	return nil
}

func tokenToType(t token.Token) types.Type {
	var typ types.Type
	switch t.Type {
	case token.IntType:
		typ = types.Typ[types.Int]
	case token.FloatType:
		typ = types.Typ[types.Float]
	case token.StringType:
		typ = types.Typ[types.String]
	case token.BoolType:
		typ = types.Typ[types.Bool]
	default:
		panic("Handle this at some point")
	}

	return typ
}
