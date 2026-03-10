package parser

import (
	"strings"
	"testing"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/lexer"
)

func TestLetStatements(t *testing.T) {
	program := makeProgram(t, `
		let x = 5;
		let y = 10;`)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	expectStatements(t, program, 2)

	tests := []struct {
		expectIdentifier string
	}{
		{"x"},
		{"y"},
	}
	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectIdentifier) {
			return
		}
	}
}

func TestReturnStatements(t *testing.T) {
	program := makeProgram(t, `
		return 5;
		return 10;`)

	expectStatements(t, program, 2)

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

	expectStatements(t, program, 1)

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
		t.Errorf("ident.TokenLiteral, expected \"foobar\", got %q", ident.TokenLiteral())
	}
}

func makeProgram(t *testing.T, input string) *ast.Program {
	l := lexer.New(strings.NewReader(input), "<test>")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	return program
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
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

func expectStatements(t *testing.T, program *ast.Program, count int) {
	if len(program.Statements) != count {
		t.Fatalf("expected %d program.Statements, got %d", count, len(program.Statements))
	}
}
