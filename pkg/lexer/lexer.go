package lexer

import (
	"errors"
	"log"
	"slices"
	"strconv"
)

var (
	tabSpaces = 8
	spaces    = []string{"\t", "\r", "\n", " "}
	symbols   = []string{"+", "-", "*", "%", "/", "=", "!", "<", ">", "(", ")", ":", "[", "]", ","}
	numbers   = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	letters   = []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
		"A", "B", "C", "D", "E", "J", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
		"_",
	}
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

type LocationInfo struct {
	Line        int
	Column      int
	LineLiteral string
}

func (l *Lexer) GetLocation(token *Token) *LocationInfo {
	line := 1
	column := 0
	lineLiteral := ""

	for charIndex, char := range l.scanner.streamLookup {
		if charIndex == token.Offset {
			column++
			break
		}
		if charIndex > token.Offset && char == '\n' {
			break
		}

		if charIndex > token.Offset {
			lineLiteral += string(char)
		} else if char == '\n' {
			line++
			column = 0
			lineLiteral = ""
		} else {
			column++
			lineLiteral += string(char)
		}
	}

	return &LocationInfo{line, column, lineLiteral}
}

func (l *Lexer) Peek(tokenAmount int) []Token {
	if len(l.tokenBuffer) == 0 {
		l.tokenBuffer = append(l.tokenBuffer, l.Consume(false))
	}

	for len(l.tokenBuffer) < tokenAmount {
		l.tokenBuffer = append(l.tokenBuffer, l.Consume(true))
	}

	return l.tokenBuffer[:tokenAmount]
}

func (l *Lexer) Consume(keepBuffer bool) Token {
	// :param keepBuffer:
	// This is only useful for when we want to look ahead for more than one token. (Inside of Lexer.Peek())
	// In that case, only the first token is actually consumed and the rest of the tokens are only inspected.
	if len(l.tokenBuffer) > 0 && !keepBuffer {
		token := l.tokenBuffer[0]
		l.tokenBuffer = l.tokenBuffer[1:]
		return token
	}

	nextChar := l.scanner.Peek()
	for {
		if slices.Contains(spaces, nextChar) {
			return l.handleSpaces(nextChar, keepBuffer)
		} else if nextChar == "#" {
			l.handleComment(nextChar)
			return l.Consume(keepBuffer)
		} else if nextChar != "" && l.isNewLine {
			if l.indentLevel > l.indentStack[len(l.indentStack)-1] {
				return l.handleIndent()
			} else if l.indentLevel < l.indentStack[len(l.indentStack)-1] {
				return l.handleDedent()
			}
			// A new line ends once we encounter the first symbol of the new line
			// which is not a space or a comment (already handled in the previous two cases)
			// AND after we have emitted all the necessary indent/dedent tokens
			l.isNewLine = false
		} else if slices.Contains(symbols, nextChar) {
			return l.handleSymbols(nextChar)
		} else if slices.Contains(letters, nextChar) {
			return l.handleName(nextChar)
		} else if slices.Contains(numbers, nextChar) {
			return l.handleIntegerLiteral(nextChar)
		} else if nextChar == string('"') {
			return l.handleStringLiteral()
		} else if nextChar == "" {
			return l.handleEndOfFile()
		} else {
			log.Fatal(errors.New("invalid symbol in input"))
		}
	}
}

func (l *Lexer) handleSpaces(nextChar string, keepBuffer bool) Token {
	switch nextChar {
	case "\n", "\r":
		// We only want to emit a newline token after a regular line has ended
		// This prevents emitting multiple newline tokens for a series of newlines and instead only emits a single newline for them
		if !l.isNewLine {
			l.isNewLine = true
			l.indentLevel = 0
			l.scanner.Consume()
			return Token{NEWLINE, nil, l.scanner.offset - 1}
		}
	case " ":
		if l.isNewLine {
			l.indentLevel += 1
		}
	case "\t":
		if l.isNewLine {
			// The reason we are subtracing (indentLevel mod tabSpaces) is to end up with proper indentation
			// if for example the source text has been indented via '   \t' or ' \t' (both will lead to 8 spaces)
			normalizedIndentLevel := tabSpaces - l.indentLevel%tabSpaces
			l.indentLevel += normalizedIndentLevel
		}
	}
	l.scanner.Consume()
	return l.Consume(keepBuffer)
}

