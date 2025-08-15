package lexer

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

	nextChar := t.scanner.Consume()
	_ = nextChar

	for {
		break
	}

	return Token{}
}
