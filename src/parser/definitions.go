package parser

import (
	"chogopy/src/ast"
	"chogopy/src/lexer"
)

func (p *Parser) parseDefinitions() []ast.Node {
	definitions := []ast.Node{}

	for {
		if p.check(lexer.IDENTIFIER, lexer.COLON) {
			varDef := p.parseVarDef()
			definitions = append(definitions, varDef)
			continue
		}
		if p.check(lexer.DEF) {
			funcDef := p.parseFuncDef()
			definitions = append(definitions, funcDef)
			continue
		}
		break
	}

	return definitions
}

func (p *Parser) parseVarDef() ast.Node {
	varNameToken := p.match(lexer.IDENTIFIER)
	varName := varNameToken.Value.(string)
	p.match(lexer.COLON)
	varType := p.parseType()
	p.match(lexer.ASSIGN)
	literal := p.parseLiteral()
	p.match(lexer.NEWLINE)

	return &ast.VarDef{
		TypedVar: &ast.TypedVar{
			VarName: varName,
			VarType: varType,
		},
		Literal: literal,
	}
}

func (p *Parser) parseType() ast.Node {
	if p.check(lexer.INT) {
		p.match(lexer.INT)
		return &ast.NamedType{
			TypeName: "int",
		}
	}

	if p.check(lexer.STR) {
		p.match(lexer.STR)
		return &ast.NamedType{
			TypeName: "str",
		}
	}

	if p.check(lexer.BOOL) {
		p.match(lexer.BOOL)
		return &ast.NamedType{
			TypeName: "bool",
		}
	}

	if p.check(lexer.OBJECT) {
		p.match(lexer.OBJECT)
		return &ast.NamedType{
			TypeName: "object",
		}
	}

	if p.check(lexer.LSQUAREBRACKET) {
		p.match(lexer.LSQUAREBRACKET)
		elemType := p.parseType()
		p.match(lexer.RSQUAREBRACKET)
		return &ast.ListType{
			ElemType: elemType,
		}
	}

	p.syntaxError(UnknownType)
	return nil
}

func (p *Parser) parseLiteral() ast.Node {
	if p.check(lexer.NONE) {
		p.match(lexer.NONE)
		return &ast.LiteralExpr{
			Value: nil,
		}
	}

	if p.check(lexer.TRUE) {
		p.match(lexer.TRUE)
		return &ast.LiteralExpr{
			Value: true,
		}
	}

	if p.check(lexer.FALSE) {
		p.match(lexer.FALSE)
		return &ast.LiteralExpr{
			Value: false,
		}
	}

	if p.check(lexer.INTEGER) {
		integerToken := p.match(lexer.INTEGER)
		integerValue := integerToken.Value.(int)
		return &ast.LiteralExpr{
			Value: integerValue,
		}
	}

	if p.check(lexer.STRING) {
		stringToken := p.match(lexer.STRING)
		stringValue := stringToken.Value.(string)
		return &ast.LiteralExpr{
			Value: stringValue,
		}
	}

	p.syntaxError(TokenNotFound)
	return nil
}

func (p *Parser) parseFuncDef() ast.Node {
	p.match(lexer.DEF)
	functionNameToken := p.match(lexer.IDENTIFIER)
	functionName := functionNameToken.Value.(string)

	p.match(lexer.LROUNDBRACKET)
	parameters := p.parseFuncParams()
	p.match(lexer.RROUNDBRACKET)

	returnType := p.parseFuncReturnType()

	p.match(lexer.COLON)
	p.match(lexer.NEWLINE)
	p.match(lexer.INDENT)

	funcDeclarations := p.parseFuncDeclarations()
	funcStatements := p.parseStatements()
	funcBody := append(funcDeclarations, funcStatements...)

	if p.check(lexer.ASSIGN) {
		p.syntaxError(NoLhsInAssignment)
	}
	if p.check(lexer.INDENT) {
		p.syntaxError(UnexpectedIndentation)
	}
	if len(funcBody) == 0 {
		p.syntaxError(Indentation)
	}

	p.match(lexer.DEDENT)

	return &ast.FuncDef{
		FuncName:   functionName,
		Parameters: parameters,
		FuncBody:   funcBody,
		ReturnType: returnType,
	}
}

func (p *Parser) parseFuncParams() []ast.Node {
	parameters := []ast.Node{}

	paramIndex := 0
	for p.check(lexer.IDENTIFIER) ||
		p.check(lexer.COMMA) {

		if paramIndex > 0 {
			if !p.check(lexer.COMMA) {
				p.syntaxError(CommaExpected)
			}
			p.match(lexer.COMMA)
		}

		varNameToken := p.match(lexer.IDENTIFIER)
		varName := varNameToken.Value.(string)
		p.match(lexer.COLON)
		varType := p.parseType()

		parameter := &ast.TypedVar{VarName: varName, VarType: varType}
		parameters = append(parameters, parameter)
		paramIndex++
	}

	return parameters
}

func (p *Parser) parseFuncReturnType() ast.Node {
	returnNone := &ast.NamedType{TypeName: "<None>"}

	if p.check(lexer.RARROW) {
		p.match(lexer.RARROW)
		returnType := p.parseType()
		return returnType
	}

	return returnNone
}

func (p *Parser) parseFuncDeclarations() []ast.Node {
	funcDeclarations := []ast.Node{}

	if p.check(lexer.NONLOCAL) {
		p.match(lexer.NONLOCAL)
		declNameToken := p.match(lexer.IDENTIFIER)
		declName := declNameToken.Value.(string)
		p.match(lexer.NEWLINE)
		nonLocalDecl := &ast.NonLocalDecl{DeclName: declName}
		funcDeclarations = append(funcDeclarations, nonLocalDecl)
		funcDeclarations = append(funcDeclarations, p.parseFuncDeclarations()...)
	}

	if p.check(lexer.GLOBAL) {
		p.match(lexer.GLOBAL)
		declNameToken := p.match(lexer.IDENTIFIER)
		declName := declNameToken.Value.(string)
		p.match(lexer.NEWLINE)
		globalDecl := &ast.GlobalDecl{DeclName: declName}
		funcDeclarations = append(funcDeclarations, globalDecl)
		funcDeclarations = append(funcDeclarations, p.parseFuncDeclarations()...)
	}

	if p.check(lexer.IDENTIFIER, lexer.COLON) {
		varDef := p.parseVarDef()
		funcDeclarations = append(funcDeclarations, varDef)
		funcDeclarations = append(funcDeclarations, p.parseFuncDeclarations()...)
	}

	return funcDeclarations
}
