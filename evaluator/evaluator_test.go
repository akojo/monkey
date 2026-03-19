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
	expect(t, "5 + 5 + 5", 15)
	expect(t, "2 * 2 * 2 * 2 * 2", 32)
	expect(t, "-50 + 100 + -50", 0)
	expect(t, "5 * 2 + 10", 20)
	expect(t, "20 + 2 * -10", 0)
	expect(t, "50 / 2 * 2 + 10", 60)
	expect(t, "2 * (5 + 10)", 30)
	expect(t, "3 * 3 * 3 + 10", 37)
	expect(t, "3 * (3 * 3) + 10", 37)
	expect(t, "(5 + 10 * 2 + 15 / 3) * 2 + -10", 50)
}

func TestBooleanExpression(t *testing.T) {
	expect(t, "true", true)
	expect(t, "false", false)
	expect(t, "1 < 2", true)
	expect(t, "1 > 2", false)
	expect(t, "1 < 1", false)
	expect(t, "1 > 1", false)
	expect(t, "1 == 1", true)
	expect(t, "1 != 1", false)
	expect(t, "1 == 2", false)
	expect(t, "1 != 2", true)
	expect(t, "true == true", true)
	expect(t, "false == false", true)
	expect(t, "true != false", true)
	expect(t, "false != true", true)
	expect(t, "(1 < 2) == true", true)
	expect(t, "(1 < 2) == false", false)
	expect(t, "(1 > 2) != true", true)
	expect(t, "(1 > 2) != false", false)
}

func TestBangOperator(t *testing.T) {
	expect(t, "!true", false)
	expect(t, "!false", true)
	expect(t, "!5", false)
	expect(t, "!!true", true)
	expect(t, "!!false", false)
	expect(t, "!!5", true)
}

func TestIfElseExpression(t *testing.T) {
	expect(t, "if (true) { 10 }", 10)
	expect(t, "if (false) { 10 }", nil)
	expect(t, "if (1) { 10 }", 10)
	expect(t, "if (1 < 2) { 10 }", 10)
	expect(t, "if (1 > 2) { 10 }", nil)
	expect(t, "if (1 < 2) { 10 } else { 20 }", 10)
	expect(t, "if (1 > 2) { 10 } else { 20 }", 20)
}

func eval(input string) object.Object {
	p := parser.New(lexer.New(strings.NewReader(input), "<test>"))
	program := p.ParseProgram()

	return Eval(program)
}

func expect(t *testing.T, input string, expected any) {
	got := eval(input)
	if got == nil {
		t.Errorf("got unexpected nil")
		return
	}

	switch e := expected.(type) {
	case int:
		expectIntegerObject(t, got, int64(e))
	case int64:
		expectIntegerObject(t, got, e)
	case bool:
		expectBooleanObject(t, got, e)
	case nil:
		expectNullObject(t, got)
	default:
		t.Fatalf("invalid type %T", e)
	}
}

func expectIntegerObject(t *testing.T, obj object.Object, expected int64) {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("result: expected Integer, got %T", obj)
		return
	}
	if result.Value != expected {
		t.Errorf("result.Value: expected %d, got %d", expected, result.Value)
	}
}

func expectBooleanObject(t *testing.T, obj object.Object, expected bool) {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("result: expected Boolean, got %q", obj.Inspect())
		return
	}
	if result.Value != expected {
		t.Errorf("result.Value: expected %t, got %t", expected, result.Value)
	}
}

func expectNullObject(t *testing.T, obj object.Object) {
	if obj != NULL {
		t.Errorf("expected NULL, got %q", obj.Inspect())
	}
}
