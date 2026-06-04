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

type constant any

func TestIntegerArithmetic(t *testing.T) {
	expect(t, "1 + 2", []constant{1, 2}, PUSH(0), PUSH(1), ADD, POP)
	expect(t, "1 - 2", []constant{1, 2}, PUSH(0), PUSH(1), SUB, POP)
	expect(t, "1 * 2", []constant{1, 2}, PUSH(0), PUSH(1), MUL, POP)
	expect(t, "2 / 1", []constant{2, 1}, PUSH(0), PUSH(1), DIV, POP)
	expect(t, "1; 2", []constant{1, 2}, PUSH(0), POP, PUSH(1), POP)
	expect(t, "-1", []constant{1}, PUSH(0), NEG, POP)
}

func TestBooleanExpressions(t *testing.T) {
	expect(t, "false", []constant{}, FALSE, POP)
	expect(t, "true", []constant{}, TRUE, POP)
	expect(t, "1 < 2", []constant{1, 2}, PUSH(0), PUSH(1), LT, POP)
	expect(t, "1 > 2", []constant{2, 1}, PUSH(0), PUSH(1), LT, POP)
	expect(t, "1 == 2", []constant{1, 2}, PUSH(0), PUSH(1), EQ, POP)
	expect(t, "1 != 2", []constant{1, 2}, PUSH(0), PUSH(1), NEQ, POP)
	expect(t, "true == false", []constant{}, TRUE, FALSE, EQ, POP)
	expect(t, "true != false", []constant{}, TRUE, FALSE, NEQ, POP)
	expect(t, "!true", []constant{}, TRUE, NOT, POP)
}

func expect(t *testing.T, input string, constants []constant, instructions ...code.Instructions) {
	t.Helper()

	program := parse(input)

	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	err = testInstructions(slices.Concat(instructions...), bytecode.Instructions)
	if err != nil {
		t.Fatalf("%q: %s", input, err)
	}

	err = testConstants(constants, bytecode.Constants)
	if err != nil {
		t.Fatalf("%q: %s", input, err)
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

func testConstants(expected []constant, actual []object.Object) error {
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

func PUSH(index int) []byte {
	return code.Make(code.OpConstant, index)
}

var POP []byte = code.Make(code.OpPop)

var FALSE []byte = code.Make(code.OpFalse)
var TRUE []byte = code.Make(code.OpTrue)

var EQ []byte = code.Make(code.OpEqual)
var NEQ []byte = code.Make(code.OpNotEqual)
var LT []byte = code.Make(code.OpLessThan)

var ADD []byte = code.Make(code.OpAdd)
var SUB []byte = code.Make(code.OpSub)
var MUL []byte = code.Make(code.OpMul)
var DIV []byte = code.Make(code.OpDiv)

var NEG []byte = code.Make(code.OpMinus)
var NOT []byte = code.Make(code.OpBang)
