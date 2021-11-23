// Package checker type checks the AST produced by the parser package and
// annotates the AST nodes with the proper types.
package checker

import (
	"fmt"
	"reflect"

	"github.com/Glorforidor/didactic_compiler/ast"
	"github.com/Glorforidor/didactic_compiler/symbol"
	"github.com/Glorforidor/didactic_compiler/token"
	"github.com/Glorforidor/didactic_compiler/types"
)

func Check(program *ast.Program) error {
	return check(program, program.SymbolTable)
}


var currentFunc *ast.Identifier

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
	case *ast.SelectorExpression:
		if err := check(node.X, symbolTable); err != nil {
			return err
		}

		switch v := node.X.(type) {
		case *ast.Identifier:
			x, ok := v.T.(*types.Struct)
			if !ok {
				return fmt.Errorf(
					"type error: selecting field on identifier: %s, which is not a struct",
					v,
				)
			}
			if !identifierInStruct(node.Field, x) {
				return fmt.Errorf(
					"type error: identifier: %s is not a field in struct: %s",
					node.Field.Value,
					v.Value,
				)
			}

		default:
			return fmt.Errorf("type error: select X is not an Identifier")
		}

		node.T = node.Field.T
	case *ast.AssignStatement:
		if err := check(node.Name, symbolTable); err != nil {
			return err
		}

		if err := check(node.Value, symbolTable); err != nil {
			return err
		}

		if node.Name.Type() != node.Value.Type() {
			return fmt.Errorf(
				"type error: identifier: %q of type: %s is assigned the wrong type: %s",
				node.Name,
				node.Name.Type(),
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
		currentFunc = node.Name
		// TODO: Could maybe have a stack of functions if the function is
		// defined within a function.
		// Remove the function after it has been used.
		defer func() { currentFunc = nil }()

		var signature types.Signature

		var parameter types.Type
		if node.Signature.Parameter != nil {
			if err := check(node.Signature.Parameter, node.SymbolTable); err != nil {
				return err
			}

			parameter = node.Signature.Parameter.T
		} else {
			parameter = types.Typ[types.Nil]
		}
		signature.Parameter = parameter

		var result types.Type
		// TODO: remember that result can be Ident (struct) or a func type.
		if node.Signature.Result.Literal != "" {
			if node.Signature.Result.Type == token.Ident {
				sym, _ := symbolTable.Resolve(node.Signature.Result.Literal)
				result = sym.Type.(*types.Struct)
			} else {
				result = tokenToType(node.Signature.Result)
			}
		} else {
			result = types.Typ[types.Nil]
		}
		signature.Result = result

		if node.Body != nil {
			// special case that a prototype have been defined
			sym, _ := symbolTable.Resolve(node.Name.Value)
			if v, ok := sym.Type.(*types.Signature); ok {
				if !reflect.DeepEqual(signature, *v) {
					return fmt.Errorf("type error: function: %q's prototype and definition differ in signature", node.Name.Value)
				}
			}

			if err := check(node.Body, node.SymbolTable); err != nil {
				return err
			}

			// TODO: Now we only check that the last statement is a return
			// statement and bail out if it is not. We should probably also loop
			// over the body and check for multiple return statements (early
			// returns) and see if they have the correct type.
			if result.Kind() != types.Nil {
				lastIndex := len(node.Body.Statements) - 1
				n, ok := node.Body.Statements[lastIndex].(*ast.ReturnStatement)
				if !ok {
					return fmt.Errorf("type error: function: %s, is missing return statement at the end", node.Name)
				}

				if n.Value.Type() != result {
					return fmt.Errorf("type error: function: %s, returns type: %s, but expect to return: %s", node.Name, n.Value.Type(), result)
				}
			}
		}

		node.Name.T = &signature

		// Update the symbol of function to its proper type.
		sym, _ := symbolTable.Resolve(node.Name.Value)
		sym.Type = node.Name.T
	case *ast.ReturnStatement:
		if currentFunc == nil {
			return fmt.Errorf("checker error: Return statement can not be declared outside of function")
		}

		if err := check(node.Value, symbolTable); err != nil {
			return err
		}

		// Save the function identifier into the return.
		node.Function = currentFunc
	case *ast.CallExpression:
		if err := check(node.Function, symbolTable); err != nil {
			return err
		}

		if node.Argument != nil {
			if err := check(node.Argument, symbolTable); err != nil {
				return err
			}
		}

		sig, ok := node.Function.Type().(*types.Signature)
		if !ok {
			return fmt.Errorf(
				"type error: identifier: %q is not a function",
				node.Function.TokenLiteral(),
			)
		}

		if node.Argument == nil {
			if sig.Parameter.Kind() != types.Nil {
				// TODO: Better error message here will be great.
				return fmt.Errorf("type error: function: %q takes arguments, non was provided", node.Function.(*ast.Identifier).Value)
			}
		} else if !reflect.DeepEqual(sig.Parameter, node.Argument.Type()) {
			return fmt.Errorf(
				"type error: wrong argument type for %q, expected: %s, got: %s",
				node.Function.TokenLiteral(),
				sig.Parameter,
				node.Argument.Type(),
			)
		}

		node.T = sig.Result
	case *ast.Identifier:
		sym, _ := symbolTable.Resolve(node.Value)

		switch v := sym.Type.(type) {
		case token.TokenType:
			switch v {
			case token.IntType:
				node.T = types.Typ[types.Int]
				sym.Type = node.T
			case token.FloatType:
				node.T = types.Typ[types.Float]
				sym.Type = node.T
			case token.StringType:
				node.T = types.Typ[types.String]
				sym.Type = node.T
			case token.BoolType:
				node.T = types.Typ[types.Bool]
				sym.Type = node.T
			case token.Ident:
				// TODO: check that if an identifier has another identifier as
				// type, that it is a struct or function.
				// identifier must be a struct or func type.
				s, _ := symbolTable.Resolve(node.Ttoken.Literal)
				node.T = s.Type.(*types.Struct)
				sym.Type = node.T
			}
		case *types.Struct:
			node.T = v
		case *types.Basic:
			node.T = v
		case *types.Signature:
			node.T = v
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
			sym.Type = node.T
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
		return fmt.Errorf("type error: ast node not handled: %T", node)
	}

	return nil
}

// identifierInStruct checks if the identifier is in the struct. If it is, then
// updates that identifier with the same type as the one in the struct and
// returns true. Otherwise returns false.
func identifierInStruct(id *ast.Identifier, s *types.Struct) bool {
	for _, f := range s.Fields {
		if f.Name == id.Value {
			id.T = f.Type
			return true
		}
	}

	return false
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
