package astanalysis

import (
	"chogopy/pkg/parser"
	"log"
)

type Type interface{}

type BasicType struct {
	typeName string
	Type
}

type ListType struct {
	elemType Type
	Type
}

type BottomType struct {
	Type
}

type ObjectType struct {
	Type
}

type FunctionType struct {
	paramTypes []Type
	returnType Type
}

var (
	intType    = BasicType{typeName: "int"}
	boolType   = BasicType{typeName: "bool"}
	strType    = BasicType{typeName: "str"}
	noneType   = BasicType{typeName: "<None>"}
	emptyType  = BasicType{typeName: "<Empty>"}
	bottomType = BottomType{}
	objectType = ObjectType{}
)

// TODO: type from op
func typeFromOp(op parser.Operation) Type {
	switch op := op.(type) {
	case *parser.NamedType:
		switch op.TypeName {
		case "int":
			return intType
		case "bool":
			return boolType
		case "str":
			return strType
		case "<None>":
			return noneType
		case "<Empty>":
			return emptyType
		case "object":
			return objectType
		}
	case *parser.ListType:
		elemType := typeFromOp(op.ElemType)
		return ListType{elemType: elemType}
	}
	return nil
}

func join(t1 Type, t2 Type) Type {
	if isAssignmentCompatible(t1, t2) {
		return t2
	}
	if isAssignmentCompatible(t2, t1) {
		return t1
	}
	return objectType
}

func isSubType(t1 Type, t2 Type) bool {
	_, t1IsList := t1.(ListType)

	switch {
	case t1 == t2:
		return true
	case (t1 == intType || t1 == boolType || t1 == strType || t1IsList) && t2 == objectType:
		return true
	case t1 == noneType && t2 == objectType:
		return true
	case t1 == emptyType && t2 == objectType:
		return true
	case t1 == bottomType:
		return true
	}
	return false
}

func isAssignmentCompatible(t1 Type, t2 Type) bool {
	_, t1IsList := t1.(ListType)
	_, t2IsList := t2.(ListType)

	switch {
	case isSubType(t1, t2) && t1 != bottomType:
		return true
	case t1 == noneType && t2 != intType && t2 != boolType && t2 != strType && t2 != bottomType:
		return true
	case t1 == emptyType && t2IsList:
		return true
	case t1IsList && t2IsList && t1.(ListType).elemType == noneType && isAssignmentCompatible(noneType, t2.(ListType).elemType):
		return true
	}
	return false
}

func checkAssignmentCompatible(t1 Type, t2 Type) {
	if !isAssignmentCompatible(t1, t2) {
		log.Fatalf("Semantic Error: Expected '%v' and '%v' to be assignment compatible", t1, t2)
	}
}

func checkType(found Type, expected Type) {
	if found != expected {
		log.Fatalf("Semantic Error: Expected '%v' but found '%v'", expected, found)
	}
}

func checkListType(found Type) Type {
	_, foundIsList := found.(ListType)
	if !foundIsList {
		log.Fatalf("Semantic Error: Expected list type but found '%v'", found)
	}

	return found.(ListType).elemType
}

// DefType  TODO: just replace EnvInfo with any if it causes problems
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

type EnvironmentBuilder struct {
	LocalEnvironment LocalEnvironment
	parser.BaseVisitor
}

func (eb *EnvironmentBuilder) Analyze(program *parser.Program) {
	eb.LocalEnvironment = LocalEnvironment{
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

func (eb *EnvironmentBuilder) VisitTypedVar(typedVar *parser.TypedVar) {
	varName := typedVar.VarName
	varType := typeFromOp(typedVar.VarType)
	eb.LocalEnvironment[varName] = varType
}

func (eb *EnvironmentBuilder) VisitFuncDef(funcDef *parser.FuncDef) {
	funcName := funcDef.FuncName

	paramNames := []string{}
	paramTypes := []Type{}
	for _, param := range funcDef.Parameters {
		paramName := param.(*parser.TypedVar).VarName
		paramType := typeFromOp(param.(*parser.TypedVar).VarType)
		paramNames = append(paramNames, paramName)
		paramTypes = append(paramTypes, paramType)
	}

	returnType := typeFromOp(funcDef.ReturnType)

	nestedDefsBuilder := &EnvironmentBuilder{LocalEnvironment: map[string]DefType{}}
	for _, bodyOp := range funcDef.FuncBody {
		bodyOp.Visit(nestedDefsBuilder)
	}

	nestedDefs := []Definition{}
	for nestedDefName, nestedDefType := range nestedDefsBuilder.LocalEnvironment {
		nestedDef := Definition{defName: nestedDefName, defType: nestedDefType}
		nestedDefs = append(nestedDefs, nestedDef)
	}

	eb.LocalEnvironment[funcName] = FunctionInfo{
		funcType:   FunctionType{paramTypes: paramTypes, returnType: returnType},
		paramNames: paramNames,
		nestedDefs: nestedDefs,
	}
}

type StaticTyping struct {
	parser.BaseVisitor
}

func (st *StaticTyping) Analyze(program *parser.Program) {
	program.Visit(st)
}
