package token

import "fmt"

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers and literals
	IDENT  = "IDENT"  // add, x, y, ...
	INT    = "INT"    // 123456
	STRING = "STRING" // "this is a string"

	// Operators
	ASSIGN   = "="
	BANG     = "!"
	PLUS     = "+"
	MINUS    = "-"
	SLASH    = "/"
	ASTERISK = "*"

	EQ = "=="
	NE = "!="
	LT = "<"
	GT = ">"

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

func (t Token) String() string {
	if t.Type == IDENT || t.Type == INT {
		return fmt.Sprintf("%s(%s)", t.Type, t.Literal)
	}
	return string(t.Type)
}

func NewIdent(ident string) Token {
	if tok, ok := keywords[ident]; ok {
		return Token{Type: tok, Literal: ident}
	}
	return Token{Type: IDENT, Literal: ident}
}
