package astanalysis

import (
	"chogopy/pkg/ast"
)

type NameContext struct {
	names       map[string]*NameContext
	parentScope *NameContext
}

func NewNameContext() NameContext {
	return NameContext{
		names:       map[string]*NameContext{},
		parentScope: nil,
	}
}

func (nc NameContext) contains(name string) bool {
	_, nameInContext := nc.names[name]
	return nameInContext
}

func (nc NameContext) parentScopeContains(name string) bool {
	if nc.parentScope != nil {
		return nc.parentScope.contains(name) || nc.parentScope.parentScopeContains(name)
	}
	return false
}

func (nc NameContext) globalScopeContains(name string) bool {
	if nc.parentScope == nil {
		return nc.contains(name)
	}
	return nc.parentScope.globalScopeContains(name)
}

func (nc NameContext) addVarName(name string) {
	if nc.contains(name) {
		nameSemanticError(IdentifierAlreadyDefined, name)
	}
	nc.names[name] = &NameContext{}
}

func (nc NameContext) addFuncName(name string, funcContext NameContext) {
	if nc.contains(name) {
		nameSemanticError(IdentifierAlreadyDefined, name)
	}
	nc.names[name] = &funcContext
}

func (nc NameContext) getContext(name string) *NameContext {
	if !nc.contains(name) {
		nameSemanticError(IdentifierUndefined, name)
	}
	return nc.names[name]
}

type NameContextBuilder struct {
	nameContext NameContext
	ast.BaseVisitor
}

func (nb *NameContextBuilder) Analyze(program *ast.Program) {
	nb.nameContext = NewNameContext()

	program.Visit(nb)
}

func (nb *NameContextBuilder) VisitVarDef(varDef *ast.VarDef) {
	varName := varDef.TypedVar.(*ast.TypedVar).VarName
	nb.nameContext.addVarName(varName)
}

func (nb *NameContextBuilder) VisitFuncDef(funcDef *ast.FuncDef) {
	funcContext := NameContext{parentScope: &nb.nameContext}

	for _, param := range funcDef.Parameters {
		paramName := param.(*ast.TypedVar).VarName
		funcContext.addVarName(paramName)
	}

	for _, funcBodyNode := range funcDef.FuncBody {
		switch funcBodyNode := funcBodyNode.(type) {
		case *ast.NonLocalDecl:
			funcContext.addVarName(funcBodyNode.DeclName)
		case *ast.GlobalDecl:
			funcContext.addVarName(funcBodyNode.DeclName)
		}
	}

	funcContextBuilder := &NameContextBuilder{nameContext: funcContext}
	for _, funcBodyNode := range funcDef.FuncBody {
		funcBodyNode.Visit(funcContextBuilder)
	}

	nb.nameContext.addFuncName(funcDef.FuncName, funcContextBuilder.nameContext)
}

type NameScopes struct {
	nameContext *NameContext
	ast.BaseVisitor
}

func (ns *NameScopes) Analyze(program *ast.Program) {
	nameContextBuilder := NameContextBuilder{}
	nameContextBuilder.Analyze(program)

	ns.nameContext = &nameContextBuilder.nameContext
	ns.nameContext.addFuncName("print", NewNameContext())
	ns.nameContext.addFuncName("len", NewNameContext())
	ns.nameContext.addFuncName("input", NewNameContext())

	program.Visit(ns)
}

func (ns *NameScopes) VisitIdentExpr(identExpr *ast.IdentExpr) {
	identName := identExpr.Identifier

	if !ns.nameContext.contains(identName) &&
		!ns.nameContext.parentScopeContains(identName) {
		nameSemanticError(IdentifierUndefined, identName)
	}
}

func (ns *NameScopes) VisitCallExpr(callExpr *ast.CallExpr) {
	funcName := callExpr.FuncName

	if !ns.nameContext.contains(funcName) &&
		!ns.nameContext.parentScopeContains(funcName) {
		nameSemanticError(IdentifierUndefined, funcName)
	}
}

func (ns *NameScopes) VisitFuncDef(funcDef *ast.FuncDef) {
	funcName := funcDef.FuncName
	funcContext := ns.nameContext.getContext(funcName)

	funcNameScopes := &NameScopes{nameContext: funcContext}

	for _, bodyNode := range funcDef.FuncBody {
		bodyNode.Visit(funcNameScopes)
	}
}

func (ns *NameScopes) VisitForStmt(forStmt *ast.ForStmt) {
	iterName := forStmt.IterName

	if !ns.nameContext.contains(iterName) {
		nameSemanticError(IdentifierUndefined, iterName)
	}
}

func (ns *NameScopes) VisitAssignStmt(assignStmt *ast.AssignStmt) {
	_, targetIsIdentifier := assignStmt.Target.(*ast.IdentExpr)

	if targetIsIdentifier {
		identName := assignStmt.Target.(*ast.IdentExpr).Identifier

		if !ns.nameContext.contains(identName) {
			nameSemanticError(AssignTargetOutOfScope, identName)
		}
	}
}

func (ns *NameScopes) VisitNonLocalDecl(nonLocalDecl *ast.NonLocalDecl) {
	declName := nonLocalDecl.DeclName

	if !ns.nameContext.parentScopeContains(declName) {
		nameSemanticError(IdentifierNotInParentScope, declName)
	}
}

func (ns *NameScopes) VisitGlobalDecl(globalDecl *ast.GlobalDecl) {
	declName := globalDecl.DeclName

	if !ns.nameContext.globalScopeContains(declName) {
		nameSemanticError(IdentifierNotInGlobalScope, declName)
	}
}
