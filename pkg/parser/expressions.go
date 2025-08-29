package parser

import (
	"chogopy/pkg/ast"
	"chogopy/pkg/lexer"
	"slices"
)

func (p *Parser) parseExpression(insideAnd bool, insideOr bool) ast.Node {
	var expression ast.Node

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
		elseNode := p.parseExpression(false, false)
		return &ast.IfExpr{Condition: condition, IfNode: expression, ElseNode: elseNode}
	}

	if expression == nil {
		p.syntaxError(TokenNotFound)
	}
	return expression
}

func (p *Parser) parseNotExpression() ast.Node {
	p.match(lexer.NOT)

	if p.check(lexer.NOT) {
		return &ast.UnaryExpr{Op: "not", Value: p.parseNotExpression()}
	}

	return &ast.UnaryExpr{Op: "not", Value: p.parseCompoundExpression(false, false, false, false)}
}

func (p *Parser) parseAndExpression(expression ast.Node) ast.Node {
	if p.check(lexer.AND) {
		p.match(lexer.AND)
		rhs := p.parseExpression(true, false)
		andExpression := &ast.BinaryExpr{Op: "and", Lhs: expression, Rhs: rhs}
		return p.parseAndExpression(andExpression)
	}

	return expression
}

func (p *Parser) parseOrExpression(expression ast.Node) ast.Node {
	if p.check(lexer.OR) {
		p.match(lexer.OR)
		rhs := p.parseExpression(false, true)
		orExpression := &ast.BinaryExpr{Op: "or", Lhs: expression, Rhs: rhs}
		return p.parseOrExpression(orExpression)
	}

	return expression
}

func (p *Parser) parseCompoundExpression(insideNegation bool, insideMult bool, insideAdd bool, insideCompare bool) ast.Node {
	var compoundExpression ast.Node

	if p.check(lexer.MINUS) {
		compoundExpression = p.parseUnaryNegation()
	}

	if p.nextTokenIn(simpleCompoundExpressionTokens) {
		compoundExpression = p.parseSimpleCompoundExpression()
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

func (p *Parser) parseSimpleCompoundExpression() ast.Node {
	if p.nextTokenIn(literalTokens) {
		return p.parseLiteral()
	}

	if p.check(lexer.IDENTIFIER, lexer.LROUNDBRACKET) {
		funcNameToken := p.match(lexer.IDENTIFIER)
		funcName := funcNameToken.Value.(string)
		p.match(lexer.LROUNDBRACKET)
		arguments := p.parseExpressionList()
		p.match(lexer.RROUNDBRACKET)
		return &ast.CallExpr{FuncName: funcName, Arguments: arguments}
	}

	if p.check(lexer.IDENTIFIER) {
		identifierToken := p.match(lexer.IDENTIFIER)
		identifier := identifierToken.Value.(string)
		return &ast.IdentExpr{Identifier: identifier}
	}

	if p.check(lexer.LSQUAREBRACKET) {
		p.match(lexer.LSQUAREBRACKET)
		elements := p.parseExpressionList()
		p.match(lexer.RSQUAREBRACKET)
		return &ast.ListExpr{Elements: elements}
	}

	if p.check(lexer.LROUNDBRACKET) {
		p.match(lexer.LROUNDBRACKET)
		expression := p.parseExpression(false, false)
		p.match(lexer.RROUNDBRACKET)
		return expression
	}

	p.syntaxError(TokenNotFound)
	return nil
}

func (p *Parser) parseExpressionList() []ast.Node {
	expressionList := []ast.Node{}

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

func (p *Parser) parseUnaryNegation() ast.Node {
	p.match(lexer.MINUS)

	if p.check(lexer.MINUS) {
		return &ast.UnaryExpr{Op: "-", Value: p.parseUnaryNegation()}
	}

	return &ast.UnaryExpr{Op: "-", Value: p.parseCompoundExpression(true, false, false, false)}
}

func (p *Parser) parseIndexExpression(compoundExpression ast.Node) ast.Node {
	p.match(lexer.LSQUAREBRACKET)
	index := p.parseExpression(false, false)
	p.match(lexer.RSQUAREBRACKET)

	indexExpression := &ast.IndexExpr{Value: compoundExpression, Index: index}
	for p.check(lexer.LSQUAREBRACKET) {
		p.match(lexer.LSQUAREBRACKET)
		index = p.parseExpression(false, false)
		p.match(lexer.RSQUAREBRACKET)
		indexExpression = &ast.IndexExpr{Value: indexExpression, Index: index}
	}

	return indexExpression
}

func (p *Parser) parseMultExpression(compoundExpression ast.Node) ast.Node {
	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	if slices.Contains(multTokens, peekedToken.Kind) {
		op := peekedToken.Value.(string)
		p.match(peekedToken.Kind)
		rhs := p.parseCompoundExpression(false, true, false, false)

		multExpr := &ast.BinaryExpr{Op: op, Lhs: compoundExpression, Rhs: rhs}
		return p.parseMultExpression(multExpr)
	}

	return compoundExpression
}

func (p *Parser) parseAddExpression(compoundExpression ast.Node) ast.Node {
	peekedTokens := p.lexer.Peek(1)
	peekedToken := &peekedTokens[0]

	if slices.Contains(addTokens, peekedToken.Kind) {
		op := peekedToken.Value.(string)
		p.match(peekedToken.Kind)
		rhs := p.parseCompoundExpression(false, false, true, false)

		addExpr := &ast.BinaryExpr{Op: op, Lhs: compoundExpression, Rhs: rhs}
		return p.parseAddExpression(addExpr)
	}

	return compoundExpression
}

func (p *Parser) parseCompareExpression(compoundExpression ast.Node) ast.Node {
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

	return &ast.BinaryExpr{Op: op, Lhs: compoundExpression, Rhs: rhs}
}
