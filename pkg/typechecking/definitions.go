package typechecking

import (
	"chogopy/pkg/ast"
	"maps"
)

func (st *StaticTyping) VisitFuncDef(funcDef *ast.FuncDef) {
	funcInfo := st.localEnv.check(funcDef.FuncName, false)

	paramNames := funcInfo.(FunctionInfo).paramNames
	paramTypes := funcInfo.(FunctionInfo).funcType.paramTypes
	returnType := funcInfo.(FunctionInfo).funcType.returnType
	nestedDefs := funcInfo.(FunctionInfo).nestedDefs

	extendedEnv := maps.Clone(st.localEnv)
	for i := range len(paramNames) {
		paramName := paramNames[i]
		paramType := paramTypes[i]
		extendedEnv[paramName] = paramType
	}
	for i := range len(nestedDefs) {
		nestedDefName := nestedDefs[i].defName
		nestedDefType := nestedDefs[i].defType
		extendedEnv[nestedDefName] = nestedDefType
	}

	funcBodyVisitor := &StaticTyping{
		localEnv:   extendedEnv,
		returnType: returnType,
	}
	for _, funcBodyNode := range funcDef.FuncBody {
		funcBodyNode.Visit(funcBodyVisitor)
	}
}

func (st *StaticTyping) VisitGlobalDecl(globalDecl *ast.GlobalDecl) {}

func (st *StaticTyping) VisitNonLocalDecl(nonLocalDecl *ast.NonLocalDecl) {}

func (st *StaticTyping) VisitVarDef(varDef *ast.VarDef) {
	varName := varDef.TypedVar.(*ast.TypedVar).VarName
	varType := st.localEnv.check(varName, true)

	varDef.Literal.Visit(st)
	literalType := st.visitedType

	checkAssignmentCompatible(literalType, varType)
}
