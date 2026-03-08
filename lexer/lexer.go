package lexer

import (
	"github.com/akojo/monkey/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

var ops map[byte]token.Token = map[byte]token.Token{
	'=': {Type: token.ASSIGN, Literal: "="},
	'!': {Type: token.BANG, Literal: "!"},
	'+': {Type: token.PLUS, Literal: "+"},
	'-': {Type: token.MINUS, Literal: "-"},
	'/': {Type: token.SLASH, Literal: "/"},
	'*': {Type: token.ASTERISK, Literal: "*"},
	'<': {Type: token.LT, Literal: "<"},
	'>': {Type: token.GT, Literal: ">"},
	',': {Type: token.COMMA, Literal: ","},
	';': {Type: token.SEMICOLON, Literal: ";"},
	'(': {Type: token.LPAREN, Literal: "("},
	')': {Type: token.RPAREN, Literal: ")"},
	'{': {Type: token.LBRACE, Literal: "{"},
	'}': {Type: token.RBRACE, Literal: "}"},
	0:   {Type: token.EOF, Literal: ""},
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	if op, found := ops[l.ch]; found {
		// Special handling for "==" and "!="
		if l.ch == '=' && l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: "=="}
		} else if l.ch == '!' && l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NE, Literal: "!="}
		} else {
			tok = op
		}
	} else {
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
