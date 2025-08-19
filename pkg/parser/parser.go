// Package parser implements a parser for the chocopy language
package parser

import (
	"chogopy/pkg/lexer"
	"errors"
	"log"
	"slices"
)

var addTokens = lexer.TokenSlice(
	lexer.PLUS,
	lexer.MINUS,
)

var multTokens = lexer.TokenSlice(
	lexer.MUL,
	lexer.DIV,
	lexer.MOD,
)

var compareTokens = lexer.TokenSlice(
	lexer.EQ,
	lexer.NE,
	lexer.LE,
	lexer.GE,
	lexer.LT,
	lexer.GT,
	lexer.IS,
)

var literalTokens = lexer.TokenSlice(
	lexer.NONE,
	lexer.TRUE,
	lexer.FALSE,
	lexer.INTEGER,
	lexer.STRING,
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

	// TODO: syntax error
	// just print a syntax error right here instead of the unnecessary checks before each match
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

	p.match(lexer.TokenSlice(lexer.ASSIGN))

	literal := p.parseLiteral()

	p.match(lexer.TokenSlice(lexer.NEWLINE))

	return &VarDef{
		TypedVar: &TypedVar{
			VarName: varName,
			VarType: varType,
		},
		Literal: literal,
	}
}

func (p *Parser) parseType() Operation {
	if p.check(lexer.TokenSlice(lexer.INT)) {
		p.match(lexer.TokenSlice(lexer.INT))
		return &NamedType{
			TypeName: "int",
		}
	}

	if p.check(lexer.TokenSlice(lexer.STR)) {
		p.match(lexer.TokenSlice(lexer.STR))
		return &NamedType{
			TypeName: "str",
		}
	}

	if p.check(lexer.TokenSlice(lexer.BOOL)) {
		p.match(lexer.TokenSlice(lexer.BOOL))
		return &NamedType{
			TypeName: "bool",
		}
	}

	if p.check(lexer.TokenSlice(lexer.OBJECT)) {
		p.match(lexer.TokenSlice(lexer.OBJECT))
		return &NamedType{
			TypeName: "object",
		}
	}

	if p.check(lexer.TokenSlice(lexer.LSQUAREBRACKET, lexer.INTEGER, lexer.RSQUAREBRACKET)) {
		p.match(lexer.TokenSlice(lexer.LSQUAREBRACKET))
		elemType := p.parseType()
		p.match(lexer.TokenSlice(lexer.RSQUAREBRACKET))
		return &ListType{
			ElemType: elemType,
		}
	}

	// TODO: syntax error
	return nil
}

func (p *Parser) parseLiteral() Operation {
	if p.check(lexer.TokenSlice(lexer.NONE)) {
		p.match(lexer.TokenSlice(lexer.NONE))
		return &LiteralExpr{
			Value: nil,
		}
	}

	if p.check(lexer.TokenSlice(lexer.TRUE)) {
		p.match(lexer.TokenSlice(lexer.TRUE))
		return &LiteralExpr{
			Value: true,
		}
	}

	if p.check(lexer.TokenSlice(lexer.FALSE)) {
		p.match(lexer.TokenSlice(lexer.FALSE))
		return &LiteralExpr{
			Value: false,
		}
	}

	if p.check(lexer.TokenSlice(lexer.INTEGER)) {
		integerToken := p.match(lexer.TokenSlice(lexer.INTEGER))
		integerValue := integerToken.Value.(int)
		return &LiteralExpr{
			Value: integerValue,
		}
	}

	if p.check(lexer.TokenSlice(lexer.STRING)) {
		stringToken := p.match(lexer.TokenSlice(lexer.STRING))
		stringValue := stringToken.Value.(string)
		return &LiteralExpr{
			Value: stringValue,
		}
	}

	// TODO: error invalid literal
	return nil
}

