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

func (nc NameContext) addVarName(name string) {
	if nc.contains(name) {
		nameSemanticError(IdentifierAlreadyDefined, name)
	}
	nc.names[name] = &NameContext{}
}

func (nc NameContext) addFuncName(name string, nestedContext NameContext) {
	if nc.contains(name) {
		nameSemanticError(IdentifierAlreadyDefined, name)
	}
	nc.names[name] = &nestedContext
}

func (nc NameContext) getContext(name string) *NameContext {
	if !nc.contains(name) {
		nameSemanticError(IdentifierUndefined, name)
	}
	return nc.names[name]
}

func (nc NameContext) getGlobalContext() NameContext {
	if nc.parentScope == nil {
		return nc
	}
	return nc.parentScope.getGlobalContext()
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
	nestedContext := NameContext{parentScope: &nb.nameContext}

	for _, param := range funcDef.Parameters {
		paramName := param.(*ast.TypedVar).VarName
		nestedContext.addVarName(paramName)
	}

	for _, funcBodyNode := range funcDef.FuncBody {
		switch funcBodyNode := funcBodyNode.(type) {
		case *ast.NonLocalDecl:
			nestedContext.addVarName(funcBodyNode.DeclName)
		case *ast.GlobalDecl:
			nestedContext.addVarName(funcBodyNode.DeclName)
		}
	}

	nestedContextBuilder := &NameContextBuilder{nameContext: nestedContext}
	for _, funcBodyNode := range funcDef.FuncBody {
		funcBodyNode.Visit(nestedContextBuilder)
	}

	nb.nameContext.addFuncName(funcDef.FuncName, nestedContextBuilder.nameContext)
}

type NameScopes struct {
	nameContext NameContext
	ast.BaseVisitor
}

func (ns *NameScopes) Analyze(program *ast.Program) {
	nameContextBuilder := NameContextBuilder{}
	nameContextBuilder.Analyze(program)

	ns.nameContext = nameContextBuilder.nameContext
	ns.nameContext.addFuncName("print", NewNameContext())
	ns.nameContext.addFuncName("len", NewNameContext())
	ns.nameContext.addFuncName("input", NewNameContext())

	program.Visit(ns)
}

func (ns *NameScopes) VisitIdentExpr(identExpr *ast.IdentExpr)          {}
func (ns *NameScopes) VisitCallExpr(callExpr *ast.CallExpr)             {}
func (ns *NameScopes) VisitFuncDef(funcDef *ast.FuncDef)                {}
func (ns *NameScopes) VisitForStmt(forStmt *ast.ForStmt)                {}
func (ns *NameScopes) VisitAssignStmt(assignStmt *ast.AssignStmt)       {}
func (ns *NameScopes) VisitNonLocalDecl(nonLocalDecl *ast.NonLocalDecl) {}
func (ns *NameScopes) VisitGlobalDecl(globalDecl *ast.GlobalDecl)       {}
