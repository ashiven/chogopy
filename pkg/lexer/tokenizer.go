package lexer

import "slices"

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

func (t *Tokenizer) handleSpace(nextChar string, keepBuffer bool) Token {
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

func (t *Tokenizer) Consume(keepBuffer bool) Token {
	// :param keepBuffer:
	// This is only useful for when we want to look ahead for more than one token. (Inside of Tokenizer.Peek())
	// In that case, only the first token is actually consumed and the rest of the tokens are only inspected.
	if len(t.tokenBuffer) > 0 && !keepBuffer {
		token := t.tokenBuffer[0]
		t.tokenBuffer = t.tokenBuffer[1:]
		return token
	}

	nextChar := t.scanner.Peek()

	for {
		spaceChars := []string{"\t", "\r", "\n", " "}
		if slices.Contains(spaceChars, nextChar) {
			return t.handleSpace(nextChar, keepBuffer)
		}

	}

	return Token{}
}
