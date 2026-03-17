package evaluator

import (
	"strings"
	"testing"

	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/object"
	"github.com/akojo/monkey/parser"
)

func TestIntegerExpression(t *testing.T) {
	expect(t, "5", 5)
	expect(t, "100", 100)
	expect(t, "-5", -5)
	expect(t, "-100", -100)
}

func TestBooleanExpression(t *testing.T) {
	expect(t, "true", true)
	expect(t, "false", false)
}

func TestBangOperator(t *testing.T) {
	expect(t, "!true", false)
	expect(t, "!false", true)
	expect(t, "!5", false)
	expect(t, "!!true", true)
	expect(t, "!!false", false)
	expect(t, "!!5", true)
}

func eval(input string) object.Object {
	p := parser.New(lexer.New(strings.NewReader(input), "<test>"))
	program := p.ParseProgram()

	return Eval(program)
}

func expect(t *testing.T, input string, expected any) {
	got := eval(input)
	switch expected := expected.(type) {
	case int64:
		expectIntegerObject(t, got, expected)
	case bool:
		expectBooleanObject(t, got, expected)
	}
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
