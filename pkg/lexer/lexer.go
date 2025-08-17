package lexer

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
)

var (
	tabSpaces = 8
	spaces    = []string{"\t", "\r", "\n", " "}
	symbols   = []string{"+", "-", "*", "%", "/", "=", "!", "<", ">", "(", ")", ":", "[", "]", ","}
	numbers   = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	letters   = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "J", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "_"}
)

type Lexer struct {
	scanner     Scanner
	tokenBuffer []Token
	isNewLine   bool
	indentLevel int
	indentStack []int
}

func NewLexer(stream string) Lexer {
	scanner := NewScanner(stream)
	return Lexer{
		scanner:     scanner,
		tokenBuffer: []Token{},
		isNewLine:   true,
		indentLevel: 0,
		indentStack: []int{0},
	}
}

func (t *Lexer) Peek(tokenAmount int) []Token {
	if len(t.tokenBuffer) == 0 {
		t.tokenBuffer = append(t.tokenBuffer, t.Consume(false))
	}

	for len(t.tokenBuffer) < tokenAmount {
		t.tokenBuffer = append(t.tokenBuffer, t.Consume(true))
	}

	return t.tokenBuffer[:tokenAmount]
}

func (t *Lexer) Consume(keepBuffer bool) Token {
	// :param keepBuffer:
	// This is only useful for when we want to look ahead for more than one token. (Inside of Lexer.Peek())
	// In that case, only the first token is actually consumed and the rest of the tokens are only inspected.
	if len(t.tokenBuffer) > 0 && !keepBuffer {
		token := t.tokenBuffer[0]
		t.tokenBuffer = t.tokenBuffer[1:]
		return token
	}

	nextChar := t.scanner.Peek()
	for {
		if slices.Contains(spaces, nextChar) {
			fmt.Printf("[%d] handling space\n", t.scanner.offset)
			return t.handleSpaces(nextChar, keepBuffer)
		} else if nextChar == "#" {
			fmt.Printf("[%d] handling comment\n", t.scanner.offset)
			t.handleComment(nextChar)
			continue
		} else if nextChar != "" && t.isNewLine {
			// A new line ends once we encounter the first symbol of the new line
			// which is not a space or a comment (already handled in the previous two cases)
			t.isNewLine = false
			if t.indentLevel > t.indentStack[len(t.indentStack)-1] {
				fmt.Printf("[%d] handling indent\n", t.scanner.offset)
				return t.handleIndent()
			} else if t.indentLevel < t.indentStack[len(t.indentStack)-1] {
				fmt.Printf("[%d] handling dedent\n", t.scanner.offset)
				return t.handleDedent()
			}
		} else if slices.Contains(symbols, nextChar) {
			fmt.Printf("[%d] handling symbol\n", t.scanner.offset)
			return t.handleSymbols(nextChar)
		} else if slices.Contains(letters, nextChar) {
			fmt.Printf("[%d] handling name\n", t.scanner.offset)
			return t.handleName(nextChar)
		} else if slices.Contains(numbers, nextChar) {
			fmt.Printf("[%d] handling int literal\n", t.scanner.offset)
			return t.handleIntegerLiteral(nextChar)
		} else if nextChar == string('"') {
			fmt.Printf("[%d] handling string literal\n", t.scanner.offset)
			return t.handleStringLiteral(nextChar)
		} else if nextChar == "" {
			fmt.Printf("[%d] handling eof\n", t.scanner.offset)
			return t.handleEndOfFile()
		} else {
			log.Fatal(errors.New("invalid symbol in input"))
		}
	}
}

func (t *Lexer) handleSpaces(nextChar string, keepBuffer bool) Token {
	switch nextChar {
	case "\n", "\r":
		// We only want to emit a newline token after a regular line has ended
		// This prevents emitting multiple newline tokens for a series of newlines and instead only emits a single newline for them
		if !t.isNewLine {
			t.isNewLine = true
			t.indentLevel = 0
			t.scanner.Consume()
			return Token{NEWLINE, nil, t.scanner.offset - 1}
		}
	case " ":
		if t.isNewLine {
			t.indentLevel += 1
		}
	case "\t":
		if t.isNewLine {
			// The reason we are subtracing (indentLevel mod tabSpaces) is to end up with proper indentation
			// if for example the source text has been indented via '   \t' or ' \t' (both will lead to 8 spaces)
			normalizedIndentLevel := tabSpaces - t.indentLevel%tabSpaces
			t.indentLevel += normalizedIndentLevel
		}
	}
	t.scanner.Consume()
	return t.Consume(keepBuffer)
}

