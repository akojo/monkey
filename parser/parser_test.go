package parser

import (
	"strconv"
	"strings"
	"testing"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/lexer"
)

func TestLetStatements(t *testing.T) {
	test := func(input string, ident string, value any) {
		t.Run(input, func(t *testing.T) {
			program := makeProgram(t, input)

			expectStatementCount(t, program.Statements, 1)

			stmt := program.Statements[0]

			if stmt.TokenLiteral() != "let" {
				t.Errorf("s.TokenLiteral: expected \"let\", got %q", stmt.TokenLiteral())
				return
			}

			letStmt, ok := stmt.(*ast.LetStatement)
			if !ok {
				t.Errorf("s: expected *ast.LetStatement, got %T", stmt)
				return
			}

			expectIdentifier(t, letStmt.Name, ident)
			expectLiteralExpression(t, letStmt.Value, value)
		})
	}

	test("let x = 5;", "x", 5)
	test("let y = true;", "y", true)
	test("let foobar = y;", "foobar", "y")
}

func TestReturnStatements(t *testing.T) {
	test := func(input string, value any) {
		t.Run(input, func(t *testing.T) {
			program := makeProgram(t, input)

			expectStatementCount(t, program.Statements, 1)

			stmt := program.Statements[0]

			returnStmt, ok := stmt.(*ast.ReturnStatement)
			if !ok {
				t.Errorf("stmt: expected *ast.ReturnStatement, got %T", stmt)
			}
			if returnStmt.TokenLiteral() != "return" {
				t.Errorf("returnStmt.TokenLiteral: expected \"return\", got %q", returnStmt.TokenLiteral())
			}

			expectLiteralExpression(t, returnStmt.ReturnValue, value)
		})
	}

	test("return 5;", 5)
	test("return true;", true)
	test("return x", "x")
}

func TestIdentifierExpression(t *testing.T) {
	program := makeProgram(t, "foobar;")

	expectStatementCount(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt: expected *ast.ExpressionStatement, got %T", stmt)
	}

	expectIdentifier(t, stmt.Expression, "foobar")
}

func TestIntegerLiteralExpression(t *testing.T) {
	program := makeProgram(t, "5;")

	expectStatementCount(t, program.Statements, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])

	expectIntegerLiteral(t, stmt.Expression, 5)
}

func TestBooleanExpression(t *testing.T) {
	program := makeProgram(t, "true;")

	expectStatementCount(t, program.Statements, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])

	expectLiteralExpression(t, stmt.Expression, true)
}

func TestStringLiteralExpression(t *testing.T) {
	program := makeProgram(t, `"hello, world";`)

	expectStatementCount(t, program.Statements, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("literal: expected *ast.StringLiteral, got %T", stmt.Expression)
	}

	const expected = "hello, world"
	if literal.Value != expected {
		t.Errorf("literal.Value: expected %q, got %q", expected, literal.Value)
	}
}

func TestPrefixExpressions(t *testing.T) {
	test := func(input string, op string, value any) {
		program := makeProgram(t, input)

		expectStatementCount(t, program.Statements, 1)

		stmt := expectExpressionStatement(t, program.Statements[0])

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("exp: expected *ast.PrefixExpression, got %T", exp)
		}
		if exp.Operator != op {
			t.Fatalf("exp.Operator: expected %q, got %q", op, exp.Operator)
		}
		expectLiteralExpression(t, exp.Right, value)
	}

	test("!5;", "!", 5)
	test("-15;", "-", 15)
	test("!true;", "!", true)
	test("!false;", "!", false)
}

func TestInfixExpressions(t *testing.T) {
	test := func(input string, leftVal any, op string, rightVal any) {
		program := makeProgram(t, input)

		expectStatementCount(t, program.Statements, 1)

		stmt := expectExpressionStatement(t, program.Statements[0])

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("exp, expected *ast.InfixExpression, got %T", exp)
		}

		expectLiteralExpression(t, exp.Left, leftVal)

		if exp.Operator != op {
			t.Fatalf("exp.Operator: expected %q, got %q", op, exp.Operator)
		}

		expectLiteralExpression(t, exp.Right, rightVal)
	}

	test("5 + 5;", 5, "+", 5)
	test("5 - 5;", 5, "-", 5)
	test("5 * 5;", 5, "*", 5)
	test("5 / 5;", 5, "/", 5)
	test("5 < 5;", 5, "<", 5)
	test("5 > 5;", 5, ">", 5)
	test("5 == 5;", 5, "==", 5)
	test("5 != 5;", 5, "!=", 5)
	test("true == true", true, "==", true)
	test("false != false", false, "!=", false)
}

