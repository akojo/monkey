package compiler

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/code"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/object"
	"github.com/akojo/monkey/parser"
)

type compilerTestCase struct {
	input              string
	expectConstants    []any
	expectInstructions code.Instructions
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:           "1 + 2",
			expectConstants: []any{1, 2},
			expectInstructions: slices.Concat(
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			),
		},
		{
			input:           "1 - 2",
			expectConstants: []any{1, 2},
			expectInstructions: slices.Concat(
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			),
		},
		{
			input:           "1 * 2",
			expectConstants: []any{1, 2},
			expectInstructions: slices.Concat(
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			),
		},
		{
			input:           "2 / 1",
			expectConstants: []any{2, 1},
			expectInstructions: slices.Concat(
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			),
		},
		{
			input:           "1; 2",
			expectConstants: []any{1, 2},
			expectInstructions: slices.Concat(
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			),
		},
	}

	for _, test := range tests {
		runCompilerTest(t, test)
	}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:           "false",
			expectConstants: []any{},
			expectInstructions: slices.Concat(
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			),
		},
		{
			input:           "true",
			expectConstants: []any{},
			expectInstructions: slices.Concat(
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			),
		},
	}

	for _, test := range tests {
		runCompilerTest(t, test)
	}
}

func runCompilerTest(t *testing.T, test compilerTestCase) {
	t.Helper()

	program := parse(test.input)

	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	err = testInstructions(test.expectInstructions, bytecode.Instructions)
	if err != nil {
		t.Fatalf("%q: %s", test.input, err)
	}

	err = testConstants(test.expectConstants, bytecode.Constants)
	if err != nil {
		t.Fatalf("%q: %s", test.input, err)
	}
}

func testInstructions(expected code.Instructions, actual code.Instructions) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("instructions:\n  want %d  %q\n  got  %d  %q", len(expected), expected, len(actual), actual)
	}

	for i, ins := range expected {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d:\n  want %q\n  got  %q", i, ins, actual[i])
		}
	}

	return nil
}

func testConstants(expected []any, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("constants:\n  want %d  %q\n  got  %d  %q", len(expected), expected, len(actual), actual)
	}

	for i, c := range expected {
		var err error
		switch c := c.(type) {
		case int:
			err = testIntegerObject(int64(c), actual[i])
		}
		if err != nil {
			return fmt.Errorf("constant %d: %s", i, err)
		}
	}

	return nil
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

func parse(input string) *ast.Program {
	l := lexer.New(strings.NewReader(input), "test")
	p := parser.New(l)
	return p.ParseProgram()
}
