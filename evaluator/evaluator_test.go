package evaluator

import (
	"errors"
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

func TestReturnStatement(t *testing.T) {
	expect(t, "return 10;", 10)
	expect(t, "return 10; 9;", 10)
	expect(t, "return 2 * 5; 9", 10)
	expect(t, "9; return 2 * 5; 9", 10)
	expect(t, "if (true) { if (true) { return 10; } return 1;}", 10)
	expect(t, "let f = fn(x) { return x; return x + 10 }; f(10);", 10)
	expect(t, `
		let f = fn(x) {
			let g = fn(y) {
				return y + 5
			}
			return g(x) + 5
		}
		f(5)
	`, 15)
}

func TestErrorHandling(t *testing.T) {
	expect(t, "5 + true;", errors.New("type mismatch: INTEGER + BOOLEAN"))
	expect(t, "5 + true; 5", errors.New("type mismatch: INTEGER + BOOLEAN"))
	expect(t, "-true", errors.New("unknown operator: -BOOLEAN"))
	expect(t, "5; true + false; 5;", errors.New("unknown operator: BOOLEAN + BOOLEAN"))
	expect(t, "if (10 > 1) { true + false; }", errors.New("unknown operator: BOOLEAN + BOOLEAN"))
	expect(t, "foo", errors.New("identifier not found: foo"))
	expect(t, `"hello" - "world"`, errors.New("unknown operator: STRING - STRING"))
}

func TestLetStatements(t *testing.T) {
	expect(t, "let a = 5; a;", 5)
	expect(t, "let a = 5 * 5; a", 25)
	expect(t, "let a = 5; let b = a; b", 5)
	expect(t, "let a = 5; let b = 5; let c = a + b + 5; c;", 15)
}

func TestFunctionObject(t *testing.T) {
	evaluated := eval("fn(x) { x + 2; };")

	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("fn: expected *object.Function, got %T", evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("fn.Parameters: expected 1, got %d", len(fn.Parameters))
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("fn.Parameters[0]: expected \"x\", got %q", fn.Parameters[0].String())
	}

	if len(fn.Body.Statements) != 1 {
		t.Fatalf("fn.Body: expected 1 statement, got %d", len(fn.Body.Statements))
	}

	expectedBody := "(x + 2)"
	gotBody := fn.Body.Statements[0].String()
	if gotBody != expectedBody {
		t.Fatalf("fn.Body: expected %q, got %q", expectedBody, gotBody)
	}
}

func TestFunctionCall(t *testing.T) {
	expect(t, "let id = fn(x) { x; }; id(5);", 5)
	expect(t, "let id = fn(x) { return x;}; id(5);", 5)
	expect(t, "let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20)
	expect(t, "fn(x) { -x }(5)", -5)
	expect(t, `
		let f = fn(x) {
			let g = fn(y) {
				return x + y
			}
			return g(5)
		}
		f(5)
	`, 10)
	expect(t, `
		let add = fn(x) { fn(y) { x + y } }
		let add2 = add(2)
		add2(2)
	`, 4)
}

func TestStringLiteral(t *testing.T) {
	expect(t, `"hello, world"`, "hello, world")
}

func TestStringConcatenation(t *testing.T) {
	expect(t, `"hello" + " " + "world"`, "hello world")
}

func TestBuiltinFunctions(t *testing.T) {
	expect(t, `len("")`, 0)
	expect(t, `len("mitä")`, 4)
	expect(t, `len("💩")`, 1)
	expect(t, `len("hello world")`, 11)
	expect(t, "len(1)", errors.New("argument to `len` not supported, got INTEGER"))
	expect(t, `len("one", "two")`, errors.New("wrong number of arguments: got 2, want 1"))
}

func eval(input string) object.Object {
	p := parser.New(lexer.New(strings.NewReader(input), "<test>"))
	program := p.ParseProgram()

	return Eval(program, object.NewEnvironment())
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
	case string:
		expectStringObject(t, got, e)
	case error:
		expectErrorObject(t, got, e.Error())
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

func expectStringObject(t *testing.T, obj object.Object, expected string) {
	str, ok := obj.(*object.String)
	if !ok {
		t.Fatalf("str: expected String, got %q", obj.Type())
	}
	if str.Value != expected {
		t.Errorf("str.Value: expected %q, got %q", expected, str.Value)
	}
}

func expectErrorObject(t *testing.T, obj object.Object, expected string) {
	err, ok := obj.(*object.Error)
	if !ok {
		t.Errorf("result: expected Error: got %q", obj.Inspect())
		return
	}
	if err.Message != expected {
		t.Errorf("err.Message: expected %q, got %q", expected, err.Message)
	}
}

func expectNullObject(t *testing.T, obj object.Object) {
	if obj != NULL {
		t.Errorf("expected NULL, got %q", obj.Inspect())
	}
}
