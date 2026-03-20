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

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements)
	case *ast.ReturnStatement:
		rv := Eval(node.ReturnValue)
		if isError(rv) {
			return rv
		}
		return &object.Return{Value: rv}

	// Expressions
	case *ast.Boolean:
		return toBoolean(node.Value)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		if isError(left) {
			return left
		}
		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IFExpression:
		return evalIfExpression(node)
	}
	return nil
}

func evalProgram(stmts []ast.Statement) object.Object {
	var result object.Object = NULL

	for _, stmt := range stmts {
		result = Eval(stmt)

		switch result := result.(type) {
		case *object.Return:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(stmts []ast.Statement) object.Object {
	var result object.Object = NULL

	for _, stmt := range stmts {
		result = Eval(stmt)

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
		return evalIntegerInfixExpression(op, left, right)
	case op == "==":
		return toBoolean(left == right)
	case op == "!=":
		return toBoolean(left != right)
	}
	return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
}

func evalIntegerInfixExpression(op string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch op {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return toBoolean(leftVal < rightVal)
	case ">":
		return toBoolean(leftVal > rightVal)
	case "==":
		return toBoolean(leftVal == rightVal)
	case "!=":
		return toBoolean(leftVal != rightVal)
	}
	return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
}

func evalIfExpression(ie *ast.IFExpression) object.Object {
	condition := Eval(ie.Condition)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative)
	} else {
		return NULL
	}
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
