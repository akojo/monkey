package compiler

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/builtin"
	"github.com/akojo/monkey/code"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/object"
	"github.com/akojo/monkey/parser"
	"github.com/akojo/monkey/testutil"
)

type constant any

func TestCompilerScopes(t *testing.T) {
	compiler := New()
	if compiler.currentScope != 0 {
		t.Errorf("scopeIndex: want 0, got %d", compiler.currentScope)
	}
	globalSymbolTable := compiler.symbolTable

	compiler.emit(code.OpMul)

	compiler.enterScope()
	if compiler.currentScope != 1 {
		t.Errorf("scopeIndex: want 1, got %d", compiler.currentScope)
	}

	compiler.emit(code.OpSub)

	if len(compiler.scopes[compiler.currentScope].instructions) != 1 {
		t.Errorf("len(instructions): want 1, got %d", len(compiler.scopes[compiler.currentScope].instructions))
	}

	current := compiler.scopes[compiler.currentScope].currentInstruction
	if current.Opcode != code.OpSub {
		t.Errorf("currentInstruction.Opcode: want %d, got %d", code.OpSub, current.Opcode)
	}

	if compiler.symbolTable.Outer != globalSymbolTable {
		t.Error("global symbol table is not enclosed")
	}

	compiler.leaveScope()
	if compiler.currentScope != 0 {
		t.Errorf("scopeIndex: want 0, got %d", compiler.currentScope)
	}

	if compiler.symbolTable != globalSymbolTable {
		t.Error("expected global symbol table")
	}
	if compiler.symbolTable.Outer != nil {
		t.Error("global symbol table should not have outer symbol table")
	}

	compiler.emit(code.OpAdd)

	if len(compiler.scopes[compiler.currentScope].instructions) != 2 {
		t.Errorf("len(instructions): want 2, got %d", len(compiler.scopes[compiler.currentScope].instructions))
	}

	current = compiler.scopes[compiler.currentScope].currentInstruction
	if current.Opcode != code.OpAdd {
		t.Errorf("currentInstruction.Opcode: want %d, got %d", code.OpAdd, current.Opcode)
	}

	previous := compiler.scopes[compiler.currentScope].previousInstruction
	if previous.Opcode != code.OpMul {
		t.Errorf("currentInstruction.Opcode: want %d, got %d", code.OpMul, previous.Opcode)
	}
}

func TestIntegerArithmetic(t *testing.T) {
	expect(t, "1 + 2", []constant{}, PUSHI(1), PUSHI(2), ADD, POP)
	expect(t, "1 - 2", []constant{}, PUSHI(1), PUSHI(2), SUB, POP)
	expect(t, "1 * 2", []constant{}, PUSHI(1), PUSHI(2), MUL, POP)
	expect(t, "2 / 1", []constant{}, PUSHI(2), PUSHI(1), DIV, POP)
	expect(t, "1; 2", []constant{}, PUSHI(1), POP, PUSHI(2), POP)
	expect(t, "-1", []constant{}, PUSHI(1), NEG, POP)
}

func TestLargeIntegerConstants(t *testing.T) {
	expect(t, "1 + 100000", []constant{100000}, PUSHI(1), PUSH(0), ADD, POP)
	expect(t, "100000 + 1", []constant{100000}, PUSH(0), PUSHI(1), ADD, POP)
	expect(t, "100000 + 200000", []constant{100000, 200000}, PUSH(0), PUSH(1), ADD, POP)
}

func TestBooleanExpressions(t *testing.T) {
	expect(t, "false", []constant{}, FALSE, POP)
	expect(t, "true", []constant{}, TRUE, POP)
	expect(t, "1 < 2", []constant{}, PUSHI(1), PUSHI(2), LT, POP)
	expect(t, "1 > 2", []constant{}, PUSHI(2), PUSHI(1), LT, POP)
	expect(t, "1 == 2", []constant{}, PUSHI(1), PUSHI(2), EQ, POP)
	expect(t, "1 != 2", []constant{}, PUSHI(1), PUSHI(2), NEQ, POP)
	expect(t, "true == false", []constant{}, TRUE, FALSE, EQ, POP)
	expect(t, "true != false", []constant{}, TRUE, FALSE, NEQ, POP)
	expect(t, "!true", []constant{}, TRUE, NOT, POP)
}