func (p *Parser) parseFuncDef() Operation {
	p.match(lexer.TokenSlice(lexer.DEF))
	functionNameToken := p.match(lexer.TokenSlice(lexer.IDENTIFIER))
	functionName := functionNameToken.Value.(string)

	p.match(lexer.TokenSlice(lexer.LROUNDBRACKET))

	parameters := p.parseFuncParams()

	p.match(lexer.TokenSlice(lexer.RROUNDBRACKET))

	returnType := p.parseFuncReturnType()

	p.match(lexer.TokenSlice(lexer.COLON))
	p.match(lexer.TokenSlice(lexer.NEWLINE))
	p.match(lexer.TokenSlice(lexer.INDENT))

	funcDeclarations := p.parseFuncDeclarations()
	funcStatements := p.parseStatements()
	funcBody := append(funcDeclarations, funcStatements...)

	// TODO: multiple syntax errors

	p.match(lexer.TokenSlice(lexer.DEDENT))

	return &FuncDef{
		FuncName:   functionName,
		Parameters: parameters,
		FuncBody:   funcBody,
		ReturnType: returnType,
	}
}

func (p *Parser) parseFuncParams() []Operation {
	parameters := []Operation{}

	paramIndex := 0
	for p.check(lexer.TokenSlice(lexer.IDENTIFIER)) {

		if p.check(lexer.TokenSlice(lexer.RROUNDBRACKET)) ||
			p.check(lexer.TokenSlice(lexer.NEWLINE)) {
			break
		}

		if paramIndex > 0 && !p.check(lexer.TokenSlice(lexer.COMMA)) {
			// TODO: syntax error
		}

		varNameToken := p.match(lexer.TokenSlice(lexer.IDENTIFIER))
		varName := varNameToken.Value.(string)

		p.match(lexer.TokenSlice(lexer.COLON))

		varType := p.parseType()

		parameter := &TypedVar{VarName: varName, VarType: varType}
		parameters = append(parameters, parameter)
		paramIndex++
	}

	return parameters
}

func (p *Parser) parseFuncReturnType() Operation {
	returnNone := &NamedType{TypeName: "<None>"}

	if p.check(lexer.TokenSlice(lexer.RARROW)) {
		p.match(lexer.TokenSlice(lexer.RARROW))
		returnType := p.parseType()
		return returnType
	}

	return returnNone
}

func (p *Parser) parseFuncDeclarations() []Operation {
	funcDeclarations := []Operation{}

	if p.check(lexer.TokenSlice(lexer.NONLOCAL)) {
		p.match(lexer.TokenSlice(lexer.NONLOCAL))

		declNameToken := p.match(lexer.TokenSlice(lexer.IDENTIFIER))
		declName := declNameToken.Value.(string)

		p.match(lexer.TokenSlice(lexer.NEWLINE))

		nonLocalDecl := &NonLocalDecl{DeclName: declName}
		funcDeclarations = append(funcDeclarations, nonLocalDecl)
		funcDeclarations = append(funcDeclarations, p.parseFuncDeclarations()...)
	}

	if p.check(lexer.TokenSlice(lexer.GLOBAL)) {
		p.match(lexer.TokenSlice(lexer.GLOBAL))

		declNameToken := p.match(lexer.TokenSlice(lexer.IDENTIFIER))
		declName := declNameToken.Value.(string)

		p.match(lexer.TokenSlice(lexer.NEWLINE))

		globalDecl := &GlobalDecl{DeclName: declName}
		funcDeclarations = append(funcDeclarations, globalDecl)
		funcDeclarations = append(funcDeclarations, p.parseFuncDeclarations()...)
	}

	if p.check(lexer.TokenSlice(lexer.IDENTIFIER, lexer.COLON)) {
		varDef := p.parseVarDef()
		funcDeclarations = append(funcDeclarations, varDef)
		funcDeclarations = append(funcDeclarations, p.parseFuncDeclarations()...)
	}

	return funcDeclarations
}

