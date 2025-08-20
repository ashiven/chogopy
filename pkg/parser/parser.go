// Package parser implements a parser for the chocopy language
package parser

import (
	"chogopy/pkg/lexer"
	"fmt"
	"os"
	"slices"
	"strings"
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

var statementTokens = []lexer.TokenKind{
	lexer.PASS,
	lexer.RETURN,
	lexer.IF,
	lexer.WHILE,
	lexer.FOR,
}

type SyntaxErrorKind int

const (
	CommaExpected SyntaxErrorKind = iota
	ComparisonNotAssociative
	ExpectedExpression
	Indentation
	NoLhsInAssignment
	TokenNotFound
	UnexpectedIndentation
	UnknownType
	UnmatchedParantheses
	VariableDefinedLater
)

type Parser struct {
	lexer *lexer.Lexer
}

func NewParser(lexer *lexer.Lexer) Parser {
	return Parser{
		lexer,
	}
}

func (p *Parser) syntaxError(errorKind SyntaxErrorKind) {
	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]
	locationInfo := p.lexer.GetLocation(peekedToken)

	switch errorKind {
	case CommaExpected:
		fmt.Printf("SyntaxError (line %d, column %d): Comma expected.\n", locationInfo.Line, locationInfo.Column)
	case ComparisonNotAssociative:
		fmt.Printf("SyntaxError (line %d, column %d): Comparison operators are not associative.\n", locationInfo.Line, locationInfo.Column)
	case ExpectedExpression:
		fmt.Printf("SyntaxError (line %d, column %d): Expected Expression.\n", locationInfo.Line, locationInfo.Column)
	case Indentation:
		fmt.Printf("SyntaxError (line %d, column %d): Expected at least one indented statement.\n", locationInfo.Line, locationInfo.Column)
	case NoLhsInAssignment:
		fmt.Printf("SyntaxError (line %d, column %d): No left-hand side in assign statement.\n", locationInfo.Line, locationInfo.Column)
	case TokenNotFound:
		fmt.Printf("SyntaxError (line %d, column %d): Expected token not found.\n", locationInfo.Line, locationInfo.Column)
	case UnexpectedIndentation:
		fmt.Printf("SyntaxError (line %d, column %d): Unexpected indentation.\n", locationInfo.Line, locationInfo.Column)
	case UnknownType:
		fmt.Printf("SyntaxError (line %d, column %d): Unknown type.\n", locationInfo.Line, locationInfo.Column)
	case UnmatchedParantheses:
		fmt.Printf("SyntaxError (line %d, column %d): Unmatched ')'.\n", locationInfo.Line, locationInfo.Column)
	case VariableDefinedLater:
		fmt.Printf("SyntaxError (line %d, column %d): Variable declaration after non-declaration statement.\n", locationInfo.Line, locationInfo.Column)
	}

	fmt.Printf(">>>%s\n", locationInfo.LineLiteral)
	fmt.Println(">>>" + strings.Repeat("-", locationInfo.Column-1) + "^")
	os.Exit(0)
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

func (p *Parser) ParseProgram() Program {
	definitions := p.parseDefinitions()
	statements := p.parseStatements()

	p.match(lexer.EOF)

	return Program{
		Definitions: definitions,
		Statements:  statements,
	}
}

func (p *Parser) parseDefinitions() []Operation {
	definitions := []Operation{}

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

func (p *Parser) parseStatements() []Operation {
	statements := []Operation{}

	for p.nextTokenIn(expressionTokens) || p.nextTokenIn(statementTokens) {
		statement := p.parseStatement()
		statements = append(statements, statement)
	}

	return statements
}

func (p *Parser) parseVarDef() Operation {
	varNameToken := p.match(lexer.IDENTIFIER)
	varName := varNameToken.Value.(string)
	p.match(lexer.COLON)
	varType := p.parseType()
	p.match(lexer.ASSIGN)
	literal := p.parseLiteral()
	p.match(lexer.NEWLINE)

	return &VarDef{
		TypedVar: &TypedVar{
			VarName: varName,
			VarType: varType,
		},
		Literal: literal,
	}
}

func (p *Parser) parseType() Operation {
	if p.check(lexer.INT) {
		p.match(lexer.INT)
		return &NamedType{
			TypeName: "int",
		}
	}

	if p.check(lexer.STR) {
		p.match(lexer.STR)
		return &NamedType{
			TypeName: "str",
		}
	}

	if p.check(lexer.BOOL) {
		p.match(lexer.BOOL)
		return &NamedType{
			TypeName: "bool",
		}
	}

	if p.check(lexer.OBJECT) {
		p.match(lexer.OBJECT)
		return &NamedType{
			TypeName: "object",
		}
	}

	if p.check(lexer.LSQUAREBRACKET, lexer.INTEGER, lexer.RSQUAREBRACKET) {
		p.match(lexer.LSQUAREBRACKET)
		elemType := p.parseType()
		p.match(lexer.RSQUAREBRACKET)
		return &ListType{
			ElemType: elemType,
		}
	}

	p.syntaxError(UnknownType)
	return nil
}

func (p *Parser) parseLiteral() Operation {
	if p.check(lexer.NONE) {
		p.match(lexer.NONE)
		return &LiteralExpr{
			Value: nil,
		}
	}

	if p.check(lexer.TRUE) {
		p.match(lexer.TRUE)
		return &LiteralExpr{
			Value: true,
		}
	}

	if p.check(lexer.FALSE) {
		p.match(lexer.FALSE)
		return &LiteralExpr{
			Value: false,
		}
	}

	if p.check(lexer.INTEGER) {
		integerToken := p.match(lexer.INTEGER)
		integerValue := integerToken.Value.(int)
		return &LiteralExpr{
			Value: integerValue,
		}
	}

	if p.check(lexer.STRING) {
		stringToken := p.match(lexer.STRING)
		stringValue := stringToken.Value.(string)
		return &LiteralExpr{
			Value: stringValue,
		}
	}

	p.syntaxError(TokenNotFound)
	return nil
}

func (p *Parser) parseFuncDef() Operation {
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
	for p.check(lexer.IDENTIFIER) {

		if p.check(lexer.RROUNDBRACKET) ||
			p.check(lexer.NEWLINE) {
			break
		}

		if paramIndex > 0 && !p.check(lexer.COMMA) {
			p.syntaxError(CommaExpected)
		}

		varNameToken := p.match(lexer.IDENTIFIER)
		varName := varNameToken.Value.(string)
		p.match(lexer.COLON)
		varType := p.parseType()

		parameter := &TypedVar{VarName: varName, VarType: varType}
		parameters = append(parameters, parameter)
		paramIndex++
	}

	return parameters
}

func (p *Parser) parseFuncReturnType() Operation {
	returnNone := &NamedType{TypeName: "<None>"}

	if p.check(lexer.RARROW) {
		p.match(lexer.RARROW)
		returnType := p.parseType()
		return returnType
	}

	return returnNone
}

func (p *Parser) parseFuncDeclarations() []Operation {
	funcDeclarations := []Operation{}

	if p.check(lexer.NONLOCAL) {
		p.match(lexer.NONLOCAL)
		declNameToken := p.match(lexer.IDENTIFIER)
		declName := declNameToken.Value.(string)
		p.match(lexer.NEWLINE)
		nonLocalDecl := &NonLocalDecl{DeclName: declName}
		funcDeclarations = append(funcDeclarations, nonLocalDecl)
		funcDeclarations = append(funcDeclarations, p.parseFuncDeclarations()...)
	}

	if p.check(lexer.GLOBAL) {
		p.match(lexer.GLOBAL)
		declNameToken := p.match(lexer.IDENTIFIER)
		declName := declNameToken.Value.(string)
		p.match(lexer.NEWLINE)
		globalDecl := &GlobalDecl{DeclName: declName}
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

func (p *Parser) parseStatement() Operation {
	if p.nextTokenIn(expressionTokens) ||
		p.check(lexer.PASS) ||
		p.check(lexer.RETURN) {
		simpleStatement := p.parseSimpleStatement()
		p.match(lexer.NEWLINE)
		return simpleStatement
	}

	if p.check(lexer.IF) {
		p.match(lexer.IF)
		condition := p.parseExpression(false, false)
		p.match(lexer.COLON)
		p.match(lexer.NEWLINE)
		p.match(lexer.INDENT)
		ifBody := p.parseStatements()
		if len(ifBody) == 0 {
			p.syntaxError(Indentation)
		}
		if p.check(lexer.INDENT) {
			p.syntaxError(UnexpectedIndentation)
		}
		elseBody := p.parseElseBody()
		p.match(lexer.DEDENT)
		return &IfStmt{Condition: condition, IfBody: ifBody, ElseBody: elseBody}
	}

	if p.check(lexer.WHILE) {
		p.match(lexer.WHILE)
		condition := p.parseExpression(false, false)
		p.match(lexer.COLON)
		p.match(lexer.NEWLINE)
		p.match(lexer.INDENT)
		body := p.parseStatements()
		if len(body) == 0 {
			p.syntaxError(Indentation)
		}
		if p.check(lexer.INDENT) {
			p.syntaxError(UnexpectedIndentation)
		}
		p.match(lexer.DEDENT)
		return &WhileStmt{Condition: condition, Body: body}
	}

	if p.check(lexer.FOR) {
		p.match(lexer.FOR)
		iterNameToken := p.match(lexer.IDENTIFIER)
		iterName := iterNameToken.Value.(string)
		p.match(lexer.IN)
		iter := p.parseExpression(false, false)
		p.match(lexer.COLON)
		p.match(lexer.NEWLINE)
		p.match(lexer.INDENT)
		body := p.parseStatements()
		if len(body) == 0 {
			p.syntaxError(Indentation)
		}
		if p.check(lexer.INDENT) {
			p.syntaxError(UnexpectedIndentation)
		}
		p.match(lexer.DEDENT)
		return &ForStmt{IterName: iterName, Iter: iter, Body: body}
	}

	p.syntaxError(TokenNotFound)
	return nil
}

func (p *Parser) parseElseBody() []Operation {
	elseBody := []Operation{}

	if p.check(lexer.ELIF) {
		p.match(lexer.ELIF)

		condition := p.parseExpression(false, false)

		p.match(lexer.COLON)
		p.match(lexer.NEWLINE)
		p.match(lexer.INDENT)

		elifIfBody := p.parseStatements()
		if len(elifIfBody) == 0 {
			p.syntaxError(Indentation)
		}
		if p.check(lexer.INDENT) {
			p.syntaxError(UnexpectedIndentation)
		}
		elifElseBody := p.parseElseBody()

		elif := &IfStmt{Condition: condition, IfBody: elifIfBody, ElseBody: elifElseBody}

		elseBody = append(elseBody, elif)
		return elseBody
	}

	if p.check(lexer.ELSE) {
		p.match(lexer.ELSE)
		p.match(lexer.COLON)
		p.match(lexer.NEWLINE)
		p.match(lexer.INDENT)

		return p.parseStatements()
	}

	return elseBody
}

func (p *Parser) parseSimpleStatement() Operation {
	if p.check(lexer.IDENTIFIER, lexer.COLON) {
		p.syntaxError(VariableDefinedLater)
	}

	if p.check(lexer.PASS) {
		p.match(lexer.PASS)
		return &PassStmt{}
	}

	if p.check(lexer.RETURN) {
		p.match(lexer.RETURN)
		var returnVal Operation
		if p.nextTokenIn(expressionTokens) {
			returnVal = p.parseExpression(false, false)
		}
		return &ReturnStmt{ReturnVal: returnVal}
	}

	if p.nextTokenIn(expressionTokens) {
		return p.parseExpressionAssignList()
	}

	p.syntaxError(TokenNotFound)
	return nil
}

func (p *Parser) parseExpressionAssignList() Operation {
	expression := p.parseExpression(false, false)

	if p.check(lexer.ASSIGN) {
		p.match(lexer.ASSIGN)
		return &AssignStmt{Target: expression, Value: p.parseExpressionAssignList()}
	}

	return expression
}

func (p *Parser) parseExpression(insideAnd bool, insideOr bool) Operation {
	var expression Operation

	if p.check(lexer.NOT) {
		expression = p.parseNotExpression()
	}

	if p.nextTokenIn(expressionTokens) {
		expression = p.parseCompoundExpression(false, false, false, false)
	}

	if p.check(lexer.AND) && !insideAnd {
		expression = p.parseAndExpression(expression)
	}

	if p.check(lexer.OR) && !insideAnd && !insideOr {
		expression = p.parseOrExpression(expression)
	}

	if p.check(lexer.IF) && !insideAnd && !insideOr {
		p.match(lexer.IF)
		condition := p.parseExpression(false, false)
		p.match(lexer.ELSE)
		elseOp := p.parseExpression(false, false)
		return &IfExpr{Condition: condition, IfOp: expression, ElseOp: elseOp}
	}

	if expression == nil {
		p.syntaxError(TokenNotFound)
	}
	return expression
}

func (p *Parser) parseNotExpression() Operation {
	p.match(lexer.NOT)

	if p.check(lexer.NOT) {
		return &UnaryExpr{Op: "not", Value: p.parseNotExpression()}
	}

	return &UnaryExpr{Op: "not", Value: p.parseCompoundExpression(false, false, false, false)}
}

func (p *Parser) parseAndExpression(expression Operation) Operation {
	if p.check(lexer.AND) {
		p.match(lexer.AND)
		rhs := p.parseExpression(true, false)
		andExpression := &BinaryExpr{Op: "and", Lhs: expression, Rhs: rhs}
		return p.parseAndExpression(andExpression)
	}

	return expression
}

func (p *Parser) parseOrExpression(expression Operation) Operation {
	if p.check(lexer.OR) {
		p.match(lexer.OR)
		rhs := p.parseExpression(false, true)
		orExpression := &BinaryExpr{Op: "or", Lhs: expression, Rhs: rhs}
		return p.parseOrExpression(orExpression)
	}

	return expression
}

func (p *Parser) parseCompoundExpression(insideNegation bool, insideMult bool, insideAdd bool, insideCompare bool) Operation {
	var compoundExpression Operation

	if p.check(lexer.MINUS) {
		compoundExpression = p.parseUnaryNegation()
	}

	if p.nextTokenIn(literalTokens) {
		compoundExpression = p.parseLiteral()
	}

	if p.check(lexer.IDENTIFIER, lexer.LROUNDBRACKET) {
		funcNameToken := p.match(lexer.IDENTIFIER)
		funcName := funcNameToken.Value.(string)
		p.match(lexer.LROUNDBRACKET)
		arguments := p.parseExpressionList()
		p.match(lexer.RROUNDBRACKET)
		compoundExpression = &CallExpr{FuncName: funcName, Arguments: arguments}
	}

	if p.check(lexer.IDENTIFIER) {
		identifierToken := p.match(lexer.IDENTIFIER)
		identifier := identifierToken.Value.(string)
		compoundExpression = &IdentExpr{Identifier: identifier}
	}

	if p.check(lexer.LSQUAREBRACKET) {
		p.match(lexer.LSQUAREBRACKET)
		elements := p.parseExpressionList()
		p.match(lexer.RSQUAREBRACKET)
		compoundExpression = &ListExpr{Elements: elements}
	}

	if p.check(lexer.LROUNDBRACKET) {
		p.match(lexer.LROUNDBRACKET)
		expression := p.parseExpression(false, false)
		p.match(lexer.RROUNDBRACKET)
		compoundExpression = expression
	}

	if p.check(lexer.LSQUAREBRACKET) {
		compoundExpression = p.parseIndexExpression(compoundExpression)
	}

	if p.nextTokenIn(multTokens) && !insideNegation && !insideMult {
		compoundExpression = p.parseMultExpression(compoundExpression)
	}

	if p.nextTokenIn(addTokens) && !insideNegation && !insideMult && !insideAdd {
		compoundExpression = p.parseAddExpression(compoundExpression)
	}

	if p.nextTokenIn(compareTokens) && !insideNegation && !insideMult && !insideAdd && !insideCompare {
		compoundExpression = p.parseCompareExpression(compoundExpression)
	}

	if compoundExpression == nil {
		p.syntaxError(TokenNotFound)
	}
	return compoundExpression
}

func (p *Parser) parseExpressionList() []Operation {
	expressionList := []Operation{}

	if p.nextTokenIn(expressionTokens) {
		expressionList = append(expressionList, p.parseExpression(false, false))

		for !p.check(lexer.RROUNDBRACKET) &&
			!p.check(lexer.RSQUAREBRACKET) &&
			!p.check(lexer.NEWLINE) {
			p.match(lexer.COMMA)
			expressionList = append(expressionList, p.parseExpression(false, false))
		}
	}

	return expressionList
}

func (p *Parser) parseUnaryNegation() Operation {
	p.match(lexer.MINUS)

	if p.check(lexer.MINUS) {
		return &UnaryExpr{Op: "-", Value: p.parseUnaryNegation()}
	}

	return &UnaryExpr{Op: "-", Value: p.parseCompoundExpression(true, false, false, false)}
}

func (p *Parser) parseIndexExpression(compoundExpression Operation) Operation {
	p.match(lexer.LSQUAREBRACKET)
	index := p.parseExpression(false, false)
	p.match(lexer.RSQUAREBRACKET)

	indexExpression := &IndexExpr{Value: compoundExpression, Index: index}
	for p.check(lexer.LSQUAREBRACKET) {
		p.match(lexer.LSQUAREBRACKET)
		index = p.parseExpression(false, false)
		p.match(lexer.RSQUAREBRACKET)
		indexExpression = &IndexExpr{Value: indexExpression, Index: index}
	}

	return indexExpression
}

func (p *Parser) parseMultExpression(compoundExpression Operation) Operation {
	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	if slices.Contains(multTokens, peekedToken.Kind) {
		op := peekedToken.Value.(string)
		p.match(peekedToken.Kind)
		rhs := p.parseCompoundExpression(false, true, false, false)

		multExpr := &BinaryExpr{Op: op, Lhs: compoundExpression, Rhs: rhs}
		return p.parseMultExpression(multExpr)
	}

	return compoundExpression
}

func (p *Parser) parseAddExpression(compoundExpression Operation) Operation {
	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	if slices.Contains(addTokens, peekedToken.Kind) {
		op := peekedToken.Value.(string)
		p.match(peekedToken.Kind)
		rhs := p.parseCompoundExpression(false, false, true, false)

		addExpr := &BinaryExpr{Op: op, Lhs: compoundExpression, Rhs: rhs}
		return p.parseAddExpression(addExpr)
	}

	return compoundExpression
}

func (p *Parser) parseCompareExpression(compoundExpression Operation) Operation {
	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	op := peekedToken.Value.(string)
	p.match(peekedToken.Kind)
	rhs := p.parseCompoundExpression(false, false, false, true)

	// Comparison binary operations can not be nested
	peekedTokens = p.lexer.Peek(1)
	peekedToken = &peekedTokens[0]
	if slices.Contains(compareTokens, peekedToken.Kind) {
		p.syntaxError(ComparisonNotAssociative)
	}

	return &BinaryExpr{Op: op, Lhs: compoundExpression, Rhs: rhs}
}