func TestConditionals(t *testing.T) {
	expect(t, "if (true) { 10 }; 3333;", []constant{},
		TRUE, BNE(8), PUSHI(10), B(3), NULL, POP, PUSHI(3333), POP)
	//  0000  0001    0004       0007  0010  0011 0012         0015

	expect(t, "if (true) { 10 } else { 20 }; 3333;", []constant{},
		TRUE, BNE(8), PUSHI(10), B(5), PUSHI(20), POP, PUSHI(3333), POP)
	//  0000  0001     0004      0007  0010       0013 0014         0017

	expect(t, "if (true) { } else { }; 3333;", []constant{},
		TRUE, BNE(6), NULL, B(3), NULL, POP, PUSHI(3333), POP)
	//  0000  0001    0004  0005  0008  0009 0010         0013

	expect(t, "if (true) { }; 3333;", []constant{},
		TRUE, BNE(6), NULL, B(3), NULL, POP, PUSHI(3333), POP)
	//  0000  0001    0004  0005  0008  0009 0010         0013
}

func TestGlobalStatements(t *testing.T) {
	expect(t, "let one = 1; let two = 2;", []constant{}, PUSHI(1), SETG(0), PUSHI(2), SETG(1))
	expect(t, "let one = 1; one;", []constant{}, PUSHI(1), SETG(0), GETG(0), POP)
	expect(t, "let one = 1; let two = one; two;", []constant{}, PUSHI(1), SETG(0), GETG(0), SETG(1), GETG(1), POP)
}

func TestStringExpressions(t *testing.T) {
	expect(t, `"monkey"`, []constant{"monkey"}, PUSH(0), POP)
	expect(t, `"mon" + "key"`, []constant{"mon", "key"}, PUSH(0), PUSH(1), ADD, POP)
}

func TestArrayLiterals(t *testing.T) {
	expect(t, "[]", []constant{}, ARRAY(0), POP)
	expect(t, "[1, 2, 3]", []constant{}, PUSHI(1), PUSHI(2), PUSHI(3), ARRAY(3), POP)
	expect(t, "[1 + 2, 3 - 4, 5 * 6]", []constant{},
		PUSHI(1), PUSHI(2), ADD, PUSHI(3), PUSHI(4), SUB, PUSHI(5), PUSHI(6), MUL, ARRAY(3), POP)
}

func TestHashLiterals(t *testing.T) {
	expect(t, "{}", []constant{}, HASH(0), POP)
	expect(t, "{1: 2, 3: 4}", []constant{},
		PUSHI(1), PUSHI(2), PUSHI(3), PUSHI(4), HASH(2), POP)
	expect(t, "{1: 2 + 3, 4: 5 * 6}", []constant{},
		PUSHI(1), PUSHI(2), PUSHI(3), ADD, PUSHI(4), PUSHI(5), PUSHI(6), MUL, HASH(2), POP)
}

func TestIndexExpressions(t *testing.T) {
	expect(t, "[2, 3, 4][0 + 1]", []constant{},
		PUSHI(2), PUSHI(3), PUSHI(4), ARRAY(3), PUSHI(0), PUSHI(1), ADD, INDEX, POP)
	expect(t, "{1: 2}[1]", []constant{},
		PUSHI(1), PUSHI(2), HASH(1), PUSHI(1), INDEX, POP)
}

func TestSliceExpressions(t *testing.T) {
	expect(t, "[1, 2][0:2]", []constant{},
		PUSHI(1), PUSHI(2), ARRAY(2), PUSHI(0), PUSHI(2), SLICE, INDEX, POP)

	expect(t, "[1, 2][0:]", []constant{},
		PUSHI(1), PUSHI(2), ARRAY(2), PUSHI(0), NULL, SLICE, INDEX, POP)

	expect(t, "[1, 2][:2]", []constant{},
		PUSHI(1), PUSHI(2), ARRAY(2), PUSHI(0), PUSHI(2), SLICE, INDEX, POP)
}

