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
	RSQAUREBRACKET
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
	RSQAUREBRACKET: "RSQAUREBRACKET",
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

func (t *Token) repr() string {
	if t.kind == STRING {
		valCopy := strings.Clone(t.value.(string))
		strings.ReplaceAll(valCopy, "\\", "\\\\")
		strings.ReplaceAll(valCopy, "\t", "\\t")
		strings.ReplaceAll(valCopy, "\r", "\\r")
		strings.ReplaceAll(valCopy, "\n", "\\n")
		// TODO:
		// strings.Replace(t.value.(string), '"', '\\"', -1)
		return t.kind.String() + ":" + valCopy
	}
	return t.kind.String() + fmt.Sprintf("%#v", t.value)
}