func (l *Lexer) handleComment(nextChar string) {
	for nextChar != "" && nextChar != "\n" && nextChar != "\r" {
		l.scanner.Consume()
		nextChar = l.scanner.Peek()
	}
	l.indentLevel = 0
}

func (l *Lexer) handleIndent() Token {
	l.indentStack = append(l.indentStack, l.indentLevel)
	indentTokenSize := l.indentStack[len(l.indentStack)-1] - l.indentStack[len(l.indentStack)-2]
	_ = indentTokenSize
	return Token{INDENT, nil, l.scanner.offset}
}

func (l *Lexer) handleDedent() Token {
	// if the current indentation doesn't match any of the previous ones -> mismatch
	if !slices.Contains(l.indentStack, l.indentLevel) {
		log.Fatal(errors.New("indentation: mismatched blocks"))
	}
	l.indentStack = l.indentStack[:len(l.indentStack)-1]
	return Token{DEDENT, nil, l.scanner.offset}
}

func (l *Lexer) handleSymbols(nextChar string) Token {
	switch nextChar {
	case "+":
		l.scanner.Consume()
		return Token{PLUS, "+", l.scanner.offset - 1}
	case "-":
		l.scanner.Consume()
		if l.scanner.Peek() == ">" {
			l.scanner.Consume()
			return Token{RARROW, "->", l.scanner.offset - 2}
		}
		return Token{MINUS, "-", l.scanner.offset - 1}
	case "*":
		l.scanner.Consume()
		return Token{MUL, "*", l.scanner.offset - 1}
	case "%":
		l.scanner.Consume()
		return Token{MOD, "%", l.scanner.offset - 1}
	case "/":
		l.scanner.Consume()
		if l.scanner.Peek() == "/" {
			l.scanner.Consume()
			return Token{DIV, "//", l.scanner.offset - 2}
		}
		log.Fatal(errors.New("unknown symbol: '/'"))
	case "=":
		l.scanner.Consume()
		if l.scanner.Peek() == "=" {
			l.scanner.Consume()
			return Token{EQ, "==", l.scanner.offset - 2}
		}
		return Token{ASSIGN, "=", l.scanner.offset - 1}
	case "!":
		l.scanner.Consume()
		if l.scanner.Peek() == "=" {
			l.scanner.Consume()
			return Token{NE, "!=", l.scanner.offset - 2}
		}
		log.Fatal(errors.New("unknown symbol: '!'"))
	case "<":
		l.scanner.Consume()
		if l.scanner.Peek() == "=" {
			l.scanner.Consume()
			return Token{LE, "<=", l.scanner.offset - 2}
		}
		return Token{LT, "<", l.scanner.offset - 1}
	case ">":
		l.scanner.Consume()
		if l.scanner.Peek() == "=" {
			l.scanner.Consume()
			return Token{GE, ">=", l.scanner.offset - 2}
		}
		return Token{GT, ">", l.scanner.offset - 1}
	case "(":
		l.scanner.Consume()
		return Token{LROUNDBRACKET, "(", l.scanner.offset - 1}
	case ")":
		l.scanner.Consume()
		return Token{RROUNDBRACKET, ")", l.scanner.offset - 1}
	case ":":
		l.scanner.Consume()
		return Token{COLON, ":", l.scanner.offset - 1}
	case "[":
		l.scanner.Consume()
		return Token{LSQUAREBRACKET, "[", l.scanner.offset - 1}
	case "]":
		l.scanner.Consume()
		return Token{RSQUAREBRACKET, "]", l.scanner.offset - 1}
	case ",":
		l.scanner.Consume()
		return Token{COMMA, ",", l.scanner.offset - 1}
	}
	return Token{}
}

