package astanalysis

import (
	"chogopy/pkg/parser"
	"log"
	"slices"
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

func checkListType(found Type) {
	_, foundIsList := found.(ListType)
	if !foundIsList {
		log.Fatalf("Semantic Error: Expected list type but found '%v'", found)
	}
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
	localEnv   LocalEnvironment
	returnType Type
	parser.BaseVisitor
}

// Analyze performs static type checking according to the rules defined in
// chapter 5 of the chocopy language reference:
//
// https://chocopy.org/chocopy_language_reference.pdf
func (st *StaticTyping) Analyze(program *parser.Program) {
	envBuilder := &EnvironmentBuilder{LocalEnvironment: map[string]DefType{}}
	envBuilder.Analyze(program)

	st.localEnv = envBuilder.LocalEnvironment
	st.returnType = bottomType

	program.Visit(st)
}

func (st *StaticTyping) VisitVarDef(varDef *parser.VarDef) {
	varName := varDef.TypedVar.(*parser.TypedVar).VarName
	varType, varDefined := st.localEnv[varName]
	if !varDefined {
		log.Fatalf("Semantic Error: Unknown identifier used: %s", varName)
	}

	_, isFuncType := varType.(FunctionInfo)
	if isFuncType {
		log.Fatalf("Semantic Error: Found function identifier: %s but expected variable identifier", varName)
	}

	varDef.Literal.Visit(st)
	literalType := st.returnType

	checkAssignmentCompatible(literalType, varType)
}

func (st *StaticTyping) VisitLiteralExpr(literalExpr *parser.LiteralExpr) {
	switch literalExpr.Value.(type) {
	case int:
		st.returnType = intType
	case bool:
		st.returnType = boolType
	case string:
		st.returnType = strType
	default:
		st.returnType = noneType
	}
}

func (st *StaticTyping) VisitIdentExpr(identExpr *parser.IdentExpr) {
	varType, varDefined := st.localEnv[identExpr.Identifier]
	if !varDefined {
		log.Fatalf("Semantic Error: Unknown identifier used: %s", identExpr.Identifier)
	}

	_, isFuncType := varType.(FunctionInfo)
	if isFuncType {
		log.Fatalf("Semantic Error: Found function identifier: %s but expected variable identifier", identExpr.Identifier)
	}
}

func (st *StaticTyping) VisitUnaryExpr(unaryExpr *parser.UnaryExpr) {
	unaryExpr.Value.Visit(st)

	switch unaryExpr.Op {
	case "-":
		checkType(st.returnType, intType)
		st.returnType = intType
	case "not":
		checkType(st.returnType, boolType)
		st.returnType = boolType
	}
}

func (st *StaticTyping) VisitBinaryExpr(binaryExpr *parser.BinaryExpr) {
	binaryExpr.Lhs.Visit(st)
	lhsType := st.returnType

	binaryExpr.Rhs.Visit(st)
	rhsType := st.returnType

	_, lhsIsList := lhsType.(ListType)
	_, rhsIsList := rhsType.(ListType)

	lhsIsString := lhsType == strType
	rhsIsString := rhsType == strType

	switch binaryExpr.Op {
	case "and":
		checkType(lhsType, boolType)
		checkType(rhsType, boolType)
		st.returnType = boolType

	case "or":
		checkType(lhsType, boolType)
		checkType(rhsType, boolType)
		st.returnType = boolType

	case "is":
		nonObjectTypes := []Type{intType, boolType, strType}
		if slices.Contains(nonObjectTypes, lhsType) ||
			slices.Contains(nonObjectTypes, rhsType) {
			log.Fatalf("Semantic Error: Expected both operands to be of object type")
		}
		st.returnType = boolType

	case "+", "-", "*", "//", "%":
		if binaryExpr.Op == "+" && lhsIsString && rhsIsString {
			st.returnType = strType
			return
		}
		if binaryExpr.Op == "+" && lhsIsList && rhsIsList {
			st.returnType = ListType{elemType: join(lhsType, rhsType)}
			return
		}
		checkType(lhsType, intType)
		checkType(rhsType, intType)

	case "<", "<=", ">", ">=", "==", "!=":
		st.returnType = boolType
		if lhsIsString && rhsIsString {
			return
		}
		if lhsIsList && rhsIsList {
			return
		}
		checkType(lhsType, intType)
		checkType(rhsType, intType)
	}
}

func (st *StaticTyping) VisitIfExpr(ifExpr *parser.IfExpr) {
	ifExpr.Condition.Visit(st)
	condType := st.returnType

	ifExpr.IfOp.Visit(st)
	ifOpType := st.returnType

	ifExpr.ElseOp.Visit(st)
	elseOpType := st.returnType

	checkType(condType, boolType)
	st.returnType = join(ifOpType, elseOpType)
}

func (st *StaticTyping) VisitListExpr(listExpr *parser.ListExpr) {
	if len(listExpr.Elements) == 0 {
		st.returnType = emptyType
		return
	}

	elemTypes := []Type{}
	for _, elem := range listExpr.Elements {
		elem.Visit(st)
		elemTypes = append(elemTypes, st.returnType)
	}

	joinedType := elemTypes[0]
	elemTypes = elemTypes[0:]
	for _, elemType := range elemTypes {
		joinedType = join(joinedType, elemType)
	}

	st.returnType = ListType{elemType: joinedType}
}

func (st *StaticTyping) VisitCallExpr(callExpr *parser.CallExpr) {
	funcName := callExpr.FuncName
	funcInfo, funcDefined := st.localEnv[funcName]

	if !funcDefined {
		log.Fatalf("Semantic Error: Unknown function identifier used: %s", funcName)
	}

	_, isFuncInfo := funcInfo.(FunctionInfo)
	if !isFuncInfo {
		log.Fatalf("Semantic Error: Found variable identifier: %s but expected function identifier", funcName)
	}

	if len(callExpr.Arguments) != len(funcInfo.(FunctionInfo).paramNames) {
		log.Fatalf("Semantic Error: Expected %d arguments but got %d", len(funcInfo.(FunctionInfo).paramNames), len(callExpr.Arguments))
	}

	for argIdx, argument := range callExpr.Arguments {
		argument.Visit(st)
		checkAssignmentCompatible(st.returnType, funcInfo.(FunctionInfo).funcType.paramTypes[argIdx])
	}

	st.returnType = funcInfo.(FunctionInfo).funcType.returnType
}

func (st *StaticTyping) VisitIndexExpr(indexExpr *parser.IndexExpr) {
	indexExpr.Value.Visit(st)
	valueType := st.returnType

	valueIsString := valueType == strType
	_, valueIsList := valueType.(ListType)

	indexExpr.Index.Visit(st)
	indexType := st.returnType

	checkType(indexType, intType)

	if valueIsString {
		st.returnType = strType
	} else if valueIsList {
		st.returnType = valueType.(ListType).elemType
	}
}