func (p *Parser) parseStatement() Operation {
	peekedTokens := p.lexer.Peek(1)
	peekToken := &peekedTokens[0]

	if slices.Contains(expressionTokens, peekToken) ||
		p.check(lexer.TokenSlice(lexer.PASS)) ||
		p.check(lexer.TokenSlice(lexer.RETURN)) {
		simpleStatement := p.parseSimpleStatement()

		p.match(lexer.TokenSlice(lexer.NEWLINE))

		return simpleStatement
	}

	if p.check(lexer.TokenSlice(lexer.IF)) {
		p.match(lexer.TokenSlice(lexer.IF))

		condition := p.parseExpression(false, false)
		// TODO: don't check for cond nil here but rather error in parseExpr with syntax err if expr is nil

		p.match(lexer.TokenSlice(lexer.COLON))
		p.match(lexer.TokenSlice(lexer.NEWLINE))
		p.match(lexer.TokenSlice(lexer.INDENT))

		ifBody := p.parseStatements()
		if len(ifBody) == 0 {
			// TODO: syntax error
		}
		elseBody := p.parseElseBody()

		p.match(lexer.TokenSlice(lexer.DEDENT))

		return &IfStmt{Condition: condition, IfBody: ifBody, ElseBody: elseBody}
	}

	if p.check(lexer.TokenSlice(lexer.WHILE)) {
		p.match(lexer.TokenSlice(lexer.WHILE))

		condition := p.parseExpression(false, false)
		// TODO: don't check for cond nil here but rather error in parseExpr with syntax err if expr is nil

		p.match(lexer.TokenSlice(lexer.COLON))
		p.match(lexer.TokenSlice(lexer.NEWLINE))
		p.match(lexer.TokenSlice(lexer.INDENT))

		body := p.parseStatements()
		if len(body) == 0 {
			// TODO: syntax error
		}

		p.match(lexer.TokenSlice(lexer.DEDENT))

		return &WhileStmt{Condition: condition, Body: body}
	}

	if p.check(lexer.TokenSlice(lexer.FOR)) {
		p.match(lexer.TokenSlice(lexer.FOR))

		iterNameToken := p.match(lexer.TokenSlice(lexer.IDENTIFIER))
		iterName := iterNameToken.Value.(string)

		p.match(lexer.TokenSlice(lexer.IN))

		iter := p.parseExpression(false, false)

		p.match(lexer.TokenSlice(lexer.COLON))
		p.match(lexer.TokenSlice(lexer.NEWLINE))
		p.match(lexer.TokenSlice(lexer.INDENT))

		body := p.parseStatements()
		if len(body) == 0 {
			// TODO: syntax error
		}

		p.match(lexer.TokenSlice(lexer.DEDENT))

		return &ForStmt{IterName: iterName, Iter: iter, Body: body}
	}

	// TODO: error invalid stmt
	return nil
}

func (p *Parser) parseElseBody() []Operation {
	elseBody := []Operation{}

	if p.check(lexer.TokenSlice(lexer.ELIF)) {
		p.match(lexer.TokenSlice(lexer.ELIF))

		condition := p.parseExpression(false, false)
		// TODO: handle no expression inside parseExpression

		p.match(lexer.TokenSlice(lexer.COLON))
		p.match(lexer.TokenSlice(lexer.NEWLINE))
		p.match(lexer.TokenSlice(lexer.INDENT))

		elifIfBody := p.parseStatements()
		if len(elifIfBody) == 0 {
			// TODO: syntax error
		}
		elifElseBody := p.parseElseBody()

		elif := &IfStmt{Condition: condition, IfBody: elifIfBody, ElseBody: elifElseBody}

		elseBody = append(elseBody, elif)
		return elseBody
	}

	if p.check(lexer.TokenSlice(lexer.ELSE)) {
		p.match(lexer.TokenSlice(lexer.ELSE))
		p.match(lexer.TokenSlice(lexer.COLON))
		p.match(lexer.TokenSlice(lexer.NEWLINE))
		p.match(lexer.TokenSlice(lexer.INDENT))

		return p.parseStatements()
	}

	return elseBody
}

func (p *Parser) parseSimpleStatement() Operation {
	if p.check(lexer.TokenSlice(lexer.IDENTIFIER, lexer.COLON)) {
		// TODO: syntax error: variable defined later
	}

	if p.check(lexer.TokenSlice(lexer.PASS)) {
		p.match(lexer.TokenSlice(lexer.PASS))
		return &PassStmt{}
	}

	if p.check(lexer.TokenSlice(lexer.RETURN)) {
		p.match(lexer.TokenSlice(lexer.RETURN))

		peekedTokens := p.lexer.Peek(1)
		peekedToken := &peekedTokens[0]

		// TODO: check if this is correct
		var returnVal Operation
		if slices.Contains(expressionTokens, peekedToken) {
			returnVal = p.parseExpression(false, false)
		}

		return &ReturnStmt{ReturnVal: returnVal}
	}

	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]
	if slices.Contains(expressionTokens, peekedToken) {
		return p.parseExpressionAssignList()
	}

	// TODO: raise error invalid simple stmt
	return nil
}

