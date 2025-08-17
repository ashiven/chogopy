package lexer

import (
	"testing"
)

func TestArithmetic(t *testing.T) {
	stream := "def foo():\n 	1 + 2"

	expectedTokenList := []Token{
		{DEF, "def", 0},
		{IDENTIFIER, "foo", 4},
		{LROUNDBRACKET, "(", 7},
		{RROUNDBRACKET, ")", 8},
		{COLON, ":", 9},
		{NEWLINE, nil, 10},
		{INDENT, nil, 13},
		{INTEGER, 1, 13},
		{PLUS, "+", 15},
		{INTEGER, 2, 17},
		{NEWLINE, nil, 18},
		{DEDENT, nil, 18},
		{EOF, nil, 18},
	}

	lexer := NewLexer(stream)

	for _, expectedToken := range expectedTokenList {
		token := lexer.Consume(false)
		if token.kind != expectedToken.kind || token.value != expectedToken.value || token.offset != expectedToken.offset {
			t.Fatalf("expected: %v (%v) got: %v (%v)", expectedToken.kind.String(), expectedToken, token.kind.String(), token)
		}
	}
}

func TestBrackets(t *testing.T) {
	stream := `def foo():


		[0]`

	expectedTokenList := []Token{
		{DEF, "def", 0},
		{IDENTIFIER, "foo", 4},
		{LROUNDBRACKET, "(", 7},
		{RROUNDBRACKET, ")", 8},
		{COLON, ":", 9},
		{NEWLINE, nil, 10},
		{INDENT, nil, 15},
		{LSQUAREBRACKET, "[", 15},
		{INTEGER, 0, 16},
		{RSQUAREBRACKET, "]", 17},
		{NEWLINE, nil, 18},
		{DEDENT, nil, 18},
		{EOF, nil, 18},
	}

	lexer := NewLexer(stream)

	for _, expectedToken := range expectedTokenList {
		token := lexer.Consume(false)
		if token.kind != expectedToken.kind || token.value != expectedToken.value || token.offset != expectedToken.offset {
			t.Fatalf("expected: %v (%v) got: %v (%v)", expectedToken.kind.String(), expectedToken, token.kind.String(), token)
		}
	}
}
