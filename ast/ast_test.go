package ast

import (
	"testing"

	"github.com/akojo/monkey/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "var"}, Value: "var"},
				Value: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "value"}, Value: "value"},
			},
		},
	}

	expected := "let var = value;\n"
	got := program.String()

	if got != expected {
		t.Errorf("program.String(): expected %q, got %q", expected, got)
	}
}
