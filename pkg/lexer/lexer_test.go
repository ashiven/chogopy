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

func TestLiterals(t *testing.T) {
	stream := `
None
True
False
"Hello"
"He\"ll\"o"
"He\nllo"
"He\\\"llo"
`

	expectedTokenList := []Token{
		{NONE, "None", 1},
		{NEWLINE, nil, 5},
		{TRUE, "True", 6},
		{NEWLINE, nil, 10},
		{FALSE, "False", 11},
		{NEWLINE, nil, 16},
		{STRING, "Hello", 17},
		{NEWLINE, nil, 24},
		{STRING, "He\\\"ll\\\"o", 25},
		{NEWLINE, nil, 36},
		{STRING, "He\\nllo", 37},
		{NEWLINE, nil, 46},
		{STRING, "He\\\\\\\"llo", 47},
	}

	lexer := NewLexer(stream)

	for _, expectedToken := range expectedTokenList {
		token := lexer.Consume(false)
		if token.kind != expectedToken.kind || token.value != expectedToken.value || token.offset != expectedToken.offset {
			t.Fatalf("expected: %v (%v) got: %v (%v)", expectedToken.kind.String(), expectedToken, token.kind.String(), token)
		}
	}
}

func TestPartialDedent(t *testing.T) {
	stream := `
def foo():
	if True:
		pass
	elif False:
		return
`

	expectedTokenList := []Token{
		{DEF, "def", 0},
		{IDENTIFIER, "foo", 0},
		{LROUNDBRACKET, "(", 0},
		{RROUNDBRACKET, ")", 0},
		{COLON, ":", 0},
		{NEWLINE, nil, 0},
		{INDENT, nil, 0},
		{IF, "if", 0},
		{TRUE, "True", 0},
		{COLON, ":", 0},
		{NEWLINE, nil, 0},
		{INDENT, nil, 0},
		{PASS, "pass", 0},
		{NEWLINE, nil, 0},
		{DEDENT, nil, 0},
		{ELIF, "elif", 0},
		{FALSE, "False", 0},
		{COLON, ":", 0},
		{NEWLINE, nil, 0},
		{INDENT, nil, 0},
		{RETURN, "return", 0},
		{NEWLINE, nil, 0},
		{DEDENT, nil, 0},
		{DEDENT, nil, 0},
		{EOF, nil, 0},
	}

	lexer := NewLexer(stream)

	for _, expectedToken := range expectedTokenList {
		token := lexer.Consume(false)
		if token.kind != expectedToken.kind || token.value != expectedToken.value {
			t.Fatalf("expected: %v (%v) got: %v (%v)", expectedToken.kind.String(), expectedToken, token.kind.String(), token)
		}
	}
}
