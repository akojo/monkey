package lexer

import (
	"strings"
	"testing"

	"github.com/akojo/monkey/token"
)

type expectToken struct {
	expectedType    token.TokenType
	expectedLiteral string
}

func TestTokenizeAssignment(t *testing.T) {
	input := `let five = 5;`
	expectTokens := []expectToken{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	testLexer(t, input, expectTokens)
}

func TestTokenizeFunctionDefinition(t *testing.T) {
	input := `
		fn(x, y) {
			x + y;
		};`
	epxectTokens := []expectToken{
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	testLexer(t, input, epxectTokens)
}

func TestTokenizeFunctionCall(t *testing.T) {
	input := `add(five, ten);`
	expectTokens := []expectToken{
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	testLexer(t, input, expectTokens)
}

func TestTokenizeIllegal(t *testing.T) {
	input := "5\a"
	expectTokens := []expectToken{
		{token.INT, "5"},
		{token.ILLEGAL, "\a"},
		{token.EOF, ""},
	}

	testLexer(t, input, expectTokens)
}

func TestTokenizeOperators(t *testing.T) {
	input := `!+-*/5<>;`
	expectTokens := []expectToken{
		{token.BANG, "!"},
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	testLexer(t, input, expectTokens)
}

func TestTokenizeEquals(t *testing.T) {
	input := `
		10 == a;
		b != 9;
	`
	expectTokens := []expectToken{
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.IDENT, "a"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "b"},
		{token.NE, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
	}

	testLexer(t, input, expectTokens)
}

func TestTokenizeKeywords(t *testing.T) {
	input := `
		if (a < 10) {
			return true;
		} else {
			return false;
		}`
	expectTokens := []expectToken{
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
	}

	testLexer(t, input, expectTokens)
}

func TestSkipComments(t *testing.T) {
	input := `
		a = 1; // this is a comment
		b = /* inline comment */ 2;`
	expectTokens := []expectToken{
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "b"},
		{token.ASSIGN, "="},
		{token.INT, "2"},
		{token.SEMICOLON, ";"},
	}

	testLexer(t, input, expectTokens)
}

func TestUnicodeIdentifiers(t *testing.T) {
	input := `ö; í;`
	expectTokens := []expectToken{
		{token.IDENT, "ö"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "í"},
		{token.SEMICOLON, ";"},
	}

	testLexer(t, input, expectTokens)
}

func TestStrings(t *testing.T) {
	input := `
		"foobar"
		"foo bar"
	`
	expectTokens := []expectToken{
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
	}

	testLexer(t, input, expectTokens)
}

func TestArray(t *testing.T) {
	input := `[1, 2]`

	expectTokens := []expectToken{
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
	}

	testLexer(t, input, expectTokens)
}

func testLexer(t *testing.T, input string, expectTokens []expectToken) {
	l := New(strings.NewReader(input), "<test>")

	for i, expectToken := range expectTokens {
		token := l.NextToken()

		if token.Type != expectToken.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=\"%s\", got=\"%s(%s)\"",
				i, expectToken.expectedType, token.Type, token.Literal)
		}

		if token.Literal != expectToken.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, expectToken.expectedLiteral, token.Literal)
		}
	}
}
