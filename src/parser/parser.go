// Package parser implements a parser for the chocopy language
package parser

import (
	"chogopy/src/ast"
	"chogopy/src/lexer"
	"slices"
)

var addTokens = []lexer.TokenKind{
	lexer.PLUS,
	lexer.MINUS,
}

var multTokens = []lexer.TokenKind{
	lexer.MUL,
	lexer.DIV,
	lexer.MOD,
}

var compareTokens = []lexer.TokenKind{
	lexer.EQ,
	lexer.NE,
	lexer.LE,
	lexer.GE,
	lexer.LT,
	lexer.GT,
	lexer.IS,
}

var literalTokens = []lexer.TokenKind{
	lexer.NONE,
	lexer.TRUE,
	lexer.FALSE,
	lexer.INTEGER,
	lexer.STRING,
}

var expressionTokens = []lexer.TokenKind{
	lexer.NOT,
	lexer.IDENTIFIER,
	lexer.NONE,
	lexer.TRUE,
	lexer.FALSE,
	lexer.INTEGER,
	lexer.STRING,
	lexer.LSQUAREBRACKET,
	lexer.LROUNDBRACKET,
	lexer.MINUS,
}

var simpleCompoundExpressionTokens = append(literalTokens, []lexer.TokenKind{
	lexer.IDENTIFIER,
	lexer.LSQUAREBRACKET,
	lexer.LROUNDBRACKET,
}...)

var statementTokens = []lexer.TokenKind{
	lexer.PASS,
	lexer.RETURN,
	lexer.IF,
	lexer.WHILE,
	lexer.FOR,
}

type Parser struct {
	lexer *lexer.Lexer
}

func NewParser(lexer *lexer.Lexer) Parser {
	return Parser{
		lexer,
	}
}

func (p *Parser) nextTokenIn(tokenKindSlice []lexer.TokenKind) bool {
	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	return slices.Contains(tokenKindSlice, peekedToken.Kind)
}

func (p *Parser) check(expectedTokenKinds ...lexer.TokenKind) bool {
	peekedTokens := p.lexer.Peek(len(expectedTokenKinds))
	if len(peekedTokens) < len(expectedTokenKinds) {
		return false
	}

	for i, expectedTokenKind := range expectedTokenKinds {
		peekedToken := &peekedTokens[i]
		if expectedTokenKind != peekedToken.Kind {
			return false
		}
	}
	return true
}

func (p *Parser) match(expected lexer.TokenKind) lexer.Token {
	if p.check(expected) {
		token := p.lexer.Consume(false)
		return token
	}

	p.syntaxError(TokenNotFound)
	return lexer.Token{}
}

func (p *Parser) ParseProgram() ast.Program {
	definitions := p.parseDefinitions()
	statements := p.parseStatements()

	p.match(lexer.EOF)

	return ast.Program{
		Definitions: definitions,
		Statements:  statements,
	}
}