func (p *Parser) parseExpressionAssignList() Operation {
	expression := p.parseExpression(false, false)

	if p.check(lexer.TokenSlice(lexer.ASSIGN)) {
		p.match(lexer.TokenSlice(lexer.ASSIGN))
		return &AssignStmt{Target: expression, Value: p.parseExpressionAssignList()}
	}

	return expression
}

func (p *Parser) parseExpression(insideAnd bool, insideOr bool) Operation {
	var expression Operation

	if p.check(lexer.TokenSlice(lexer.NOT)) {
		expression = p.parseNotExpression()
	}

	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]
	if slices.Contains(expressionTokens, peekedToken) {
		expression = p.parseCompoundExpression(false, false, false, false)
	}

	if p.check(lexer.TokenSlice(lexer.AND)) && !insideAnd {
		expression = p.parseAndExpression(expression)
	}

	if p.check(lexer.TokenSlice(lexer.OR)) && !insideAnd && !insideOr {
		expression = p.parseOrExpression(expression)
	}

	if p.check(lexer.TokenSlice(lexer.IF)) && !insideAnd && !insideOr {
		p.match(lexer.TokenSlice(lexer.IF))
		condition := p.parseExpression(false, false)
		p.match(lexer.TokenSlice(lexer.ELSE))
		elseOp := p.parseExpression(false, false)
		return &IfExpr{Condition: condition, IfOp: expression, ElseOp: elseOp}
	}

	if expression == nil {
		// TODO: syntax error no valid expression found
	}
	return expression
}

func (p *Parser) parseNotExpression() Operation {
	p.match(lexer.TokenSlice(lexer.NOT))

	if p.check(lexer.TokenSlice(lexer.NOT)) {
		return &UnaryExpr{Op: "not", Value: p.parseNotExpression()}
	}

	return &UnaryExpr{Op: "not", Value: p.parseCompoundExpression(false, false, false, false)}
}

func (p *Parser) parseAndExpression(expression Operation) Operation {
	if p.check(lexer.TokenSlice(lexer.AND)) {
		p.match(lexer.TokenSlice(lexer.AND))
		rhs := p.parseExpression(true, false)
		andExpression := &BinaryExpr{Op: "and", Lhs: expression, Rhs: rhs}
		return p.parseAndExpression(andExpression)
	}

	return expression
}

func (p *Parser) parseOrExpression(expression Operation) Operation {
	if p.check(lexer.TokenSlice(lexer.OR)) {
		p.match(lexer.TokenSlice(lexer.OR))
		rhs := p.parseExpression(false, true)
		orExpression := &BinaryExpr{Op: "or", Lhs: expression, Rhs: rhs}
		return p.parseOrExpression(orExpression)
	}

	return expression
}

func (p *Parser) parseCompoundExpression(insideNegation bool, insideMult bool, insideAdd bool, insideCompare bool) Operation {
	var compoundExpression Operation

	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	if slices.Contains(literalTokens, peekedToken) {
		compoundExpression = p.parseLiteral()
	}

	if p.check(lexer.TokenSlice(lexer.IDENTIFIER, lexer.LROUNDBRACKET)) {
		funcNameToken := p.match(lexer.TokenSlice(lexer.IDENTIFIER))
		funcName := funcNameToken.Value.(string)
		p.match(lexer.TokenSlice(lexer.LROUNDBRACKET))
		arguments := p.parseExpressionList()
		p.match(lexer.TokenSlice(lexer.RROUNDBRACKET))
		compoundExpression = &CallExpr{FuncName: funcName, Arguments: arguments}
	}

	if p.check(lexer.TokenSlice(lexer.IDENTIFIER)) {
		identifierToken := p.match(lexer.TokenSlice(lexer.IDENTIFIER))
		identifier := identifierToken.Value.(string)
		compoundExpression = &IdentExpr{Identifier: identifier}
	}

	if p.check(lexer.TokenSlice(lexer.LSQUAREBRACKET)) {
		p.match(lexer.TokenSlice(lexer.LSQUAREBRACKET))
		elements := p.parseExpressionList()
		p.match(lexer.TokenSlice(lexer.RSQUAREBRACKET))
		compoundExpression = &ListExpr{Elements: elements}
	}

	if p.check(lexer.TokenSlice(lexer.LROUNDBRACKET)) {
		p.match(lexer.TokenSlice(lexer.LROUNDBRACKET))
		expression := p.parseExpression(false, false)
		p.match(lexer.TokenSlice(lexer.RROUNDBRACKET))
		compoundExpression = expression
	}

	if p.check(lexer.TokenSlice(lexer.MINUS)) {
		compoundExpression = p.parseUnaryNegation()
	}

	if p.check(lexer.TokenSlice(lexer.LSQUAREBRACKET)) {
		compoundExpression = p.parseIndexExpression(compoundExpression)
	}

	if slices.Contains(multTokens, peekedToken) && !insideNegation && !insideMult {
		compoundExpression = p.parseMultExpression(compoundExpression)
	}

	if slices.Contains(addTokens, peekedToken) && !insideNegation && !insideMult && !insideAdd {
		compoundExpression = p.parseAddExpression(compoundExpression)
	}

	if slices.Contains(compareTokens, peekedToken) && !insideNegation && !insideMult && !insideAdd && !insideCompare {
		compoundExpression = p.parseCompareExpression(compoundExpression)
	}

	if compoundExpression == nil {
		// TODO: syntax error
	}
	return compoundExpression
}