func TestOperatorPrecedences(t *testing.T) {
	test := func(input string, expected string) {
		program := makeProgram(t, input)

		got := program.String()
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	}

	test("-a * b", "((-a) * b);\n")
	test("!-a", "(!(-a));\n")
	test("a + b + c", "((a + b) + c);\n")
	test("a + b - c", "((a + b) - c);\n")
	test("a * b * c", "((a * b) * c);\n")
	test("a * b / c", "((a * b) / c);\n")
	test("a + b / c", "(a + (b / c));\n")
	test("a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f);\n")
	test("3 + 4; -5 * 5", "(3 + 4);\n((-5) * 5);\n")
	test("5 > 4 == 3 < 4", "((5 > 4) == (3 < 4));\n")
	test("5 < 4 != 3 > 4", "((5 < 4) != (3 > 4));\n")
	test("3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)));\n")
	test("true", "true;\n")
	test("false", "false;\n")
	test("3 > 5 == false", "((3 > 5) == false);\n")
	test("3 < 5 == true", "((3 < 5) == true);\n")
	test("1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4);\n")
	test("(5 + 5) * 2", "((5 + 5) * 2);\n")
	test("2 / (5 + 5)", "(2 / (5 + 5));\n")
	test("-(5 + 5)", "(-(5 + 5));\n")
	test("!(true == true)", "(!(true == true));\n")
	test("a + add(b * c) + d", "((a + add((b * c))) + d);\n")
	test("add(a, 1, 2 * 3, add(6, 7 *8))", "add(a, 1, (2 * 3), add(6, (7 * 8)));\n")
	test("add(a + b * c + d)", "add(((a + (b * c)) + d));\n")
	test("a * [1, 2, 3][b * c]", "(a * ([1, 2, 3][(b * c)]));\n")
	test("add(a * -b[2])", "add((a * (-(b[2]))));\n")
}

func TestIFExpression(t *testing.T) {
	program := makeProgram(t, "if (x < y) { x }")

	expectStatementCount(t, program.Statements, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])

	exp, ok := stmt.Expression.(*ast.IFExpression)
	if !ok {
		t.Fatalf("stmt.Expression: expected *ast.IFExpression, got %T", exp)
	}

	expectInfixExpression(t, exp.Condition, "x", "<", "y")

	expectStatementCount(t, exp.Consequence.Statements, 1)
	consequence := expectExpressionStatement(t, exp.Consequence.Statements[0])
	expectIdentifier(t, consequence.Expression, "x")

	if exp.Alternative != nil {
		t.Errorf("exp.Alternative: expected nil, got %+v", exp.Alternative)
	}
}

func TestIFElseExpression(t *testing.T) {
	program := makeProgram(t, "if (x < y) { x } else { y }")

	expectStatementCount(t, program.Statements, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])

	exp, ok := stmt.Expression.(*ast.IFExpression)
	if !ok {
		t.Fatalf("stmt.Expression: expected *ast.IFExpression, got %T", exp)
	}

	expectInfixExpression(t, exp.Condition, "x", "<", "y")

	expectStatementCount(t, exp.Consequence.Statements, 1)
	consequence := expectExpressionStatement(t, exp.Consequence.Statements[0])
	expectIdentifier(t, consequence.Expression, "x")

	expectStatementCount(t, exp.Alternative.Statements, 1)
	alternative := expectExpressionStatement(t, exp.Alternative.Statements[0])
	expectIdentifier(t, alternative.Expression, "y")
}

func TestFunctionLiteral(t *testing.T) {
	program := makeProgram(t, "fn(x, y) { x + y; }")

	expectStatementCount(t, program.Statements, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression: expected *ast.FunctionLiteral, got %T", stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Errorf("function.Parameters: expected 2, got %d", len(function.Parameters))
	}
	expectLiteralExpression(t, function.Parameters[0], "x")
	expectLiteralExpression(t, function.Parameters[1], "y")

	expectStatementCount(t, function.Body.Statements, 1)
	bodyStmt := expectExpressionStatement(t, function.Body.Statements[0])
	expectInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestCallExpression(t *testing.T) {
	program := makeProgram(t, "add(1, 2 * 3);")

	expectStatementCount(t, program.Statements, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression: expected *ast.CallExpression, got %T", stmt.Expression)
	}

	expectIdentifier(t, exp.Function, "add")

	if len(exp.Arguments) != 2 {
		t.Fatalf("exp.Arguments: expected 2, got %d", len(exp.Arguments))
	}

	expectLiteralExpression(t, exp.Arguments[0], 1)
	expectInfixExpression(t, exp.Arguments[1], 2, "*", 3)
}

func TestArrayLiteral(t *testing.T) {
	program := makeProgram(t, "[1, 2 * 2, 3 + 3]")

	expectStatementCount(t, program.Statements, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])

	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("stmt.Expression: expected *ast.ArrayLiteral, got %T", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("array.Elements: expected 3, got %d", len(array.Elements))
	}

	expectIntegerLiteral(t, array.Elements[0], 1)
	expectInfixExpression(t, array.Elements[1], 2, "*", 2)
	expectInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestEmptyArrayLiteral(t *testing.T) {
	program := makeProgram(t, "[]")

	expectStatementCount(t, program.Statements, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("stmt.Expression: expected *ast.ArrayLiteral, got %T", stmt.Expression)
	}

	if len(array.Elements) != 0 {
		t.Fatalf("array.Elements: expected 0, got %d", len(array.Elements))
	}
}

func TestIndexExpressions(t *testing.T) {
	program := makeProgram(t, "myArray[1 + 1]")

	expectStatementCount(t, program.Statements, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])
	index, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("index: expected *ast.IndexExpression. got %T", stmt.Expression)
	}

	expectIdentifier(t, index.Left, "myArray")
	expectInfixExpression(t, index.Index, 1, "+", 1)
}