func (l *Lexer) handleName(nextChar string) Token {
	name := ""
	for slices.Contains(letters, nextChar) || slices.Contains(numbers, nextChar) {
		name += nextChar
		l.scanner.Consume()
		nextChar = l.scanner.Peek()
	}

	switch name {
	case "class":
		return Token{CLASS, name, l.scanner.offset - len(name)}
	case "def":
		return Token{DEF, name, l.scanner.offset - len(name)}
	case "global":
		return Token{GLOBAL, name, l.scanner.offset - len(name)}
	case "nonlocal":
		return Token{NONLOCAL, name, l.scanner.offset - len(name)}
	case "if":
		return Token{IF, name, l.scanner.offset - len(name)}
	case "elif":
		return Token{ELIF, name, l.scanner.offset - len(name)}
	case "else":
		return Token{ELSE, name, l.scanner.offset - len(name)}
	case "while":
		return Token{WHILE, name, l.scanner.offset - len(name)}
	case "for":
		return Token{FOR, name, l.scanner.offset - len(name)}
	case "in":
		return Token{IN, name, l.scanner.offset - len(name)}
	case "None":
		return Token{NONE, name, l.scanner.offset - len(name)}
	case "True":
		return Token{TRUE, name, l.scanner.offset - len(name)}
	case "False":
		return Token{FALSE, name, l.scanner.offset - len(name)}
	case "pass":
		return Token{PASS, name, l.scanner.offset - len(name)}
	case "or":
		return Token{OR, name, l.scanner.offset - len(name)}
	case "and":
		return Token{AND, name, l.scanner.offset - len(name)}
	case "not":
		return Token{NOT, name, l.scanner.offset - len(name)}
	case "is":
		return Token{IS, name, l.scanner.offset - len(name)}
	case "object":
		return Token{OBJECT, name, l.scanner.offset - len(name)}
	case "int":
		return Token{INT, name, l.scanner.offset - len(name)}
	case "bool":
		return Token{BOOL, name, l.scanner.offset - len(name)}
	case "str":
		return Token{STR, name, l.scanner.offset - len(name)}
	case "return":
		return Token{RETURN, name, l.scanner.offset - len(name)}
	}

	return Token{IDENTIFIER, name, l.scanner.offset - len(name)}
}

func (l *Lexer) handleIntegerLiteral(nextChar string) Token {
	value := ""
	for slices.Contains(numbers, nextChar) {
		value += nextChar
		l.scanner.Consume()
		nextChar = l.scanner.Peek()
	}

	valueInt, err := strconv.Atoi(value)
	if err != nil {
		log.Fatal(errors.New("failed to convert integer literal"))
	}

	return Token{INTEGER, valueInt, l.scanner.offset - len(value)}
}

func (l *Lexer) handleStringLiteral() Token {
	value := ""
	l.scanner.Consume()
	nextChar := l.scanner.Peek()

	allowedEscapedChars := []string{"t", "n", "\\", string('"')}
	for nextChar != string('"') {
		if nextChar == "\\" {
			value += nextChar
			l.scanner.Consume()
			nextChar = l.scanner.Peek()
			if !slices.Contains(allowedEscapedChars, nextChar) {
				log.Fatal(errors.New("unknown escape sequence"))
			}
		}
		value += nextChar
		l.scanner.Consume()
		nextChar = l.scanner.Peek()
	}

	// consume the closing " without adding it to the value
	// and adjust the offset for the length of the surrounding "" that are not part of the value
	l.scanner.Consume()

	return Token{STRING, value, l.scanner.offset - len(value) - 2}
}

func (l *Lexer) handleEndOfFile() Token {
	// automatically emit a new line when at the end of the last line
	if !l.isNewLine {
		l.isNewLine = true
		// act as if we had consumed a newline token in the scanner to keep the offsets consistent
		l.scanner.offset += 1
		return Token{NEWLINE, nil, l.scanner.offset - 1}
	}
	// emit a dedent token for all remaining indentation levels
	if l.indentStack[len(l.indentStack)-1] > 0 {
		dedentTokenSize := l.indentStack[len(l.indentStack)-1] - l.indentStack[len(l.indentStack)-2]
		_ = dedentTokenSize
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		return Token{DEDENT, nil, l.scanner.offset}
	}
	return Token{EOF, nil, l.scanner.offset}
}
