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
		if err := check(node.Name, symbolTable); err != nil {
			return err
		}

		if err := check(node.Type, symbolTable); err != nil {
			return err
		}
	case *ast.StructType:
		for _, f := range node.Fields {
			// In struct only allow for basic types

			// TODO: maybe later allow for struct inside structs.
			if f.Ttoken.Type == token.Ident {
				return fmt.Errorf("type error: struct fields can only be a basic type")
			} else {
				f.T = tokenToType(f.Ttoken)
			}
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
		var signature types.Signature

		if node.Parameter != nil {
			if err := check(node.Parameter, node.SymbolTable); err != nil {
				return err
			}

			signature.Parameter = node.Parameter.T
		}

		var result types.Type
		// TODO: remember that result can be Ident (struct) or a func type.
		if node.Result.Literal != "" {
			if node.Result.Type == token.Ident {
				sym, _ := symbolTable.Resolve(node.Result.Literal)
				result = sym.Type.(*types.Struct)
			} else {
				result = tokenToType(node.Result)
			}
		}
		signature.Result = result

		if err := check(node.Body, node.SymbolTable); err != nil {
			return err
		}

		// TODO: Now we only check that the last statement is a return
		// statement and bail out if it is not. We should probably also loop
		// over the body and check for multiple return statements (early
		// returns) and see if they have the correct type.
		if result != nil {
			lastIndex := len(node.Body.Statements) - 1
			n, ok := node.Body.Statements[lastIndex].(*ast.ReturnStatement)
			if !ok {
				return fmt.Errorf("type error: function: %s, is missing return statement at the end", node.Name)
			}

			if n.Value.Type() != result {
				return fmt.Errorf("type error: function: %s, returns type: %s, but expect to return: %s", node.Name, n.Value.Type(), result)
			}
		}

		node.Name.T = &signature
	case *ast.ReturnStatement:
		if err := check(node.Value, symbolTable); err != nil {
			return err
		}
	case *ast.Identifier:
		sym, _ := symbolTable.Resolve(node.Value)

		switch sym.Type {
		case token.IntType:
			node.T = types.Typ[types.Int]
			sym.UpdateType(node.T)
		case token.FloatType:
			node.T = types.Typ[types.Float]
			sym.UpdateType(node.T)
		case token.StringType:
			node.T = types.Typ[types.String]
			sym.UpdateType(node.T)
		case token.BoolType:
			node.T = types.Typ[types.Bool]
			sym.UpdateType(node.T)
		case token.Ident:
			// TODO: check that if an identifier has another identifier as
			// type, that it is a struct or function.
			// identifier must be a struct or func type.
			s, _ := symbolTable.Resolve(node.Ttoken.Literal)
			node.T = s.Type.(*types.Struct)
			sym.UpdateType(node.T)
		default:
			if _, ok := sym.Type.(ast.TypeNode); ok {
				// let the other switch construct the proper type.
				break
			}

			// assert that the end type is a proper type.
			if _, ok := sym.Type.(types.Type); !ok {
				panic("the symbol should have been updated with the proper type")
			}
		}

		switch v := sym.Type.(type) {
		case *ast.StructType:
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

			node.T = &types.Struct{Fields: fields}
			sym.UpdateType(node.T)
		case *types.Struct:
			node.T = v
		case *types.Basic:
			node.T = v
		default:
			return fmt.Errorf("type error: identifier: %q has the unknown type: %q", node.Value, node.Ttoken)
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
		panic(fmt.Sprintf("tokenToType can not handle this token type: %s", t))
	}

	return typ
}