func (t *Lexer) handleComment(nextChar string) {
	for nextChar != "" && nextChar != "\n" && nextChar != "\r" {
		t.scanner.Consume()
		nextChar = t.scanner.Peek()
	}
}

func (t *Lexer) handleIndent() Token {
	t.indentStack = append(t.indentStack, t.indentLevel)
	// The offset of the scanner needs to be adjusted to mark the beginning of the indent token
	// (the beginning is actually on the same level as the end of the previous indentation)
	indentTokenSize := t.indentStack[len(t.indentStack)-1] - t.indentStack[len(t.indentStack)-2]
	_ = indentTokenSize
	return Token{INDENT, nil, t.scanner.offset}
}

func (t *Lexer) handleDedent() Token {
	dedentTokenSize := t.indentStack[len(t.indentStack)-1] - t.indentLevel
	expectedDedentTokenSize := t.indentStack[len(t.indentStack)-1] - t.indentStack[len(t.indentStack)-2]
	if dedentTokenSize != expectedDedentTokenSize {
		log.Fatal(errors.New("indentation: mismatched blocks"))
	}
	t.indentStack = t.indentStack[:len(t.indentStack)-1]
	return Token{DEDENT, nil, t.scanner.offset}
}

func (t *Lexer) handleSymbols(nextChar string) Token {
	switch nextChar {
	case "+":
		t.scanner.Consume()
		return Token{PLUS, "+", t.scanner.offset - 1}
	case "-":
		t.scanner.Consume()
		if t.scanner.Peek() == ">" {
			t.scanner.Consume()
			return Token{RARROW, "->", t.scanner.offset - 2}
		}
		return Token{MINUS, "-", t.scanner.offset - 1}
	case "*":
		t.scanner.Consume()
		return Token{MUL, "*", t.scanner.offset - 1}
	case "%":
		t.scanner.Consume()
		return Token{MOD, "%", t.scanner.offset - 1}
	case "/":
		t.scanner.Consume()
		if t.scanner.Peek() == "/" {
			t.scanner.Consume()
			return Token{DIV, "//", t.scanner.offset - 2}
		}
		log.Fatal(errors.New("unknown symbol: '/'"))
	case "=":
		t.scanner.Consume()
		if t.scanner.Peek() == "=" {
			t.scanner.Consume()
			return Token{EQ, "==", t.scanner.offset - 2}
		}
		return Token{ASSIGN, "=", t.scanner.offset - 1}
	case "!":
		t.scanner.Consume()
		if t.scanner.Peek() == "=" {
			t.scanner.Consume()
			return Token{NE, "!=", t.scanner.offset - 2}
		}
		log.Fatal(errors.New("unknown symbol: '!'"))
	case "<":
		t.scanner.Consume()
		if t.scanner.Peek() == "=" {
			t.scanner.Consume()
			return Token{LE, "<=", t.scanner.offset - 2}
		}
		return Token{LT, "<", t.scanner.offset - 1}
	case ">":
		t.scanner.Consume()
		if t.scanner.Peek() == "=" {
			t.scanner.Consume()
			return Token{GE, ">=", t.scanner.offset - 2}
		}
		return Token{GT, ">", t.scanner.offset - 1}
	case "(":
		t.scanner.Consume()
		return Token{LROUNDBRACKET, "(", t.scanner.offset - 1}
	case ")":
		t.scanner.Consume()
		return Token{RROUNDBRACKET, ")", t.scanner.offset - 1}
	case ":":
		t.scanner.Consume()
		return Token{COLON, ":", t.scanner.offset - 1}
	case "[":
		t.scanner.Consume()
		return Token{LSQUAREBRACKET, "[", t.scanner.offset - 1}
	case "]":
		t.scanner.Consume()
		return Token{RSQUAREBRACKET, "]", t.scanner.offset - 1}
	case ",":
		t.scanner.Consume()
		return Token{COMMA, ",", t.scanner.offset - 1}
	}
	return Token{}
}