func TestSliceExpressions(t *testing.T) {
	test := func(input string, start *int64, end *int64) {
		program := makeProgram(t, input)

		expectStatementCount(t, program.Statements, 1)

		stmt := expectExpressionStatement(t, program.Statements[0])
		index, ok := stmt.Expression.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("slice: expected *ast.IndexExpression, got %T", stmt.Expression)
		}

		slice, ok := index.Index.(*ast.SliceExpression)
		if !ok {
			t.Fatalf("slice: expected *ast.SliceExpression, got %T", stmt.Expression)
		}

		if start != nil {
			expectIntegerLiteral(t, slice.Start, *start)
		} else if slice.Start != nil {
			t.Errorf("slice.Start: expect nil, got %+v", slice.Start)
		}

		if end != nil {
			expectIntegerLiteral(t, slice.End, *end)
		} else if slice.End != nil {
			t.Errorf("slice.End: expect nil, got %+v", slice.End)
		}
	}

	start := int64(1)
	end := int64(2)

	test("myarray[1:2]", &start, &end)
	test("[1, 2][1:2]", &start, &end)
	test("myarray[1:]", &start, nil)
	test("myarray[:2]", nil, &end)
	test("myarray[:]", nil, nil)
}

func makeProgram(t *testing.T, input string) *ast.Program {
	l := lexer.New(strings.NewReader(input), "<test>")
	p := New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	checkParserErrors(t, p)

	return program
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	for _, err := range errors {
		t.Errorf("%s", err)
	}
	t.FailNow()
}

func expectStatementCount(t *testing.T, stmts []ast.Statement, count int) {
	if len(stmts) != count {
		t.Fatalf("expected %d statements, got %d", count, len(stmts))
	}
}

func expectLiteralExpression(t *testing.T, exp ast.Expression, expected any) {
	switch v := expected.(type) {
	case int:
		expectIntegerLiteral(t, exp, int64(v))
	case int64:
		expectIntegerLiteral(t, exp, v)
	case string:
		expectIdentifier(t, exp, v)
	case bool:
		expectBoolean(t, exp, v)
	default:
		t.Errorf("exp: invalid type, got %T", v)
	}
}

func expectIdentifier(t *testing.T, exp ast.Expression, value string) {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Fatalf("ident: expected *ast.Identifier, got %T", ident)
	}
	if ident.Value != value {
		t.Errorf("ident.Value: expected %q, got %q", value, ident.Value)
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral: expected %q, got %q", value, ident.TokenLiteral())
	}
}

func expectIntegerLiteral(t *testing.T, exp ast.Expression, value int64) {
	intLiteral, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("int: expected *ast.IntegerLiteral, got %T", exp)
	}
	if intLiteral.Value != value {
		t.Errorf("int.Value: expected %d, got %d", value, intLiteral.Value)
	}
	tokenLiteral := strconv.FormatInt(value, 10)
	if intLiteral.TokenLiteral() != tokenLiteral {
		t.Errorf("int.TokenLiteral: expected %q, got %q", tokenLiteral, intLiteral.TokenLiteral())
	}
}

func expectBoolean(t *testing.T, exp ast.Expression, value bool) {
	boolean, ok := exp.(*ast.Boolean)
	if !ok {
		t.Fatalf("boolean: expected *ast.Boolean, got %T", boolean)
	}
	if boolean.Value != value {
		t.Errorf("boolean.Value: expected %t, got %t", value, boolean.Value)
	}
	tokenLiteral := strconv.FormatBool(value)
	if boolean.TokenLiteral() != tokenLiteral {
		t.Errorf("boolean.TokenLiteral: expected %q, got %q", tokenLiteral, boolean.TokenLiteral())
	}

}

func expectInfixExpression(t *testing.T, exp ast.Expression, left any, op string, right any) {
	infix, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("exp: expected *ast.InfixExpression, got %T", exp)
	}

	expectLiteralExpression(t, infix.Left, left)

	if infix.Operator != op {
		t.Errorf("infix.Operator: expected %q, got %q", op, infix.Operator)
	}

	expectLiteralExpression(t, infix.Right, right)
}

func expectExpressionStatement(t *testing.T, stmt ast.Statement) *ast.ExpressionStatement {
	expressionStmt, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt: expected *ast.ExpressionStatement, got %T", stmt)
	}
	return expressionStmt
}
