package typechecks

import (
	"chogopy/src/ast"
)

type DefType interface {
	Type | FunctionInfo
}

type Definition struct {
	defName string
	defType DefType
}

type FunctionInfo struct {
	funcType   FunctionType
	paramNames []string
	nestedDefs []Definition
}

// LocalEnvironment associates every declared variable and function with their type.
// It maps the names of the variables/functions to their type.
type LocalEnvironment map[string]DefType

func (le LocalEnvironment) check(defName string, expectVarDef bool) DefType {
	defType, defExists := le[defName]
	if !defExists {
		semanticError(UnknownIdentifierUsed, nil, nil, defName, 0, 0)
	}

	_, isFuncInfo := defType.(FunctionInfo)
	if isFuncInfo && expectVarDef {
		semanticError(ExpectedVariableIdentifier, nil, nil, defName, 0, 0)
	}

	if !isFuncInfo && !expectVarDef {
		semanticError(ExpectedFunctionIdentifier, nil, nil, defName, 0, 0)
	}

	return defType
}

// EnvironmentBuilder is responsible for traversing the AST and constructing the above-defined
// LocalEnvironment by checking every VarDef and FuncDef AST node
type EnvironmentBuilder struct {
	LocalEnv LocalEnvironment
	ast.BaseVisitor
}

func (eb *EnvironmentBuilder) Build(program *ast.Program) {
	eb.LocalEnv = LocalEnvironment{}

	eb.LocalEnv = LocalEnvironment{
		"len": FunctionInfo{
			funcType:   FunctionType{paramTypes: []Type{objectType}, returnType: intType},
			paramNames: []string{"arg"},
			nestedDefs: []Definition{},
		},
		"print": FunctionInfo{
			funcType:   FunctionType{paramTypes: []Type{objectType}, returnType: noneType},
			paramNames: []string{"arg"},
			nestedDefs: []Definition{},
		},
		"input": FunctionInfo{
			funcType:   FunctionType{paramTypes: []Type{}, returnType: strType},
			paramNames: []string{},
			nestedDefs: []Definition{},
		},
	}

	program.Visit(eb)
}

func (eb *EnvironmentBuilder) VisitTypedVar(typedVar *ast.TypedVar) {
	varName := typedVar.VarName
	varType := typeFromNode(typedVar.VarType)
	eb.LocalEnv[varName] = varType
}

func (eb *EnvironmentBuilder) VisitFuncDef(funcDef *ast.FuncDef) {
	funcName := funcDef.FuncName

	paramNames := []string{}
	paramTypes := []Type{}
	for _, param := range funcDef.Parameters {
		paramName := param.(*ast.TypedVar).VarName
		paramType := typeFromNode(param.(*ast.TypedVar).VarType)
		paramNames = append(paramNames, paramName)
		paramTypes = append(paramTypes, paramType)
	}

	returnType := typeFromNode(funcDef.ReturnType)

	nestedDefsBuilder := &EnvironmentBuilder{LocalEnv: LocalEnvironment{}}
	for _, bodyNode := range funcDef.FuncBody {
		bodyNode.Visit(nestedDefsBuilder)
	}

	nestedDefs := []Definition{}
	for nestedDefName, nestedDefType := range nestedDefsBuilder.LocalEnv {
		nestedDef := Definition{defName: nestedDefName, defType: nestedDefType}
		nestedDefs = append(nestedDefs, nestedDef)
	}

	eb.LocalEnv[funcName] = FunctionInfo{
		funcType:   FunctionType{paramTypes: paramTypes, returnType: returnType},
		paramNames: paramNames,
		nestedDefs: nestedDefs,
	}
}
