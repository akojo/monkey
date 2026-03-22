package lexer

import (
	"io"
	"strings"
	"text/scanner"
	"unicode"

	"github.com/akojo/monkey/token"
)

type Lexer struct {
	Filename string
	scanner  scanner.Scanner
}

var ops map[rune]token.Token = map[rune]token.Token{
	'=': {Type: token.ASSIGN, Literal: "="},
	'!': {Type: token.BANG, Literal: "!"},
	'+': {Type: token.PLUS, Literal: "+"},
	'-': {Type: token.MINUS, Literal: "-"},
	'/': {Type: token.SLASH, Literal: "/"},
	'*': {Type: token.ASTERISK, Literal: "*"},
	':': {Type: token.COLON, Literal: ":"},
	'<': {Type: token.LT, Literal: "<"},
	'>': {Type: token.GT, Literal: ">"},
	',': {Type: token.COMMA, Literal: ","},
	';': {Type: token.SEMICOLON, Literal: ";"},
	'(': {Type: token.LPAREN, Literal: "("},
	')': {Type: token.RPAREN, Literal: ")"},
	'{': {Type: token.LBRACE, Literal: "{"},
	'}': {Type: token.RBRACE, Literal: "}"},
	'[': {Type: token.LBRACKET, Literal: "["},
	']': {Type: token.RBRACKET, Literal: "]"},
	0:   {Type: token.EOF, Literal: ""},
}

func New(input io.Reader, filename string) *Lexer {
	l := &Lexer{Filename: filename}

	l.scanner.Init(input)
	l.scanner.Filename = filename
	l.scanner.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanStrings | scanner.ScanComments | scanner.SkipComments
	l.scanner.IsIdentRune = func(ch rune, i int) bool {
		return unicode.IsLetter(ch) || ch == '_' || (i > 0 && unicode.IsDigit(ch))
	}

	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	ch := l.scanner.Scan()

	if ch == scanner.EOF {
		return token.Token{Type: token.EOF, Literal: ""}
	}

	line := l.scanner.Position.Line
	column := l.scanner.Position.Column

	if op, found := ops[ch]; found {
		// Special handling for "==" and "!="
		if ch == '=' && l.scanner.Peek() == '=' {
			l.scanner.Next()
			tok = token.Token{Type: token.EQ, Literal: "=="}
		} else if ch == '!' && l.scanner.Peek() == '=' {
			l.scanner.Next()
			tok = token.Token{Type: token.NE, Literal: "!="}
		} else {
			tok = op
		}
	} else {
		switch ch {
		case scanner.Ident:
			tok = token.NewIdent(l.scanner.TokenText())
		case scanner.Int:
			tok = token.Token{Type: token.INT, Literal: l.scanner.TokenText()}
		case scanner.String:
			tok = token.Token{Type: token.STRING, Literal: strings.Trim(l.scanner.TokenText(), `"`)}
		default:
			tok = token.Token{Type: token.ILLEGAL, Literal: string(ch)}
		}
	}

	tok.Line = line
	tok.Column = column

	return tok
}
