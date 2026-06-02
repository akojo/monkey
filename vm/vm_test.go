package vm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/compiler"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/object"
	"github.com/akojo/monkey/parser"
)

func TestIntegerArithmetic(t *testing.T) {
	runVmTest(t, "1", 1)
	runVmTest(t, "2", 2)
	runVmTest(t, "1 + 2", 3)
}

func runVmTest(t *testing.T, input string, expected any) {
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

	if err := testExpectedObject(expected, vm.StackAboveTop()); err != nil {
		t.Errorf("%q: %s", input, err)
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(strings.NewReader(input), "test")
	p := parser.New(l)
	return p.ParseProgram()
}

func testExpectedObject(expected any, actual object.Object) error {
	var err error
	switch expected := expected.(type) {
	case int:
		err = testIntegerObject(int64(expected), actual)
	}
	return err
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("want Integer, got %T (%+v)", actual, actual)
	}

	if expected != result.Value {
		return fmt.Errorf("want %d, got %d", expected, result.Value)
	}

	return nil
}
