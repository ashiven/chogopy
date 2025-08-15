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
		t.tokenBuffer = append(t.tokenBuffer, t.Consume())
	}

	for len(t.tokenBuffer) < tokenAmount {
		t.tokenBuffer = append(t.tokenBuffer, t.Consume())
	}

	return t.tokenBuffer[:tokenAmount]
}

func (t *Tokenizer) Consume() Token {
	return Token{}
}
