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
		{DEDENT, nil, 19},
		{EOF, nil, 19},
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
		{DEDENT, nil, 19},
		{EOF, nil, 19},
	}

	lexer := NewLexer(stream)

	for _, expectedToken := range expectedTokenList {
		token := lexer.Consume(false)
		if token.kind != expectedToken.kind || token.value != expectedToken.value || token.offset != expectedToken.offset {
			t.Fatalf("expected: %v (%v) got: %v (%v)", expectedToken.kind.String(), expectedToken, token.kind.String(), token)
		}
	}
}

func TestDivision(t *testing.T) {
	stream := "0 // 1"

	expectedTokenList := []Token{
		{INTEGER, 0, 0},
		{DIV, "//", 2},
		{INTEGER, 1, 5},
		{NEWLINE, nil, 6},
		{EOF, nil, 7},
	}

	lexer := NewLexer(stream)

	for _, expectedToken := range expectedTokenList {
		token := lexer.Consume(false)
		if token.kind != expectedToken.kind || token.value != expectedToken.value || token.offset != expectedToken.offset {
			t.Fatalf("expected: %v (%v) got: %v (%v)", expectedToken.kind.String(), expectedToken, token.kind.String(), token)
		}
	}
}

func TestEndWithComment(t *testing.T) {
	stream := `def foo():
		0 # Comment with newline
`

	expectedTokenList := []Token{
		{DEF, "def", 0},
		{IDENTIFIER, "foo", 4},
		{LROUNDBRACKET, "(", 7},
		{RROUNDBRACKET, ")", 8},
		{COLON, ":", 9},
		{NEWLINE, nil, 10},
		{INDENT, nil, 13},
		{INTEGER, 0, 13},
		{NEWLINE, nil, 37},
		{DEDENT, nil, 38},
		{EOF, nil, 38},
	}

	lexer := NewLexer(stream)

	for _, expectedToken := range expectedTokenList {
		token := lexer.Consume(false)
		if token.kind != expectedToken.kind || token.value != expectedToken.value || token.offset != expectedToken.offset {
			t.Fatalf("expected: %v (%v) got: %v (%v)", expectedToken.kind.String(), expectedToken, token.kind.String(), token)
		}
	}
}
