// Package parser implements a parser for the chocopy language
package parser

import "chogopy/pkg/lexer"

type Parser struct {
	lexer *lexer.Lexer
}

func NewParser(lexer *lexer.Lexer) Parser {
	return Parser{
		lexer,
	}
}

func (p *Parser) check(expected []*lexer.Token) bool {
	peekedTokens := p.lexer.Peek(len(expected))
	if len(peekedTokens) < len(expected) {
		return false
	}

	for i, expectedToken := range expected {
		peekedToken := &peekedTokens[i]
		if !expectedToken.Equals(peekedToken) {
			return false
		}
	}
	return true
}

func (p *Parser) match()