func TestFunctions(t *testing.T) {
	expect(t, "fn() { return 5 + 10 }",
		[]constant{function(PUSHI(5), PUSHI(10), ADD, RET)},
		CLOSURE(0, 0), POP)

	expect(t, "fn() { 5 + 10 }",
		[]constant{function(PUSHI(5), PUSHI(10), ADD, RET)},
		CLOSURE(0, 0), POP)

	expect(t, "fn() { 1; 2 }",
		[]constant{function(PUSHI(1), POP, PUSHI(2), RET)},
		CLOSURE(0, 0), POP)

	expect(t, "fn() {}", []constant{function(NULL, RET)}, CLOSURE(0, 0), POP)
}

func TestFunctionCalls(t *testing.T) {
	expect(t, "fn() { 24 }();",
		[]constant{function(PUSHI(24), RET)},
		CLOSURE(0, 0), CALL(0), POP)

	expect(t, "let f = fn() { 24 }; f();",
		[]constant{function(PUSHI(24), RET)},
		CLOSURE(0, 0), SETG(0), GETG(0), CALL(0), POP)

	expect(t, "let f = fn(a) { }; f(24);",
		[]constant{function(NULL, RET)},
		CLOSURE(0, 0), SETG(0), GETG(0), PUSHI(24), CALL(1), POP)

	expect(t, "let f = fn(a, b, c) { }; f(24, 25, 26);",
		[]constant{function(NULL, RET)},
		CLOSURE(0, 0), SETG(0), GETG(0), PUSHI(24), PUSHI(25), PUSHI(26), CALL(3), POP)

	expect(t, "let f = fn(a) { a }; f(24);",
		[]constant{function(GET(0), RET)},
		CLOSURE(0, 0), SETG(0), GETG(0), PUSHI(24), CALL(1), POP)

	expect(t, "let f = fn(a, b) { a + b }; f(24, 25);",
		[]constant{function(GET(0), GET(1), ADD, RET)},
		CLOSURE(0, 0), SETG(0), GETG(0), PUSHI(24), PUSHI(25), CALL(2), POP)
}

func TestLetStatementScopes(t *testing.T) {
	expect(t, "let num = 55; fn() { num }",
		[]constant{function(GETG(0), RET)},
		PUSHI(55), SETG(0), CLOSURE(0, 0), POP)

	expect(t, "fn() { let num = 55; num }",
		[]constant{function(PUSHI(55), SET(0), GET(0), RET)},
		CLOSURE(0, 0), POP)

	expect(t, "fn() { let a = 55; let b = 77; a + b }",
		[]constant{function(PUSHI(55), SET(0), PUSHI(77), SET(1), GET(0), GET(1), ADD, RET)},
		CLOSURE(0, 0), POP)
}

func TestBuiltins(t *testing.T) {
	expect(t, "len([]); append([], 1);",
		[]constant{},
		BUILTIN("len"), ARRAY(0), CALL(1), POP, BUILTIN("append"), ARRAY(0), PUSHI(1), CALL(2), POP)

	expect(t, "fn() { len([]); }",
		[]constant{function(BUILTIN("len"), ARRAY(0), CALL(1), RET)},
		CLOSURE(0, 0), POP)
}

func TestClosures(t *testing.T) {
	expect(t, `
		fn(a) {
			fn(b) { a + b }
		}`,
		[]constant{
			function(GETF(0), GET(0), ADD, RET),
			function(GET(0), CLOSURE(0, 1), RET),
		},
		CLOSURE(1, 0), POP)

	expect(t, `
		fn(a) {
			fn(b) {
				fn(c) { a + b + c }
			}
		}`,
		[]constant{
			function(GETF(0), GETF(1), ADD, GET(0), ADD, RET),
			function(GETF(0), GET(0), CLOSURE(0, 2), RET),
			function(GET(0), CLOSURE(1, 1), RET),
		},
		CLOSURE(2, 0), POP)

	expect(t, `
		let global = 55;

		fn() {
			let a = 66;

			fn() {
				let b = 77;

				fn() {
					let c = 88
					global + a + b + c;
				}
			}
		}`,
		[]constant{
			function(PUSHI(88), SET(0), GETG(0), GETF(0), ADD, GETF(1), ADD, GET(0), ADD, RET),
			function(PUSHI(77), SET(0), GETF(0), GET(0), CLOSURE(0, 2), RET),
			function(PUSHI(66), SET(0), GET(0), CLOSURE(1, 1), RET),
		},
		PUSHI(55), SETG(0), CLOSURE(2, 0), POP)
}

