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
	expect(t, "true == true", true)
	expect(t, "false == false", true)

	// XOR
	expect(t, "false != false", false)
	expect(t, "false != true", true)
	expect(t, "true != false", true)
	expect(t, "true != true", false)

	// OR
	expect(t, "false + false", false)
	expect(t, "false + true", true)
	expect(t, "true + false", true)
	expect(t, "true + true", true)

	// AND
	expect(t, "false * false", false)
	expect(t, "false * true", false)
	expect(t, "true * false", false)
	expect(t, "true * true", true)

	expect(t, "1 < 2", true)
	expect(t, "1 > 2", false)
	expect(t, "1 < 1", false)
	expect(t, "1 > 1", false)
	expect(t, "1 == 1", true)
	expect(t, "1 != 1", false)
	expect(t, "1 == 2", false)
	expect(t, "1 != 2", true)

	expect(t, `"foo" == "foo"`, true)
	expect(t, `"foo" != "foo"`, false)
	expect(t, `"foo" == "bar"`, false)
	expect(t, `"foo" != "bar"`, true)

	expect(t, "[] == []", true)
	expect(t, "[1, 2] == [1, 2]", true)
	expect(t, "[1, 2] != [1, 2]", false)
	expect(t, "[1, 2] == [2, 3]", false)
	expect(t, "[1, 2] == [1, 2, 3]", false)
	expect(t, "[1, [2, 3]] == [1, [2, 3]]", true)
	expect(t, `["a", "b"] == ["a", "b"]`, true)
	expect(t, `[true, 1, "2", [3]] == [true, 1, "2", [3]]`, true)

	expect(t, "(1 < 2) == true", true)
	expect(t, "(1 < 2) == false", false)
	expect(t, "(1 > 2) != true", true)
	expect(t, "(1 > 2) != false", false)

	expect(t, "1 == true", false)
	expect(t, "1 == false", false)
	expect(t, `1 == "1"`, false)
	expect(t, `0 == []`, false)
	expect(t, `"true" == true`, false)
	expect(t, `"true" == false`, false)
	expect(t, `"" == true`, false)
	expect(t, `"" == false`, false)
	expect(t, `"" == []`, false)
	expect(t, "[1] == true", false)
	expect(t, "[1] == false", false)
	expect(t, "[] == true", false)
	expect(t, "[] == false", false)
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
	expect(t, "foo", errors.New("identifier not found: foo"))
	expect(t, `"hello" - "world"`, errors.New("unknown operator: STRING - STRING"))
	expect(t, `{"name": "Monkey"}[fn(x) { x + 1 }]`, errors.New("cannot use as hash key: FUNCTION"))
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
	expect(t, "len([1, 2, 3])", 3)
	expect(t, "len([])", 0)

	expect(t, "let a = append([1, 2], 3); a == [1, 2, 3]", true)
	expect(t, "let a = append([], 1); a == [1]", true)
	expect(t, "let a = append([], [1]); a == [[1]]", true)

	expect(t, "equals(1, 1)", true)
	expect(t, `equals("foo", "foo")`, true)
	expect(t, "equals([1,2], [1,2])", true)
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := eval(input)

	array, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("array: expected *object.Array, got %T", evaluated)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("array.Elements: expected 3 elements, got %d", len(array.Elements))
	}

	expectIntegerObject(t, array.Elements[0], 1)
	expectIntegerObject(t, array.Elements[1], 4)
	expectIntegerObject(t, array.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	expect(t, "[1, 2, 3][0]", 1)
	expect(t, "[1, 2, 3][2]", 3)
	expect(t, "[1, 2, 3][3]", nil)
	expect(t, "let i = 0; [1, 2][i + 1]", 2)
	expect(t, "let a = [1, 2, 3]; a[2]", 3)
	expect(t, "let a = [1, 2, 3]; a[0] * a[1]", 2)
}

func TestArraySlices(t *testing.T) {
	expect(t, "let a = [1, 2, 3, 4]; slice(a, 1, 3) == [2, 3]", true)

	expect(t, "let a = [1, 2, 3, 4][1:3]; a == [2, 3]", true)
	expect(t, "let a = [1, 2, 3, 4][1:]; a == [2, 3, 4]", true)
	expect(t, "let a = [1, 2, 3, 4][:2]; a == [1, 2]", true)
}

func TestArrayConcatenation(t *testing.T) {
	expect(t, "[1, 2] + [3, 4] == [1, 2, 3, 4]", true)
	expect(t, "[1, 2] + [] == [1, 2]", true)
	expect(t, "[] + [] == []", true)
}

func TestHashLiterals(t *testing.T) {
	test := func(hash *object.Hash, key object.HashKey, value any) {
		pair, ok := hash.Pairs[key]
		if !ok {
			t.Errorf("no pair for %v", key)
		}
		expectObject(t, pair.Value, value)
	}

	evaluated := eval(`
	let two = "two";
	{
		"one": 10 - 9,
		two: "two",
		4: 4,
		true: true,
	}`)
	hash, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("result: expected *object.Hash, got %T (%+v)", evaluated, evaluated)
	}
	if len(hash.Pairs) != 4 {
		t.Fatalf("hash.Pairs: expected 4, got %d", len(hash.Pairs))
	}

	test(hash, (&object.String{Value: "one"}).Hash(), 1)
	test(hash, (&object.String{Value: "two"}).Hash(), "two")
	test(hash, (&object.Integer{Value: 4}).Hash(), 4)
	test(hash, (&object.Boolean{Value: true}).Hash(), true)
}

func TestHashIndexExpressions(t *testing.T) {
	expect(t, `{"foo": 5}["foo"]`, 5)
	expect(t, `{"foo": 5}["bar"]`, nil)
	expect(t, `{}["foo"]`, nil)
	expect(t, `{5: 10}[5]`, 10)
	expect(t, `{true: 5}[true]`, 5)
	expect(t, `{false: 5}[false]`, 5)
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
	expectObject(t, got, expected)
}

func expectObject(t *testing.T, got object.Object, expected any) {
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
		t.Errorf("result: expected Integer, got %T: %s", obj, obj.Inspect())
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
