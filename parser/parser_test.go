package parser

import (
	"strconv"
	"strings"
	"testing"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/lexer"
)

func TestLetStatements(t *testing.T) {
	testLetStatement := func(t *testing.T, s ast.Statement, name string) bool {
		if s.TokenLiteral() != "let" {
			t.Errorf("s.TokenLiteral: expected \"let\", got %q", s.TokenLiteral())
			return false
		}

		letStmt, ok := s.(*ast.LetStatement)
		if !ok {
			t.Errorf("s: expected *ast.LetStatement, got %T", s)
			return false
		}

		if letStmt.Name.Value != name {
			t.Errorf("letStmt.Name.Value: expected %q, got %q", name, letStmt.Name.Value)
			return false
		}

		if letStmt.Name.TokenLiteral() != name {
			t.Errorf("letStmt.Name.TokenLiteral(): expected %q, got %q", name, letStmt.Name.TokenLiteral())
			return false
		}

		return true
	}

	program := makeProgram(t, `
		let x = 5;
		let y = 10;`)

	expectStatementCount(t, program, 2)

	testLetStatement(t, program.Statements[0], "x")
	testLetStatement(t, program.Statements[1], "y")
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

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("ident: expected *ast.Identifier, got %T", ident)
	}
	if ident.Value != "foobar" {
		t.Errorf("ident.Value: expected \"foobar\", got %q", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral: expected \"foobar\", got %q", ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	program := makeProgram(t, "5;")

	expectStatementCount(t, program, 1)

	stmt := expectExpressionStatement(t, program.Statements[0])

	expectIntegerLiteral(t, stmt.Expression, 5)
}

func TestPrefixExpressions(t *testing.T) {
	testPrefixExpression := func(input string, op string, value int64) {
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

	testPrefixExpression("!5;", "!", 5)
	testPrefixExpression("-15;", "-", 15)
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
		t.Fatalf("expected %d program.Statements, got %d", count, len(program.Statements))
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