func (t *Lexer) handleName(nextChar string) Token {
	name := ""
	for slices.Contains(letters, nextChar) || slices.Contains(numbers, nextChar) {
		name += nextChar
		t.scanner.Consume()
		nextChar = t.scanner.Peek()
	}

	switch name {
	case "class":
		return Token{CLASS, name, t.scanner.offset - len(name)}
	case "def":
		return Token{DEF, name, t.scanner.offset - len(name)}
	case "global":
		return Token{GLOBAL, name, t.scanner.offset - len(name)}
	case "nonlocal":
		return Token{NONLOCAL, name, t.scanner.offset - len(name)}
	case "if":
		return Token{IF, name, t.scanner.offset - len(name)}
	case "elif":
		return Token{ELIF, name, t.scanner.offset - len(name)}
	case "else":
		return Token{ELSE, name, t.scanner.offset - len(name)}
	case "while":
		return Token{WHILE, name, t.scanner.offset - len(name)}
	case "for":
		return Token{FOR, name, t.scanner.offset - len(name)}
	case "in":
		return Token{IN, name, t.scanner.offset - len(name)}
	case "None":
		return Token{NONE, name, t.scanner.offset - len(name)}
	case "True":
		return Token{TRUE, name, t.scanner.offset - len(name)}
	case "False":
		return Token{FALSE, name, t.scanner.offset - len(name)}
	case "pass":
		return Token{PASS, name, t.scanner.offset - len(name)}
	case "or":
		return Token{OR, name, t.scanner.offset - len(name)}
	case "and":
		return Token{AND, name, t.scanner.offset - len(name)}
	case "not":
		return Token{NOT, name, t.scanner.offset - len(name)}
	case "is":
		return Token{IS, name, t.scanner.offset - len(name)}
	case "object":
		return Token{OBJECT, name, t.scanner.offset - len(name)}
	case "int":
		return Token{INT, name, t.scanner.offset - len(name)}
	case "bool":
		return Token{BOOL, name, t.scanner.offset - len(name)}
	case "str":
		return Token{STR, name, t.scanner.offset - len(name)}
	case "return":
		return Token{RETURN, name, t.scanner.offset - len(name)}
	}

	return Token{IDENTIFIER, name, t.scanner.offset - len(name)}
}

func (t *Lexer) handleIntegerLiteral(nextChar string) Token {
	value := ""
	for slices.Contains(numbers, nextChar) {
		value += nextChar
		t.scanner.Consume()
		nextChar = t.scanner.Peek()
	}

	valueInt, err := strconv.Atoi(value)
	if err != nil {
		log.Fatal(errors.New("failed to convert integer literal"))
	}

	return Token{INTEGER, valueInt, t.scanner.offset - len(value)}
}

func (t *Lexer) handleStringLiteral(nextChar string) Token {
	value := nextChar
	nextChar = t.scanner.Consume()

	allowedEscapedChars := []string{"t", "n", "\\", string('"')}
	for nextChar != string('"') {
		if nextChar == "\\" {
			escapedChar := t.scanner.Peek()
			if !slices.Contains(allowedEscapedChars, escapedChar) {
				log.Fatal(errors.New("unknown escape sequence"))
			}
		}
		nextChar = t.scanner.Consume()
		value += nextChar
	}

	return Token{STRING, value, t.scanner.offset - len(value)}
}

func (t *Lexer) handleEndOfFile() Token {
	// automatically emit a new line when at the end of the last line
	if !t.isNewLine {
		t.isNewLine = true
		return Token{NEWLINE, nil, t.scanner.offset}
	}
	// emit a dedent token for all remaining indentation levels
	if t.indentStack[len(t.indentStack)-1] > 0 {
		dedentTokenSize := t.indentStack[len(t.indentStack)-1] - t.indentStack[len(t.indentStack)-2]
		_ = dedentTokenSize
		t.indentStack = t.indentStack[:len(t.indentStack)-1]
		return Token{DEDENT, nil, t.scanner.offset}
	}
	return Token{EOF, nil, t.scanner.offset}
}
