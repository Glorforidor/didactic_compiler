package ast

import (
	"testing"

	"github.com/Glorforidor/didactic_compiler/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&PrintStatement{
				Token: token.Token{Type: token.Print, Literal: "print"},
				Value: &IntegerLiteral{
					Token: token.Token{Type: token.Int, Literal: "42"},
					Value: 42,
				},
			},
		},
	}

	if program.String() != "print 42" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
