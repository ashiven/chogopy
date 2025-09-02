package scopes

import "chogopy/pkg/ast"

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
		semanticError(IdentifierAlreadyDefined, name)
	}
	nc.names[name] = &NameContext{}
}

func (nc NameContext) addFuncName(name string, funcContext NameContext) {
	if nc.contains(name) {
		semanticError(IdentifierAlreadyDefined, name)
	}
	nc.names[name] = &funcContext
}

func (nc NameContext) getContext(name string) *NameContext {
	if !nc.contains(name) {
		semanticError(IdentifierUndefined, name)
	}
	return nc.names[name]
}

type NameContextBuilder struct {
	NameContext NameContext
	ast.BaseVisitor
}

func (nb *NameContextBuilder) Analyze(program *ast.Program) {
	nb.NameContext = NewNameContext()

	for _, definition := range program.Definitions {
		definition.Visit(nb)
	}
}

func (nb *NameContextBuilder) Traverse() bool {
	return false
}

func (nb *NameContextBuilder) VisitVarDef(varDef *ast.VarDef) {
	varName := varDef.TypedVar.(*ast.TypedVar).VarName
	nb.NameContext.addVarName(varName)
}

func (nb *NameContextBuilder) VisitFuncDef(funcDef *ast.FuncDef) {
	funcContext := NameContext{
		names:       map[string]*NameContext{},
		parentScope: &nb.NameContext,
	}

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

	funcContextBuilder := &NameContextBuilder{NameContext: funcContext}
	for _, funcBodyNode := range funcDef.FuncBody {
		funcBodyNode.Visit(funcContextBuilder)
	}

	nb.NameContext.addFuncName(funcDef.FuncName, funcContextBuilder.NameContext)
}
