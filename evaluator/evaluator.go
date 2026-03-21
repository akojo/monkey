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
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	case left.Type() == object.INTEGER && right.Type() == object.INTEGER:
		leftInt := left.(*object.Integer)
		rightInt := right.(*object.Integer)
		return evalIntegerInfixExpression(op, leftInt, rightInt)
	case left.Type() == object.STRING && right.Type() == object.STRING:
		leftStr := left.(*object.String)
		rightStr := right.(*object.String)
		return evalStringInfixExpression(op, leftStr, rightStr)
	case op == "==":
		return toBoolean(left == right)
	case op == "!=":
		return toBoolean(left != right)
	}
	return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
}

func evalIntegerInfixExpression(op string, left *object.Integer, right *object.Integer) object.Object {
	switch op {
	case "+":
		return &object.Integer{Value: left.Value + right.Value}
	case "-":
		return &object.Integer{Value: left.Value - right.Value}
	case "*":
		return &object.Integer{Value: left.Value * right.Value}
	case "/":
		return &object.Integer{Value: left.Value / right.Value}
	case "<":
		return toBoolean(left.Value < right.Value)
	case ">":
		return toBoolean(left.Value > right.Value)
	case "==":
		return toBoolean(left.Value == right.Value)
	case "!=":
		return toBoolean(left.Value != right.Value)
	}
	return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
}

func evalStringInfixExpression(op string, left *object.String, right *object.String) object.Object {
	if op != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
	return &object.String{Value: left.Value + right.Value}
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
	value, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: %s", node.Value)
	}
	return value
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
	function, ok := fn.(*object.Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}

	result := Eval(function.Body, extendFunctionEnv(function, args))
	return unwrapReturnValue(result)
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
