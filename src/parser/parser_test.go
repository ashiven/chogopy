package parser

import (
	"chogopy/src/ast"
	"chogopy/src/lexer"
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

func matchParsed(stream string, expectedAst ast.Program) bool {
	lexer := lexer.NewLexer(stream)
	parser := NewParser(&lexer)
	parsedAst := parser.ParseProgram()

	if !reflect.DeepEqual(expectedAst, parsedAst) {
		diffs := pretty.Diff(expectedAst, parsedAst)
		for _, diff := range diffs {
			pretty.Println(diff)
		}
		return false
	}
	return true
}

func TestArithmetic(t *testing.T) {
	stream := `def foo():
	1 + 2`

	expectedAst := ast.Program{
		Definitions: []ast.Node{
			&ast.FuncDef{
				FuncName:   "foo",
				Parameters: []ast.Node{},
				FuncBody: []ast.Node{
					&ast.BinaryExpr{
						Op: "+",
						Lhs: &ast.LiteralExpr{
							Value: 1,
						},
						Rhs: &ast.LiteralExpr{
							Value: 2,
						},
					},
				},
				ReturnType: &ast.NamedType{
					TypeName: "<None>",
				},
			},
		},
		Statements: []ast.Node{},
	}

	if !matchParsed(stream, expectedAst) {
		t.Fatalf("Expected AST did not match parsed AST.")
	}
}

func TestBrackets(t *testing.T) {
	stream := `
def foo():


	[0]`

	expectedAst := ast.Program{
		Definitions: []ast.Node{
			&ast.FuncDef{
				FuncName:   "foo",
				Parameters: []ast.Node{},
				FuncBody: []ast.Node{
					&ast.ListExpr{
						Elements: []ast.Node{
							&ast.LiteralExpr{
								Value: 0,
							},
						},
					},
				},
				ReturnType: &ast.NamedType{
					TypeName: "<None>",
				},
			},
		},
		Statements: []ast.Node{},
	}

	if !matchParsed(stream, expectedAst) {
		t.Fatalf("Expected AST did not match parsed AST.")
	}
}

func TestDivision(t *testing.T) {
	stream := "0 // 1"

	expectedAst := ast.Program{
		Definitions: []ast.Node{},
		Statements: []ast.Node{
			&ast.BinaryExpr{
				Op: "//",
				Lhs: &ast.LiteralExpr{
					Value: 0,
				},
				Rhs: &ast.LiteralExpr{
					Value: 1,
				},
			},
		},
	}

	if !matchParsed(stream, expectedAst) {
		t.Fatalf("Expected AST did not match parsed AST.")
	}
}

func TestEndWithComment(t *testing.T) {
	stream := `
def foo():
	0 # Comment with newline
`
	expectedAst := ast.Program{
		Definitions: []ast.Node{
			&ast.FuncDef{
				FuncName:   "foo",
				Parameters: []ast.Node{},
				FuncBody: []ast.Node{
					&ast.LiteralExpr{
						Value: 0,
					},
				},
				ReturnType: &ast.NamedType{
					TypeName: "<None>",
				},
			},
		},
		Statements: []ast.Node{},
	}

	if !matchParsed(stream, expectedAst) {
		t.Fatalf("Expected AST did not match parsed AST.")
	}
}

func TestFunctionDefinitions(t *testing.T) {
	stream := `
def foo():
	if True:
		return

def bar():
	return

pass
`

	expectedAst := ast.Program{
		Definitions: []ast.Node{
			&ast.FuncDef{
				FuncName:   "foo",
				Parameters: []ast.Node{},
				FuncBody: []ast.Node{
					&ast.IfStmt{
						Condition: &ast.LiteralExpr{
							Value: true,
						},
						IfBody: []ast.Node{
							&ast.ReturnStmt{},
						},
						ElseBody: []ast.Node{},
					},
				},
				ReturnType: &ast.NamedType{
					TypeName: "<None>",
				},
			},
			&ast.FuncDef{
				FuncName:   "bar",
				Parameters: []ast.Node{},
				FuncBody: []ast.Node{
					&ast.ReturnStmt{},
				},
				ReturnType: &ast.NamedType{
					TypeName: "<None>",
				},
			},
		},
		Statements: []ast.Node{
			&ast.PassStmt{},
		},
	}

	if !matchParsed(stream, expectedAst) {
		t.Fatalf("Expected AST did not match parsed AST.")
	}
}