func (p *Parser) parseExpressionList() []Operation {
	expressionList := []Operation{}

	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	if slices.Contains(expressionTokens, peekedToken) {
		expressionList = append(expressionList, p.parseExpression(false, false))

		for !p.check(lexer.TokenSlice(lexer.RROUNDBRACKET)) &&
			!p.check(lexer.TokenSlice(lexer.RSQUAREBRACKET)) &&
			!p.check(lexer.TokenSlice(lexer.NEWLINE)) {
			p.match(lexer.TokenSlice(lexer.COMMA))
			expressionList = append(expressionList, p.parseExpression(false, false))
		}
	}

	return expressionList
}

func (p *Parser) parseUnaryNegation() Operation {
	p.match(lexer.TokenSlice(lexer.MINUS))

	if p.check(lexer.TokenSlice(lexer.MINUS)) {
		return &UnaryExpr{Op: "-", Value: p.parseUnaryNegation()}
	}

	return &UnaryExpr{Op: "-", Value: p.parseCompoundExpression(true, false, false, false)}
}

func (p *Parser) parseIndexExpression(compoundExpression Operation) Operation {
	p.match(lexer.TokenSlice(lexer.LSQUAREBRACKET))
	index := p.parseExpression(false, false)
	p.match(lexer.TokenSlice(lexer.RSQUAREBRACKET))

	indexExpression := &IndexExpr{Value: compoundExpression, Index: index}
	for p.check(lexer.TokenSlice(lexer.LSQUAREBRACKET)) {
		p.match(lexer.TokenSlice(lexer.LSQUAREBRACKET))
		index = p.parseExpression(false, false)
		p.match(lexer.TokenSlice(lexer.RSQUAREBRACKET))
		indexExpression = &IndexExpr{Value: indexExpression, Index: index}
	}

	return indexExpression
}

func (p *Parser) parseMultExpression(compoundExpression Operation) Operation {
	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	if slices.Contains(multTokens, peekedToken) {
		op := peekedToken.Value.(string)
		rhs := p.parseCompoundExpression(false, true, false, false)
		return &BinaryExpr{Op: op, Lhs: compoundExpression, Rhs: rhs}
	}

	return compoundExpression
}

func (p *Parser) parseAddExpression(compoundExpression Operation) Operation {
	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	if slices.Contains(addTokens, peekedToken) {
		op := peekedToken.Value.(string)
		rhs := p.parseCompoundExpression(false, false, true, false)
		return &BinaryExpr{Op: op, Lhs: compoundExpression, Rhs: rhs}
	}

	return compoundExpression
}

func (p *Parser) parseCompareExpression(compoundExpression Operation) Operation {
	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	op := peekedToken.Value.(string)
	rhs := p.parseCompoundExpression(false, false, false, true)

	// Comparison binary operations can not be nested
	peekedTokens = p.lexer.Peek(1)
	peekedToken = &peekedTokens[0]
	if slices.Contains(compareTokens, peekedToken) {
		// TODO: syntax error comparison not associative
	}

	return &BinaryExpr{Op: op, Lhs: compoundExpression, Rhs: rhs}
}
