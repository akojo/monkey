package evaluator

import (
	"fmt"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

var builtins = map[string]*object.Builtin{
	"append": {Fn: builtin_append},
	"equals": {Fn: builtin_equals},
	"len":    {Fn: builtin_len},
	"print":  {Fn: builtin_print},
	"slice":  {Fn: builtin_slice},
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, env)
	case *ast.ReturnStatement:
		rv := Eval(node.ReturnValue, env)
		if isError(rv) {
			return rv
		}
		return &object.Return{Value: rv}
	case *ast.LetStatement:
		value := Eval(node.Value, env)
		if isError(value) {
			return value
		}
		env.Set(node.Name.Value, value)

	// Expressions
	case *ast.Boolean:
		return toBoolean(node.Value)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IFExpression:
		return evalIfExpression(node, env)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	case *ast.SliceExpression:
		var start object.Object
		var end object.Object
		if node.Start != nil {
			start = Eval(node.Start, env)
			if isError(start) {
				return start
			}
		}
		if node.End != nil {
			end = Eval(node.End, env)
			if isError(end) {
				return end
			}
		}
		return evalSliceExpression(start, end)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}
	return nil
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object = NULL

	for _, stmt := range stmts {
		result = Eval(stmt, env)

		switch result := result.(type) {
		case *object.Return:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object = NULL

	for _, stmt := range stmts {
		result = Eval(stmt, env)

		if isError(result) || isReturn(result) {
			break
		}
	}

	return result
}

func evalPrefixExpression(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return evalBang(right)
	case "-":
		return evalMinus(right)
	default:
		return newError("unknown operator: %s%s", op, right.Type())
	}
}

func evalBang(right object.Object) object.Object {
	return toBoolean(!isTruthy(right))
}

func evalMinus(right object.Object) object.Object {
	if right.Type() != object.INTEGER {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(op string, left object.Object, right object.Object) object.Object {
	switch {
	case op == "==":
		return toBoolean(equals(left, right))
	case op == "!=":
		return toBoolean(!equals(left, right))
	case op == "+":
		return add(left, right)
	case op == "*":
		return multiply(left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	case left.Type() == object.INTEGER && right.Type() == object.INTEGER:
		leftInt := left.(*object.Integer)
		rightInt := right.(*object.Integer)
		return evalIntegerInfixExpression(op, leftInt, rightInt)
	}
	return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
}

func evalIntegerInfixExpression(op string, left *object.Integer, right *object.Integer) object.Object {
	switch op {
	case "-":
		return &object.Integer{Value: left.Value - right.Value}
	case "/":
		return &object.Integer{Value: left.Value / right.Value}
	case "<":
		return toBoolean(left.Value < right.Value)
	case ">":
		return toBoolean(left.Value > right.Value)
	}
	return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
}

func evalIfExpression(ie *ast.IFExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if value, ok := env.Get(node.Value); ok {
		return value
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: %s", node.Value)
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	results := make([]object.Object, 0)

	for _, exp := range exps {
		result := Eval(exp, env)
		if isError(result) {
			return []object.Object{result}
		}
		results = append(results, result)
	}

	return results
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		result := Eval(fn.Body, extendFunctionEnv(fn, args))
		return unwrapReturnValue(result)
	case *object.Builtin:
		return fn.Fn(args...)
	}
	return newError("not a function: %s", fn.Type())
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := fn.Env.Extend()

	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.Return); ok {
		return returnValue.Value
	}
	return obj
}

func evalIndexExpression(left object.Object, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INTEGER:
		array := left.(*object.Array)
		i := index.(*object.Integer).Value
		if i < 0 || i > int64(len(array.Elements)-1) {
			return NULL
		}
		return array.Elements[i]
	case left.Type() == object.ARRAY && index.Type() == object.SLICE:
		array := left.(*object.Array)
		slice := index.(*object.Slice)
		return evalSliceIndexExpression(array, slice)
	case left.Type() == object.HASH:
		return evalHashIndexExpression(left, index)
	}
	return newError("index operator not supported: %s", left.Type())
}

func evalSliceIndexExpression(array *object.Array, sliceObj *object.Slice) object.Object {
	var end int64
	if sliceObj.End == nil {
		end = int64(len(array.Elements))
	} else {
		end = *sliceObj.End
	}

	return slice(array, sliceObj.Start, end)
}

func evalSliceExpression(start object.Object, end object.Object) object.Object {
	switch {
	case start == nil && end == nil:
		return &object.Slice{Start: 0, End: nil}
	case start == nil && end != nil && end.Type() == object.INTEGER:
		return &object.Slice{Start: 0, End: &end.(*object.Integer).Value}
	case start != nil && start.Type() == object.INTEGER && end == nil:
		return &object.Slice{Start: start.(*object.Integer).Value, End: nil}
	case start.Type() == object.INTEGER && end.Type() == object.INTEGER:
		return &object.Slice{Start: start.(*object.Integer).Value, End: &end.(*object.Integer).Value}
	}
	return newError("slice operator not supported: %s:%s", start.Type(), end.Type())
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("cannot use as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.Hash()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("cannot use as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.Hash()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func toBoolean(value bool) object.Object {
	if value {
		return TRUE
	}
	return FALSE
}

func isTruthy(obj object.Object) bool {
	if obj == FALSE || obj == NULL {
		return false
	}
	return true
}

func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ERROR
}

func isReturn(obj object.Object) bool {
	return obj != nil && obj.Type() == object.RETURN
}

func newError(format string, a ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
