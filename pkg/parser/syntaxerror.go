package parser

import (
	"chogopy/pkg/lexer"
	"fmt"
	"os"
	"strings"
)

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
