package evaluator

import (
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
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)

	// Expressions
	case *ast.Boolean:
		return toBoolean(node.Value)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.PrefixExpression:
		return evalPrefixExpression(node.Operator, Eval(node.Right))
	case *ast.InfixExpression:
		return evalInfixExpression(node.Operator, Eval(node.Left), Eval(node.Right))
	}
	return nil
}

func evalStatements(stmts []ast.Statement) object.Object {
	var result object.Object = NULL

	for _, stmt := range stmts {
		result = Eval(stmt)
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
		return NULL
	}
}

func evalBang(right object.Object) object.Object {
	if right == FALSE || right == NULL {
		return TRUE
	}
	return FALSE
}

func evalMinus(right object.Object) object.Object {
	if right.Type() != object.INTEGER {
		return NULL
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(op string, left object.Object, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER && right.Type() == object.INTEGER:
		return evalIntegerInfixExpression(op, left, right)
	case op == "==":
		return toBoolean(left == right)
	case op == "!=":
		return toBoolean(left != right)
	}
	return NULL
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
	return NULL
}

func toBoolean(value bool) object.Object {
	if value {
		return TRUE
	}
	return FALSE
}
