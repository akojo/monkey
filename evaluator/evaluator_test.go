package evaluator

import (
	"strings"
	"testing"

	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/object"
	"github.com/akojo/monkey/parser"
)

func TestIntegerExpression(t *testing.T) {
	test := func(input string, expected int64) {
		value := eval(input)
		expectIntegerOjbject(t, value, expected)
	}
	test("5", 5)
	test("100", 100)
}

func eval(input string) object.Object {
	l := lexer.New(strings.NewReader(input), "<test>")
	p := parser.New(l)
	program := p.ParseProgram()

	return Eval(program)
}

func expectIntegerOjbject(t *testing.T, obj object.Object, expected int64) {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object: expected Integer, got %T", obj)
		return
	}
	if result.Value != expected {
		t.Errorf("result.Value: expected %d, got %d", expected, result.Value)
	}
}
