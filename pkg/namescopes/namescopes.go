// Package namescopes defines methods for analyzing the correct usage
// of scoped variables
package namescopes

import (
	"chogopy/pkg/ast"
)

// TODO: This analysis pass is not yet complete since we have set its traverse property to false
// and have not yet explicitly defined a visitor method for each possible AST node that
// would then correctly visit the nodes' child nodes and perform the required name scope checks.

type NameScopes struct {
	NameContext *NameContext
	ast.BaseVisitor
}

func (ns *NameScopes) Analyze(program *ast.Program) {
	NameContextBuilder := NameContextBuilder{}
	NameContextBuilder.Analyze(program)

	ns.NameContext = &NameContextBuilder.NameContext
	ns.NameContext.addFuncName("print", NewNameContext())
	ns.NameContext.addFuncName("len", NewNameContext())
	ns.NameContext.addFuncName("input", NewNameContext())

	for _, definition := range program.Definitions {
		definition.Visit(ns)
	}
	for _, statement := range program.Statements {
		statement.Visit(ns)
	}
}

func (ns *NameScopes) Traverse() bool {
	return false
}

func (ns *NameScopes) VisitIdentExpr(identExpr *ast.IdentExpr) {
	identName := identExpr.Identifier

	if !ns.NameContext.contains(identName) &&
		!ns.NameContext.parentScopeContains(identName) {
		nameSemanticError(IdentifierUndefined, identName)
	}
}

func (ns *NameScopes) VisitCallExpr(callExpr *ast.CallExpr) {
	funcName := callExpr.FuncName

	if !ns.NameContext.contains(funcName) &&
		!ns.NameContext.parentScopeContains(funcName) {
		nameSemanticError(IdentifierUndefined, funcName)
	}
}

func (ns *NameScopes) VisitFuncDef(funcDef *ast.FuncDef) {
	funcName := funcDef.FuncName
	funcContext := ns.NameContext.getContext(funcName)

	funcNameScopes := &NameScopes{NameContext: funcContext}

	for _, bodyNode := range funcDef.FuncBody {
		bodyNode.Visit(funcNameScopes)
	}
}

func (ns *NameScopes) VisitForStmt(forStmt *ast.ForStmt) {
	iterName := forStmt.IterName

	if !ns.NameContext.contains(iterName) {
		nameSemanticError(IdentifierUndefined, iterName)
	}

	forStmt.Iter.Visit(ns)

	for _, bodyOp := range forStmt.Body {
		bodyOp.Visit(ns)
	}
}

func (ns *NameScopes) VisitAssignStmt(assignStmt *ast.AssignStmt) {
	_, targetIsIdentifier := assignStmt.Target.(*ast.IdentExpr)

	if targetIsIdentifier {
		identName := assignStmt.Target.(*ast.IdentExpr).Identifier

		if !ns.NameContext.contains(identName) {
			nameSemanticError(AssignTargetOutOfScope, identName)
		}
	}
}

func (ns *NameScopes) VisitNonLocalDecl(nonLocalDecl *ast.NonLocalDecl) {
	declName := nonLocalDecl.DeclName

	if !ns.NameContext.parentScopeContains(declName) {
		nameSemanticError(IdentifierNotInParentScope, declName)
	}
}

func (ns *NameScopes) VisitGlobalDecl(globalDecl *ast.GlobalDecl) {
	declName := globalDecl.DeclName

	if !ns.NameContext.globalScopeContains(declName) {
		nameSemanticError(IdentifierNotInGlobalScope, declName)
	}
}
