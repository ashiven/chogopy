package lexer

import (
	"errors"
	"log"
	"slices"
)

type Tokenizer struct {
	scanner       Scanner
	tokenBuffer   []Token
	isNewLine     bool
	isLogicalLine bool
	indentLevel   int
	indentStack   []int
}

func NewTokenizer(scanner Scanner) Tokenizer {
	return Tokenizer{
		scanner:       scanner,
		tokenBuffer:   []Token{},
		isNewLine:     true,
		isLogicalLine: false,
		indentLevel:   0,
		indentStack:   []int{0},
	}
}

func (t *Tokenizer) Peek(tokenAmount int) []Token {
	if len(t.tokenBuffer) == 0 {
		t.tokenBuffer = append(t.tokenBuffer, t.Consume(false))
	}

	for len(t.tokenBuffer) < tokenAmount {
		t.tokenBuffer = append(t.tokenBuffer, t.Consume(true))
	}

	return t.tokenBuffer[:tokenAmount]
}

func (t *Tokenizer) Consume(keepBuffer bool) Token {
	// :param keepBuffer:
	// This is only useful for when we want to look ahead for more than one token. (Inside of Tokenizer.Peek())
	// In that case, only the first token is actually consumed and the rest of the tokens are only inspected.
	if len(t.tokenBuffer) > 0 && !keepBuffer {
		token := t.tokenBuffer[0]
		t.tokenBuffer = t.tokenBuffer[1:]
		return token
	}

	spaceChars := []string{"\t", "\r", "\n", " "}
	symbols := []string{"+", "-", "*", "%", "/", "=", "!", "<", ">", "(", ")", ":", "[", "]", ","}

	nextChar := t.scanner.Peek()
	for {
		if slices.Contains(spaceChars, nextChar) {
			return t.handleSpaces(nextChar, keepBuffer)
		} else if nextChar == "#" {
			t.handleComment(nextChar)
			continue
		} else if nextChar != "" && t.isNewLine {
			// A logical line starts once we encounter the first symbol of a new line
			// which is not a space or a comment (already handled in the previous two cases)
			t.isLogicalLine = true
			t.isNewLine = false
			// The current line has more spaces than the previous indented lines
			// therefore it requires us to emit and save a new level of indentation
			if t.indentLevel > t.indentStack[len(t.indentStack)-1] {
				return t.handleIndent()
			} else if t.indentLevel < t.indentStack[len(t.indentStack)-1] {
				return t.handleDedent()
			}
		} else if slices.Contains(symbols, nextChar) {
			return t.handleSymbols(nextChar)
		}
	}
}

func (t *Tokenizer) handleSpaces(nextChar string, keepBuffer bool) Token {
	const tabSpaces = 8
	switch nextChar {
	case "\t":
		// isNewLine only gets set after encountering \n or \r (we are processing the start of a new line)
		if t.isNewLine {
			// The reason we are subtracing (indentLevel mod tabSpaces) is to end up with proper indentation
			// if for example the source text has been indented via '   \t' or ' \t' (both will lead to 8 spaces)
			t.indentLevel += tabSpaces - t.indentLevel%tabSpaces
		}
	case "\n", "\r":
		t.indentLevel = 0
		t.isNewLine = true
		if t.isLogicalLine {
			t.isLogicalLine = false
			t.scanner.Consume()
			return Token{NEWLINE, nil, t.scanner.offset}
		}
	case " ":
		if t.isNewLine {
			t.indentLevel += 1
		}
	}
	// Make sure to consume the nextChar which we have only peeked until now
	t.scanner.Consume()
	return t.Consume(keepBuffer)
}

func (t *Tokenizer) handleComment(nextChar string) {
	for nextChar != "" && nextChar != "\n" && nextChar != "\r" {
		t.scanner.Consume()
		nextChar = t.scanner.Peek()
	}
}

func (t *Tokenizer) handleIndent() Token {
	t.indentStack = append(t.indentStack, t.indentLevel)
	// The offset of the scanner needs to be adjusted to mark the beginning of the indent token
	// (the beginning is actually on the same level as the end of the previous indentation)
	indentTokenSize := t.indentStack[len(t.indentStack)-1] - t.indentStack[len(t.indentStack)-2]
	return Token{INDENT, nil, t.scanner.offset - indentTokenSize}
}

func (t *Tokenizer) handleDedent() Token {
	dedentTokenSize := t.indentStack[len(t.indentStack)-1] - t.indentLevel
	expectedDedentTokenSize := t.indentStack[len(t.indentStack)-1] - t.indentStack[len(t.indentStack)-2]
	if dedentTokenSize != expectedDedentTokenSize {
		log.Fatal(errors.New("indentation: mismatched blocks"))
	}
	t.indentStack = t.indentStack[:len(t.indentStack)-1]
	return Token{DEDENT, nil, t.scanner.offset - dedentTokenSize}
}

func (t *Tokenizer) handleSymbols(nextChar string) Token {
	switch nextChar {
	case "+":
		t.scanner.Consume()
		return Token{PLUS, "+", t.scanner.offset}
	case "-":
		t.scanner.Consume()
		if t.scanner.Peek() == ">" {
			t.scanner.Consume()
			return Token{RARROW, "->", t.scanner.offset - 1}
		}
		return Token{MINUS, "-", t.scanner.offset}
	case "*":
		t.scanner.Consume()
		return Token{MUL, "*", t.scanner.offset}
	case "%":
		t.scanner.Consume()
		return Token{MOD, "%", t.scanner.offset}
	case "/":
		t.scanner.Consume()
		if t.scanner.Peek() == "/" {
			t.scanner.Consume()
			return Token{DIV, "//", t.scanner.offset - 1}
		}
		log.Fatal(errors.New("unknown symbol: '/'"))
	case "=":
		t.scanner.Consume()
		if t.scanner.Peek() == "=" {
			t.scanner.Consume()
			return Token{EQ, "==", t.scanner.offset - 1}
		}
		return Token{ASSIGN, "=", t.scanner.offset}
	case "!":
		t.scanner.Consume()
		if t.scanner.Peek() == "=" {
			t.scanner.Consume()
			return Token{NE, "!=", t.scanner.offset - 1}
		}
		log.Fatal(errors.New("unknown symbol: '!'"))
	case "<":
		t.scanner.Consume()
		if t.scanner.Peek() == "=" {
			t.scanner.Consume()
			return Token{LE, "<=", t.scanner.offset - 1}
		}
		return Token{LT, "<", t.scanner.offset}
	case ">":
		t.scanner.Consume()
		if t.scanner.Peek() == "=" {
			t.scanner.Consume()
			return Token{GE, ">=", t.scanner.offset - 1}
		}
		return Token{GT, ">", t.scanner.offset}
	case "(":
		t.scanner.Consume()
		return Token{PLUS, "+", t.scanner.offset}
	case ")":
		t.scanner.Consume()
		return Token{PLUS, "+", t.scanner.offset}
	case ":":
		t.scanner.Consume()
		return Token{PLUS, "+", t.scanner.offset}
	case "[":
		t.scanner.Consume()
		return Token{PLUS, "+", t.scanner.offset}
	case "]":
		t.scanner.Consume()
		return Token{PLUS, "+", t.scanner.offset}
	case ",":
		t.scanner.Consume()
		return Token{PLUS, "+", t.scanner.offset}
	}
	return Token{}
}
