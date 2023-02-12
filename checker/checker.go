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

		if !reflect.DeepEqual(node.Name.T, node.Value.Type()) {
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
			switch t := f.Tnode.(type) {
			case *ast.BasicType:
				// In struct only allow for basic types
				if t.Token.Type == token.Ident {
					return fmt.Errorf("type error: struct fields can only be a basic type [int, float, bool, string]")
				}
				f.T = typeNodetoType(t)
			default:
				// TODO: maybe later allow for struct inside structs.
				panic("StructType containing fields of other types than BasicType is not implemented")
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
			offset, ok := identifierInStruct(node.Field, x)
			if !ok {
				return fmt.Errorf(
					"type error: identifier: %s is not a field in struct: %s",
					node.Field.Value,
					v.Value,
				)
			}
			node.Offset = offset
		case *ast.CallExpression:
			x, ok := v.T.(*types.Struct)
			if !ok {
				return fmt.Errorf(
					"type error: selecting field on identifier: %s, which is not a struct",
					v,
				)
			}
			offset, ok := identifierInStruct(node.Field, x)
			if !ok {
				return fmt.Errorf(
					"type error: identifier: %s is not a field in struct: %s",
					node.Field.Value,
					v,
				)
			}
			node.Offset = offset
		default:
			return fmt.Errorf("type error: can not handle the expression for X in selection. got=%T", v)
		}

		node.T = node.Field.T
	case *ast.AssignStatement:
		if err := check(node.Name, symbolTable); err != nil {
			return err
		}

		if err := check(node.Value, symbolTable); err != nil {
			return err
		}

		if !reflect.DeepEqual(node.Name.Type(), node.Value.Type()) {
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
		// defined within a function. Remove the function after it has been
		// used.
		defer func() { currentFunc = nil }()

		signature, err := funcTypeToSignature(node.Signature, node.SymbolTable)
		if err != nil {
			return err
		}

		node.Name.T = signature

		if node.Body != nil {
			// special case that a prototype have been defined
			sym, _ := symbolTable.Resolve(node.Name.Value)
			if v, ok := sym.Type.(*types.Signature); ok {
				if !reflect.DeepEqual(*signature, *v) {
					return fmt.Errorf("type error: function: %q's prototype and definition differ in signature", node.Name.Value)
				}
			}

			if err := check(node.Body, node.SymbolTable); err != nil {
				return err
			}

			if signature.Result.Kind() != types.Nil {
				lastIndex := len(node.Body.Statements) - 1
				_, ok := node.Body.Statements[lastIndex].(*ast.ReturnStatement)
				if !ok {
					return fmt.Errorf("type error: function: %s, is missing return statement at the end", node.Name)
				}
			}
		}

		// Update the symbol of function to its proper type.
		sym, _ := symbolTable.Resolve(node.Name.Value)
		sym.Type = node.Name.T
	case *ast.ReturnStatement:
		if currentFunc == nil {
			return fmt.Errorf("checker error: Return statement can not be declared outside of function")
		}

		// Save the function identifier into the return.
		node.Function = currentFunc

		result := currentFunc.T.(*types.Signature).Result

		if node.Value == nil {
			if result.Kind() != types.Nil {
				return fmt.Errorf("type error: function: %q was expected to return %q", currentFunc.Value, result)
			}

			return nil
		}

		if err := check(node.Value, symbolTable); err != nil {
			return err
		}

		if !reflect.DeepEqual(node.Value.Type(), result) {
			return fmt.Errorf("type error: function: %q, returns type: %s, but expected to return: %s", currentFunc.Value, node.Value.Type(), result)
		}
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
		sym, ok := symbolTable.Resolve(node.Value)
		if !ok {
			if node.Token.Type == token.Blank {
				node.T = typeNodetoType(node.Tnode)
				return nil
			}
		}

		switch v := sym.Type.(type) {
		case *ast.FuncType:
			signature, err := funcTypeToSignature(v, symbolTable)
			if err != nil {
				return err
			}

			node.T = signature
			sym.Type = node.T
		case *ast.BasicType:
			if v.Token.Type == token.Ident {
				s, ok := symbolTable.Resolve(v.Token.Literal)
				if !ok {
					return fmt.Errorf(
						"checker error: identifier %q is not defined",
						v.Token.Literal,
					)
				}
				if s.Scope != symbol.TypeScope {
					return fmt.Errorf(
						"checker error: identifier %q is not a type",
						v.Token.Literal,
					)
				}
				node.T = s.Type.(*types.Struct)
				sym.Type = node.T
			} else {
				node.T = tokenToType(v.Token)
				sym.Type = node.T
			}
		case *types.Basic:
			node.T = v
		case *types.Signature:
			node.T = v
		case *types.Struct:
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
			return fmt.Errorf("type error: identifier: %q has the unknown type: %q", node.Value, node.Tnode)
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
func identifierInStruct(id *ast.Identifier, s *types.Struct) (int, bool) {
	const wordAllignment = 8
	for i, f := range s.Fields {
		if f.Name == id.Value {
			id.T = f.Type
			return i * wordAllignment, true
		}
	}

	return 0, false
}

func typeNodetoType(t ast.TypeNode) types.Type {
	switch t := t.(type) {
	case *ast.BasicType:
		return tokenToType(t.Token)
	default:
		panic("other type nodes are not implemented")
	}
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
		panic(fmt.Sprintf("tokenToType can not handle this token type: %v", t))
	}

	return typ
}

func funcTypeToSignature(ft *ast.FuncType, symbolTable *symbol.Table) (*types.Signature, error) {
	var signature types.Signature

	var parameter types.Type
	if ft.Parameter != nil {
		if err := check(ft.Parameter, symbolTable); err != nil {
			return nil, err
		}

		parameter = ft.Parameter.T
	} else {
		parameter = types.Typ[types.Nil]
	}

	signature.Parameter = parameter

	var result types.Type
	switch t := ft.Result.(type) {
	case *ast.BasicType:
		if t.Token.Type == token.Ident {
			sym, _ := symbolTable.Resolve(t.Token.Literal)
			if sym.Scope != symbol.TypeScope {
				return nil, fmt.Errorf("checker error: identifier %q is not a type", t.Token.Literal)
			}
			result = sym.Type.(*types.Struct)
			break
		}

		result = typeNodetoType(t)
	case *ast.FuncType:
		res, err := funcTypeToSignature(t, symbolTable)
		if err != nil {
			return nil, err
		}

		result = res
	case *ast.StructType:
		panic("not implemented")
	default:
		result = types.Typ[types.Nil]
	}
	signature.Result = result

	return &signature, nil
}
