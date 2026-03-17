package evaluator

import (
	"strings"
	"testing"

	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/object"
	"github.com/akojo/monkey/parser"
)

func TestIntegerExpression(t *testing.T) {
	expectIntegerObject(t, eval("5"), 5)
	expectIntegerObject(t, eval("100"), 100)
}

func TestBooleanExpression(t *testing.T) {
	expectBooleanObject(t, eval("true"), true)
	expectBooleanObject(t, eval("false"), false)
}

func eval(input string) object.Object {
	l := lexer.New(strings.NewReader(input), "<test>")
	p := parser.New(l)
	program := p.ParseProgram()

	return Eval(program)
}

func expectIntegerObject(t *testing.T, obj object.Object, expected int64) {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object: expected Integer, got %T", obj)
		return
	}
	if result.Value != expected {
		t.Errorf("result.Value: expected %d, got %d", expected, result.Value)
	}
}

func expectBooleanObject(t *testing.T, obj object.Object, expected bool) {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object: expected Boolean, got %T", obj)
		return
	}
	if result.Value != expected {
		t.Errorf("result.Value: expected %t, got %t", expected, result.Value)
	}
}
