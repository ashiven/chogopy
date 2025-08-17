package lexer

import (
	"testing"
)

func TestArithmetic(t *testing.T) {
	stream := `def foo():
	1 + 2`

	expectedTokenList := []Token{
		{DEF, "def", 0},
		{IDENTIFIER, "foo", 4},
		{LROUNDBRACKET, "(", 7},
		{RROUNDBRACKET, ")", 8},
		{COLON, ":", 9},
		{NEWLINE, nil, 10},
		{INDENT, nil, 12},
		{INTEGER, 1, 12},
		{PLUS, "+", 14},
		{INTEGER, 2, 16},
		{NEWLINE, nil, 17},
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
	stream := `
def foo():


	[0]`

	expectedTokenList := []Token{
		{DEF, "def", 1},
		{IDENTIFIER, "foo", 5},
		{LROUNDBRACKET, "(", 8},
		{RROUNDBRACKET, ")", 9},
		{COLON, ":", 10},
		{NEWLINE, nil, 11},
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
	stream := `
def foo():
	0 # Comment with newline
`

	expectedTokenList := []Token{
		{DEF, "def", 1},
		{IDENTIFIER, "foo", 5},
		{LROUNDBRACKET, "(", 8},
		{RROUNDBRACKET, ")", 9},
		{COLON, ":", 10},
		{NEWLINE, nil, 11},
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

func TestFunctionDefinitions(t *testing.T) {
	stream := `
def foo():
	if True:
		return

def bar():
	return

pass
`

	expectedTokenList := []Token{
		{DEF, "def", 1},
		{IDENTIFIER, "foo", 5},
		{LROUNDBRACKET, "(", 8},
		{RROUNDBRACKET, ")", 9},
		{COLON, ":", 10},
		{NEWLINE, nil, 11},
		{INDENT, nil, 13},
		{IF, "if", 13},
		{TRUE, "True", 16},
		{COLON, ":", 20},
		{NEWLINE, nil, 21},
		{INDENT, nil, 24},
		{RETURN, "return", 24},
		{NEWLINE, nil, 30},
		{DEDENT, nil, 32},
		{DEDENT, nil, 32},
		{DEF, "def", 32},
		{IDENTIFIER, "bar", 36},
		{LROUNDBRACKET, "(", 39},
		{RROUNDBRACKET, ")", 40},
		{COLON, ":", 41},
		{NEWLINE, nil, 42},
		{INDENT, nil, 44},
		{RETURN, "return", 44},
		{NEWLINE, nil, 50},
		{DEDENT, nil, 52},
		{PASS, "pass", 52},
		{NEWLINE, nil, 56},
		{EOF, nil, 57},
	}

	lexer := NewLexer(stream)

	for _, expectedToken := range expectedTokenList {
		token := lexer.Consume(false)
		if token.kind != expectedToken.kind || token.value != expectedToken.value || token.offset != expectedToken.offset {
			t.Fatalf("expected: %v (%v) got: %v (%v)", expectedToken.kind.String(), expectedToken, token.kind.String(), token)
		}
	}
}
