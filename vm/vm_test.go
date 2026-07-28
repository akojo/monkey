package vm

import (
	"errors"
	"strings"
	"testing"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/compiler"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/object"
	"github.com/akojo/monkey/parser"
	"github.com/akojo/monkey/testutil"
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

func TestReturnStatement(t *testing.T) {
	expect(t, "let f = fn() { return 10 }; f()", 10)
	expect(t, "let f = fn() { return 10; 9; }; f()", 10)
	expect(t, "let f = fn() { return 2 * 5; 9 }; f()", 10)
	expect(t, "let f = fn() { 9; return 2 * 5; 9 }; f()", 10)
	expect(t, `
		let f = fn() {
			if (true) {
				if (true) {
					return 10;
				}
				return 1;
			}
		};
		f()`, 10)
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

func TestGlobalLetStatements(t *testing.T) {
	expect(t, "let one = 1; 1", 1)
	expect(t, "let one = 1; let two = 2; one + two;", 3)
	expect(t, "let one = 1; let two = one + one; one + two", 3)

	expect(t, "let a = 5; a;", 5)
	expect(t, "let a = 5 * 5; a", 25)
	expect(t, "let a = 5; let b = a; b", 5)
	expect(t, "let a = 5; let b = 5; let c = a + b + 5; c;", 15)
}

func TestFunctionCall(t *testing.T) {
	expect(t, "let f = fn() { 5 + 10; }; f();", 15)
	expect(t, "let one = fn() { 1 }; let two = fn() { 2 }; one() + two();", 3)
	expect(t, "let a = fn() { 1 }; let b = fn() { a() + 1 }; b();", 2)

	expect(t, "fn() { return 99; 100 }()", 99)
	expect(t, "fn() { return 99; return 100; }()", 99)

	expect(t, "let f = fn() {}; f()", nil)
	expect(t, "let f = fn() {}; let f2 = fn() { f() }; f2()", nil)

	expect(t, "let f1 = fn() { 1 }; let f2 = fn() { f1 }; f2()()", 1)

	expect(t, "let f = fn(a) { a }; f(4);", 4)
	expect(t, "let f = fn(a, b) { a + b }; f(24, 25);", 49)
	expect(t, "let f = fn(a, b) { let c = a + b; c }; f(1, 2);", 3)
	expect(t, "let f = fn(a, b) { let c = a + b; c }; f(1, 2) + f(3, 4);", 10)
	expect(t, "let f = fn(a) { let b = 1; a + b}; f(2)", 3)
	expect(t, `
		let global = 10;

		let sum = fn(a, b) {
			let c = a + b;
			c + global;
		}

		let outer = fn() {
			sum(1, 2) + sum(3, 4) + global;
		}

		outer() + global;
	`, 50)

	expect(t, "fn() { 1; }(1)", errors.New("wrong number of arguments: want 0, got 1"))
	expect(t, "fn(a) { a; }()", errors.New("wrong number of arguments: want 1, got 0"))
	expect(t, "fn(a, b) { a + b; }(1)", errors.New("wrong number of arguments: want 2, got 1"))
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

	expect(t, `
		let start = 0;
		let end = 2;
		[1, 2, 3, 4][start:end] == [1, 2]
	`, true)

	expect(t, `[]["start":0]`, errors.New("cannot use as slice start: STRING"))
	expect(t, `[][0:"end"]`, errors.New("cannot use as slice end: STRING"))
}

func TestArrayConcatenation(t *testing.T) {
	expect(t, "[1, 2] + [3, 4] == [1, 2, 3, 4]", true)
	expect(t, "[1, 2] + [] == [1, 2]", true)
	expect(t, "[] + [] == []", true)
}

func TestHashLiterals(t *testing.T) {
	expect(t, "{}", map[object.HashKey]int64{})
	expect(t, "{1: 2, 2: 3}", map[object.HashKey]int64{
		(&object.Integer{Value: 1}).Hash(): 2,
		(&object.Integer{Value: 2}).Hash(): 3,
	})
	expect(t, "{1 + 1: 2 * 2, 3 + 3: 4 * 4}", map[object.HashKey]int64{
		(&object.Integer{Value: 2}).Hash(): 4,
		(&object.Integer{Value: 6}).Hash(): 16,
	})
}

func TestHashIndexExpressions(t *testing.T) {
	expect(t, `{"foo": 5}["foo"]`, 5)
	expect(t, `{"foo": 5}["bar"]`, nil)
	expect(t, `{}["foo"]`, nil)
	expect(t, `{5: 10}[5]`, 10)
	expect(t, `{true: 5}[true]`, 5)
	expect(t, `{false: 5}[false]`, 5)
}

func TestVariableScoping(t *testing.T) {
	expect(t, "fn() { let one = 1; one }();", 1)
	expect(t, "fn() { let one = 1; let two = 2; one + two }();", 3)
	expect(t, `
		let first = fn() { let foo = 50; foo }
		let second = fn() { let foo = 100; foo }
		first() + second()
	`, 150)
	expect(t, `
		let seed = 50
		let minusOne = fn() { let num = 1; seed - num }
		let minusTwo = fn() { let num = 2; seed - num }
		minusOne() - minusTwo()
	`, 1)
	expect(t, `
		let f = fn() { let one = fn() { return 1 }; one }
		f()()
	`, 1)
}

func TestClosures(t *testing.T) {
	expect(t, `
		let closure = fn(a) {
			fn() { a; };
		};
		let f = closure(99);
		f();
	`, 99)

	expect(t, `
		let adder = fn(a, b) {
			fn(c) { a + b + c }
		}
		adder(1, 2)(8)
	`, 11)

	expect(t, `
		let adder = fn(a, b) {
			let c = a + b
			fn(d) { c + d }
		}
		adder(1, 2)(8)
	`, 11)

	expect(t, `
		let doubleAdder = fn(a, b) {
			let c = a + b
			fn(d) {
				let e = d + c
				fn (f) { e + f }
			}
		}
		doubleAdder(1, 2)(3)(8)
	`, 14)

	expect(t, `
		let closure = fn(a, b) {
			let one = fn() { a }
			let two = fn() { b }
			fn() { one() + two() }
		}
		closure(9, 90)()
	`, 99)
}

func TestRecursiveFunctions(t *testing.T) {
	expect(t, `
		let countDown = fn(x) {
			if (x == 0) {
				return 0
			} else {
				countDown(x - 1)
			}
		}
		countDown(10)
	`, 0)

	expect(t, `
		let f = fn() {
			let countDown = fn(x) {
				if (x == 0) {
					return 0
				} else {
					countDown(x - 1)
				}
			}
			countDown(10)
		}
		f()
	`, 0)
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
		switch e := expected.(type) {
		case error:
			if e.Error() != err.Error() {
				t.Errorf("vm error: expected %q got %q", e.Error(), err.Error())
			}
			return
		}
		t.Fatalf("vm error: %s", err)
	}

	if err := testutil.ExpectObject(vm.StackAboveTop(), expected); err != nil {
		t.Errorf("%q: %s", input, err)
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(strings.NewReader(input), "test")
	p := parser.New(l)
	return p.ParseProgram()
}