func TestRecursiveFunctions(t *testing.T) {
	expect(t, `
		let countDown = fn(x) { countDown(x - 1) }
		countDown(1)
		`,
		[]constant{
			function(CURRENTCLOSURE, GET(0), PUSHI(1), SUB, CALL(1), RET),
		},
		CLOSURE(0, 0), SETG(0), GETG(0), PUSHI(1), CALL(1), POP)

	expect(t, `
		let wrapper = fn() {
			let countDown = fn(x) { countDown(x - 1) }
			countDown(2)
		}
		wrapper()
		`,
		[]constant{
			function(CURRENTCLOSURE, GET(0), PUSHI(1), SUB, CALL(1), RET),
			function(CLOSURE(0, 0), SET(0), GET(0), PUSHI(2), CALL(1), RET),
		},
		CLOSURE(1, 0), SETG(0), GETG(0), CALL(0), POP)
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
			return fmt.Errorf("wrong instruction at %d:\n  want %02x  %q\n  got  %02x  %q", i, ins, expected, actual[i], actual)
		}
	}

	return nil
}

func testConstants(expected []constant, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("constants:\n  want %d  %q\n  got  %d  %q", len(expected), expected, len(actual), actual)
	}

	for i, c := range expected {
		if err := testutil.ExpectObject(actual[i], c); err != nil {
			return fmt.Errorf("constant %d: %s", i, err)
		}
	}

	return nil
}

func parse(input string) *ast.Program {
	l := lexer.New(strings.NewReader(input), "test")
	p := parser.New(l)
	return p.ParseProgram()
}

func function(instructions ...code.Instructions) code.Instructions {
	return slices.Concat(instructions...)
}

func PUSH(index int) code.Instructions {
	return code.Make(code.OpConstant, index)
}

func PUSHI(value int) code.Instructions {
	return code.Make(code.OpPushImmediate, value)
}

var POP = code.Make(code.OpPop)

func GETG(index int) code.Instructions {
	return code.Make(code.OpGetGlobal, index)
}

func SETG(index int) code.Instructions {
	return code.Make(code.OpSetGlobal, index)
}

func GET(index int) code.Instructions {
	return code.Make(code.OpGetLocal, index)
}

func SET(index int) code.Instructions {
	return code.Make(code.OpSetLocal, index)
}

var NULL = code.Make(code.OpNull)

var FALSE = code.Make(code.OpFalse)
var TRUE = code.Make(code.OpTrue)

func ARRAY(size int) code.Instructions {
	return code.Make(code.OpArray, size)
}

func HASH(size int) code.Instructions {
	return code.Make(code.OpHash, size)
}

var INDEX = code.Make(code.OpIndex)
var SLICE = code.Make(code.OpSlice)

var EQ = code.Make(code.OpEqual)
var NEQ = code.Make(code.OpNotEqual)
var LT = code.Make(code.OpLessThan)

func B(offset int) code.Instructions {
	return code.Make(code.OpBranch, offset)
}

func BNE(offset int) code.Instructions {
	return code.Make(code.OpBranchNotEqual, offset)
}

var ADD = code.Make(code.OpAdd)
var SUB = code.Make(code.OpSub)
var MUL = code.Make(code.OpMul)
var DIV = code.Make(code.OpDiv)

var NEG = code.Make(code.OpMinus)
var NOT = code.Make(code.OpBang)

func CALL(nargs int) code.Instructions {
	return code.Make(code.OpCall, nargs)
}

var RET = code.Make(code.OpReturn)

func BUILTIN(name string) code.Instructions {
	for i, b := range builtin.Builtins {
		if b.Name == name {
			return code.Make(code.OpGetBuiltin, i)
		}
	}
	panic("invalid builtin name: " + name)
}

func CLOSURE(index, nfree int) code.Instructions {
	return code.Make(code.OpClosure, index, nfree)
}

func GETF(index int) code.Instructions {
	return code.Make(code.OpGetFree, index)
}

var CURRENTCLOSURE = code.Make(code.OpCurrentClosure)
