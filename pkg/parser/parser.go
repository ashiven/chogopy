// Package parser implements a parser for the chocopy language
package parser

import (
	"chogopy/pkg/lexer"
	"errors"
	"log"
	"slices"
)

var expressionTokens = lexer.TokenSlice(
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
)

var statementTokens = lexer.TokenSlice(
	lexer.PASS,
	lexer.RETURN,
	lexer.IF,
	lexer.WHILE,
	lexer.FOR,
)

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
		if !expectedToken.KindEquals(peekedToken) {
			return false
		}
	}
	return true
}

func (p *Parser) match(expected []*lexer.Token) lexer.Token {
	if p.check(expected) {
		token := p.lexer.Consume(false)
		return token
	}

	log.Fatal(errors.New("match: expected token"))
	return lexer.Token{}
}

func (p *Parser) ParseProgram() Program {
	definitions := p.parseDefinitions()
	statements := p.parseStatements()

	p.match(lexer.TokenSlice(lexer.EOF))

	return Program{
		Definitions: definitions,
		Statements:  statements,
	}
}

func (p *Parser) parseDefinitions() []Operation {
	definitions := []Operation{}

	for {
		if p.check(lexer.TokenSlice(lexer.IDENTIFIER, lexer.COLON)) {
			varDef := p.parseVarDef()
			definitions = append(definitions, varDef)
			continue
		}
		if p.check(lexer.TokenSlice(lexer.DEF)) {
			funcDef := p.parseFuncDef()
			definitions = append(definitions, funcDef)
			continue
		}
		break
	}

	return definitions
}

func (p *Parser) parseStatements() []Operation {
	statements := []Operation{}

	peekedTokens := p.lexer.Peek(1)
	peekToken := &peekedTokens[0]
	// TODO: this is a problem because the peekToken will actually have a real value and offset
	// while the expressionTokens/statementTokens are only dummies without real values -> slices.Contains always returns false
	for slices.Contains(expressionTokens, peekToken) || slices.Contains(statementTokens, peekToken) {
		statement := p.parseStatement()
		statements = append(statements, statement)
	}

	return statements
}

func (p *Parser) parseVarDef() Operation {
	varNameToken := p.match(lexer.TokenSlice(lexer.IDENTIFIER))
	varName := varNameToken.Value.(string)
	p.match(lexer.TokenSlice(lexer.COLON))
	varType := p.parseType()

	if !p.check(lexer.TokenSlice(lexer.ASSIGN)) {
		// TODO: syntax error
	}
	p.match(lexer.TokenSlice(lexer.ASSIGN))

	literal := p.parseLiteral()

	if !p.check(lexer.TokenSlice(lexer.NEWLINE)) {
		// TODO: syntax error
	}
	p.match(lexer.TokenSlice(lexer.NEWLINE))

	return &varDef{
		TypedVar: &typedVar{
			VarName: varName,
			VarType: varType,
		},
		Literal: literal,
	}
}

func (p *Parser) parseType() Operation {
	if p.check(lexer.TokenSlice(lexer.INT)) {
		p.match(lexer.TokenSlice(lexer.INT))
		return &namedType{
			TypeName: "int",
		}
	} else if p.check(lexer.TokenSlice(lexer.STR)) {
		p.match(lexer.TokenSlice(lexer.STR))
		return &namedType{
			TypeName: "str",
		}
	} else if p.check(lexer.TokenSlice(lexer.BOOL)) {
		p.match(lexer.TokenSlice(lexer.BOOL))
		return &namedType{
			TypeName: "bool",
		}
	} else if p.check(lexer.TokenSlice(lexer.OBJECT)) {
		p.match(lexer.TokenSlice(lexer.OBJECT))
		return &namedType{
			TypeName: "object",
		}
	} else if p.check(lexer.TokenSlice(lexer.LSQUAREBRACKET, lexer.INTEGER, lexer.RSQUAREBRACKET)) {
		p.match(lexer.TokenSlice(lexer.LSQUAREBRACKET))
		elemType := p.parseType()
		p.match(lexer.TokenSlice(lexer.RSQUAREBRACKET))
		return &listType{
			ElemType: elemType,
		}
	} else {
		// TODO: syntax error
		return nil
	}
}

func (p *Parser) parseLiteral() Operation {
	if p.check(lexer.TokenSlice(lexer.NONE)) {
		p.match(lexer.TokenSlice(lexer.NONE))
		return &literalExpr{
			Value: nil,
		}
	} else if p.check(lexer.TokenSlice(lexer.TRUE)) {
		p.match(lexer.TokenSlice(lexer.TRUE))
		return &literalExpr{
			Value: true,
		}
	} else if p.check(lexer.TokenSlice(lexer.FALSE)) {
		p.match(lexer.TokenSlice(lexer.FALSE))
		return &literalExpr{
			Value: false,
		}
	} else if p.check(lexer.TokenSlice(lexer.INTEGER)) {
		integerToken := p.match(lexer.TokenSlice(lexer.INTEGER))
		integerValue := integerToken.Value.(int)
		return &literalExpr{
			Value: integerValue,
		}
	} else if p.check(lexer.TokenSlice(lexer.STRING)) {
		stringToken := p.match(lexer.TokenSlice(lexer.STRING))
		stringValue := stringToken.Value.(string)
		return &literalExpr{
			Value: stringValue,
		}
	} else {
		return nil
		// TODO: error invalid literal
	}
}

func (p *Parser) parseFuncDef() Operation {
	return nil
}

func (p *Parser) parseStatement() Operation {
	return nil
}
