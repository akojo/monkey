package vm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/compiler"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/lib"
	"github.com/akojo/monkey/object"
	"github.com/akojo/monkey/parser"
)

func TestIntegerArithmetic(t *testing.T) {
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

func TestBooleanExpressions(t *testing.T) {
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
}

func TestStringExpressions(t *testing.T) {
	expect(t, `"monkey"`, "monkey")
	expect(t, `"mon" + "key"`, "monkey")
	expect(t, `"mon" + "key" + "banana"`, "monkeybanana")
}

func TestArrayLiterals(t *testing.T) {
	expect(t, "[]", []int{})
	expect(t, "[1, 2, 3]", []int{1, 2, 3})
	expect(t, "[1 + 2, 3 - 4, 5 * 6]", []int{3, -1, 30})
}

func TestBangOperator(t *testing.T) {
	expect(t, "!true", false)
	expect(t, "!false", true)
	expect(t, "!5", false)
	expect(t, "!!true", true)
	expect(t, "!!false", false)
	expect(t, "!!5", true)
	expect(t, "!(if (false) {})", true)
}

func TestIfElseExpression(t *testing.T) {
	expect(t, "if (true) { 10 }", 10)
	expect(t, "if (false) { 10 }", nil)
	expect(t, "if (1) { 10 }", 10)
	expect(t, "if (1 < 2) { 10 }", 10)
	expect(t, "if (1 > 2) { 10 }", nil)
	expect(t, "if (1 < 2) { 10 } else { 20 }", 10)
	expect(t, "if (1 > 2) { 10 } else { 20 }", 20)
	expect(t, "if (if (false) {}) { 10 } else { 20 }", 20)
	expect(t, "if (true) {}", nil)
}

func TestGlobalLetStatements(t *testing.T) {
	expect(t, "let one = 1; 1", 1)
	expect(t, "let one = 1; let two = 2; one + two;", 3)
	expect(t, "let one = 1; let two = one + one; one + two", 3)
}

func expect(t *testing.T, input string, expected any) {
	t.Helper()

	program := parse(input)

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(comp.Bytecode())
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	if err := expectObject(expected, vm.StackAboveTop()); err != nil {
		t.Errorf("%q: %s", input, err)
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(strings.NewReader(input), "test")
	p := parser.New(l)
	return p.ParseProgram()
}

func expectObject(expected any, actual object.Object) error {
	var err error
	switch expected := expected.(type) {
	case int:
		err = expectInteger(int64(expected), actual)
	case bool:
		err = expectBoolean(bool(expected), actual)
	case string:
		err = expectString(expected, actual)
	case []int:
		err = expectIntegerArray(expected, actual)
	case nil:
		if actual != lib.NULL {
			return fmt.Errorf("expected NULL, got %q", actual)
		}
	default:
		return fmt.Errorf("unsupported type %T", expected)
	}
	return err
}

func expectInteger(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("want Integer, got %T (%+v)", actual, actual)
	}

	if expected != result.Value {
		return fmt.Errorf("want %d, got %d", expected, result.Value)
	}

	return nil
}

func expectBoolean(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("want Boolean, got %T (%+v)", actual, actual)
	}

	if expected != result.Value {
		return fmt.Errorf("want %t, got %t", expected, result.Value)
	}

	return nil
}

func expectString(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("want String, got %T (%+v)", actual, actual)
	}

	if expected != result.Value {
		return fmt.Errorf("want %s, got %s", expected, result.Value)
	}

	return nil
}

func expectIntegerArray(expected []int, actual object.Object) error {
	array, ok := actual.(*object.Array)
	if !ok {
		return fmt.Errorf("want Array, got %T (%+v)", actual, actual)
	}

	if len(array.Elements) != len(expected) {
		return fmt.Errorf("array.Elements: want %d, got %d", len(expected), len(array.Elements))
	}

	for i, expectedElement := range expected {
		err := expectInteger(int64(expectedElement), array.Elements[i])
		if err != nil {
			return err
		}
	}
	return nil
}
