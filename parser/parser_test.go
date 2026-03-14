package parser

import (
	"strconv"
	"strings"
	"testing"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/lexer"
)

func TestLetStatements(t *testing.T) {
	test := func(t *testing.T, s ast.Statement, name string) {
		if s.TokenLiteral() != "let" {
			t.Errorf("s.TokenLiteral: expected \"let\", got %q", s.TokenLiteral())
			return
		}

		letStmt, ok := s.(*ast.LetStatement)
		if !ok {
			t.Errorf("s: expected *ast.LetStatement, got %T", s)
			return
		}

		expectIdentifier(t, letStmt.Name, name)
	}

	program := makeProgram(t, `
		let x = 5;
		let y = 10;`)

	expectStatementCount(t, program.Statements, 2)

	test(t, program.Statements[0], "x")
	test(t, program.Statements[1], "y")
}

func TestReturnStatements(t *testing.T) {
	program := makeProgram(t, `
		return 5;
		return 10;`)

	expectStatementCount(t, program.Statements, 2)

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt: expected *ast.ReturnStatement, got %T", stmt)
			continue
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral: expected \"return\", got %q", returnStmt.TokenLiteral())
		}
	}
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

	test("-a * b", "((-a) * b)")
	test("!-a", "(!(-a))")
	test("a + b + c", "((a + b) + c)")
	test("a + b - c", "((a + b) - c)")
	test("a * b * c", "((a * b) * c)")
	test("a * b / c", "((a * b) / c)")
	test("a + b / c", "(a + (b / c))")
	test("a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)")
	test("3 + 4; -5 * 5", "(3 + 4)\n((-5) * 5)")
	test("5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))")
	test("5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))")
	test("3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))")
	test("true", "true")
	test("false", "false")
	test("3 > 5 == false", "((3 > 5) == false)")
	test("3 < 5 == true", "((3 < 5) == true)")
	test("1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)")
	test("(5 + 5) * 2", "((5 + 5) * 2)")
	test("2 / (5 + 5)", "(2 / (5 + 5))")
	test("-(5 + 5)", "(-(5 + 5))")
	test("!(true == true)", "(!(true == true))")
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
	literal, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("literal: expected *ast.IntegerLiteral, got %T", literal)
	}
	if literal.Value != value {
		t.Errorf("literal.Value: expected %d, got %d", value, literal.Value)
	}
	tokenLiteral := strconv.FormatInt(value, 10)
	if literal.TokenLiteral() != tokenLiteral {
		t.Errorf("literal.TokenLiteral: expected %q, got %q", tokenLiteral, literal.TokenLiteral())
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
