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

	expectStatementCount(t, program, 2)

	test(t, program.Statements[0], "x")
	test(t, program.Statements[1], "y")
}

func TestReturnStatements(t *testing.T) {
	program := makeProgram(t, `
		return 5;
		return 10;`)

	expectStatementCount(t, program, 2)

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

	expectStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt: expected *ast.ExpressionStatement, got %T", stmt)
	}

	expectIdentifier(t, stmt.Expression, "foobar")
}

func TestIntegerLiteralExpression(t *testing.T) {
	program := makeProgram(t, "5;")

	expectStatementCount(t, program, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])

	expectIntegerLiteral(t, stmt.Expression, 5)
}

func TestPrefixExpressions(t *testing.T) {
	test := func(input string, op string, value int64) {
		program := makeProgram(t, input)

		expectStatementCount(t, program, 1)

		stmt := expectExpressionStatement(t, program.Statements[0])

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("exp: expected *ast.PrefixExpression, got %T", exp)
		}
		if exp.Operator != op {
			t.Fatalf("exp.Operator: expected %q, got %q", op, exp.Operator)
		}
		expectIntegerLiteral(t, exp.Right, value)
	}

	test("!5;", "!", 5)
	test("-15;", "-", 15)
}

func TestInfixExpressions(t *testing.T) {
	test := func(input string, leftVal int64, op string, rightVal int64) {
		program := makeProgram(t, input)

		expectStatementCount(t, program, 1)

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

func expectStatementCount(t *testing.T, program *ast.Program, count int) {
	if len(program.Statements) != count {
		t.Fatalf("%q: expected %d program.Statements, got %d", program.String(), count, len(program.Statements))
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

func expectExpressionStatement(t *testing.T, stmt ast.Statement) *ast.ExpressionStatement {
	expressionStmt, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt: expected *ast.ExpressionStatement, got %T", stmt)
	}
	return expressionStmt
}
