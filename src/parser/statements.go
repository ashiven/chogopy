package parser

import (
	"chogopy/src/ast"
	"chogopy/src/lexer"
)

func (p *Parser) parseStatements() []ast.Node {
	statements := []ast.Node{}

	for p.nextTokenIn(expressionTokens) || p.nextTokenIn(statementTokens) {
		statement := p.parseStatement()
		statements = append(statements, statement)
	}

	return statements
}

func (p *Parser) parseStatement() ast.Node {
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
		p.match(lexer.DEDENT)
		elseBody := p.parseElseBody()
		if len(elseBody) > 0 {
			p.match(lexer.DEDENT)
		}
		return &ast.IfStmt{Condition: condition, IfBody: ifBody, ElseBody: elseBody}
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
		return &ast.WhileStmt{Condition: condition, Body: body}
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
		return &ast.ForStmt{IterName: iterName, Iter: iter, Body: body}
	}

	p.syntaxError(TokenNotFound)
	return nil
}

func (p *Parser) parseElseBody() []ast.Node {
	elseBody := []ast.Node{}

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
		p.match(lexer.DEDENT)
		elifElseBody := p.parseElseBody()

		elif := &ast.IfStmt{Condition: condition, IfBody: elifIfBody, ElseBody: elifElseBody}

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

func (p *Parser) parseSimpleStatement() ast.Node {
	if p.check(lexer.IDENTIFIER, lexer.COLON) {
		p.syntaxError(VariableDefinedLater)
	}

	if p.check(lexer.PASS) {
		p.match(lexer.PASS)
		return &ast.PassStmt{}
	}

	if p.check(lexer.RETURN) {
		p.match(lexer.RETURN)
		var returnVal ast.Node
		if p.nextTokenIn(expressionTokens) {
			returnVal = p.parseExpression(false, false)
		}
		return &ast.ReturnStmt{ReturnVal: returnVal}
	}

	if p.nextTokenIn(expressionTokens) {
		return p.parseExpressionAssignList()
	}

	p.syntaxError(TokenNotFound)
	return nil
}

func (p *Parser) parseExpressionAssignList() ast.Node {
	expression := p.parseExpression(false, false)

	if p.check(lexer.ASSIGN) {
		p.match(lexer.ASSIGN)
		return &ast.AssignStmt{Target: expression, Value: p.parseExpressionAssignList()}
	}

	return expression
}
