// Package lexer implements a lexer for the chocopy programming language.
package lexer

import (
	"fmt"
	"strings"
)

type TokenKind int

const (
	EOF TokenKind = iota

	NEWLINE

	INDENT
	DEDENT

	CLASS
	DEF
	GLOBAL
	NONLOCAL
	PASS
	RETURN

	IDENTIFIER
	INTEGER
	STRING

	IF
	ELIF
	ELSE
	WHILE
	FOR
	IN

	PLUS
	MINUS
	MUL
	DIV
	MOD
	ASSIGN
	LROUNDBRACKET
	RROUNDBRACKET
	COLON
	LSQUAREBRACKET
	RSQUAREBRACKET
	COMMA
	RARROW

	EQ
	NE
	LT
	GT
	LE
	GE
	IS

	NONE
	TRUE
	FALSE

	OR
	AND
	NOT

	OBJECT
	INT
	BOOL
	STR
)

var TokenKindName = map[TokenKind]string{
	EOF: "EOF",

	NEWLINE: "NEWLINE",

	INDENT: "INDENT",
	DEDENT: "DEDENT",

	CLASS:    "CLASS",
	DEF:      "DEF",
	GLOBAL:   "GLOBAL",
	NONLOCAL: "NONLOCAL",
	PASS:     "PASS",
	RETURN:   "RETURN",

	IDENTIFIER: "IDENTIFIER",
	INTEGER:    "INTEGER",
	STRING:     "STRING",

	IF:    "IF",
	ELIF:  "ELIF",
	ELSE:  "ELSE",
	WHILE: "WHILE",
	FOR:   "FOR",
	IN:    "IN",

	PLUS:           "PLUS",
	MINUS:          "MINUS",
	MUL:            "MUL",
	DIV:            "DIV",
	MOD:            "MOD",
	ASSIGN:         "ASSIGN",
	LROUNDBRACKET:  "LROUNDBRACKET",
	RROUNDBRACKET:  "RROUNDBRACKET",
	COLON:          "COLON",
	LSQUAREBRACKET: "LSQUAREBRACKET",
	RSQUAREBRACKET: "RSQUAREBRACKET",
	COMMA:          "COMMA",
	RARROW:         "RARROW",

	EQ: "EQ",
	NE: "NE",
	LT: "LT",
	GT: "GT",
	LE: "LE",
	GE: "GE",
	IS: "IS",

	NONE:  "NONE",
	TRUE:  "TRUE",
	FALSE: "FALSE",

	OR:  "OR",
	AND: "AND",
	NOT: "NOT",

	OBJECT: "OBJECT",
	INT:    "INT",
	BOOL:   "BOOL",
	STR:    "STR",
}

func (tk TokenKind) String() string {
	return TokenKindName[tk]
}

type Token struct {
	kind   TokenKind
	value  any
	offset int
}

func (t *Token) Repr() string {
	if t.kind == STRING {
		valCopy := strings.Clone(t.value.(string))
		valCopy = strings.ReplaceAll(valCopy, "\\", "\\\\")
		valCopy = strings.ReplaceAll(valCopy, "\t", "\\t")
		valCopy = strings.ReplaceAll(valCopy, "\r", "\\r")
		valCopy = strings.ReplaceAll(valCopy, "\n", "\\n")
		// TODO:
		// strings.Replace(t.value.(string), '"', '\\"', -1)
		return t.kind.String() + ":" + valCopy
	}
	return t.kind.String() + fmt.Sprintf("%#v", t.value)
}

// KindToken creates a token that can be used for kind comparisons in the Parser and
// is populated with dummy values for other attributes.
func KindToken(kind TokenKind) *Token {
	return &Token{
		kind:   kind,
		value:  nil,
		offset: 0,
	}
}

// TokenSlice is a utility function to return a slice containing a pointer to a single token.
// It can be used by the parser for simple arg construction for match and check.
func TokenSlice(kinds ...TokenKind) []*Token {
	tokenSlice := []*Token{}

	for _, kind := range kinds {
		tokenSlice = append(tokenSlice, KindToken(kind))
	}
	return tokenSlice
}

// KindEquals is a utility function to compare two tokens only by their kind, which can be used in the Parser.
func (t *Token) KindEquals(other *Token) bool {
	return t.kind == other.kind
}
