package astanalysis

import (
	"chogopy/pkg/parser"
	"log"
	"maps"
	"slices"

	"github.com/kr/pretty"
)

type Type any

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

	log.Fatalf("Expected Operation but found %# v", op)
	return nil
}

func hintFromType(opType Type) parser.Operation {
	switch opType {
	case intType:
		return &parser.NamedType{TypeName: "int"}
	case boolType:
		return &parser.NamedType{TypeName: "bool"}
	case strType:
		return &parser.NamedType{TypeName: "str"}
	case noneType:
		return &parser.NamedType{TypeName: "<None>"}
	case emptyType:
		return &parser.NamedType{TypeName: "<Empty>"}
	case objectType:
		return &parser.NamedType{TypeName: "object"}
	}

	_, isListType := opType.(ListType)
	if isListType {
		elemType := hintFromType(opType.(ListType).elemType)
		return &parser.ListType{ElemType: elemType}
	}

	log.Fatalf("Expected Type but found %# v", opType)
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
		log.Fatalf("Semantic Error: '%# v' is not assignment compatible with '%# v'", t1, t2)
	}
}

func checkType(found Type, expected Type) {
	if found != expected {
		log.Fatalf("Semantic Error: Expected '%# v' but found '%# v'", expected, found)
	}
}

func checkListType(found Type) {
	_, foundIsList := found.(ListType)
	if !foundIsList {
		log.Fatalf("Semantic Error: Expected list type but found '%# v'", found)
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

func (le LocalEnvironment) check(defName string, expectVarDef bool) DefType {
	defType, defExists := le[defName]
	if !defExists {
		log.Fatalf("Semantic Error: Unknown identifier used: %s", pretty.Formatter(defName))
	}

	_, isFuncInfo := defType.(FunctionInfo)
	if isFuncInfo && expectVarDef {
		log.Fatalf("Semantic Error: Found function identifier: %s but expected variable identifier", pretty.Formatter(defName))
	}

	if !isFuncInfo && !expectVarDef {
		log.Fatalf("Semantic Error: Found variable identifier: %s but expected function identifier", pretty.Formatter(defName))
	}

	return defType
}

// EnvironmentBuilder is responsible for traversing the AST and constructing the above-defined
// LocalEnvironment by checking every VarDef and FuncDef AST node
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
	localEnv    LocalEnvironment
	returnType  Type
	visitedType Type
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

	for _, definition := range program.Definitions {
		definition.Visit(st)
	}
	for _, statement := range program.Statements {
		statement.Visit(st)
	}
}

// Traverse determines whether the StaticTyping visitor should traverse further down the AST
// at any given moment.
// If this returns true, each node will implicitly call the Visit() method on each of its
// child nodes whenever it is visited, which is something we want to avoid here.
// Instead, we will define when exactly a node will visit its child nodes via explicit calls
// to the child nodes' Visit() methods in each of the Visitors interface methods.
func (st *StaticTyping) Traverse() bool {
	return false
}

/* Definitions */

func (st *StaticTyping) VisitFuncDef(funcDef *parser.FuncDef) {
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
	for _, funcBodyOp := range funcDef.FuncBody {
		funcBodyOp.Visit(funcBodyVisitor)
	}
}

func (st *StaticTyping) VisitGlobalDecl(globalDecl *parser.GlobalDecl) {}

func (st *StaticTyping) VisitNonLocalDecl(nonLocalDecl *parser.NonLocalDecl) {}

func (st *StaticTyping) VisitVarDef(varDef *parser.VarDef) {
	varName := varDef.TypedVar.(*parser.TypedVar).VarName
	varType := st.localEnv.check(varName, true)

	varDef.Literal.Visit(st)
	literalType := st.visitedType

	checkAssignmentCompatible(literalType, varType)
}

/* Statements */

func (st *StaticTyping) VisitIfStmt(ifStmt *parser.IfStmt) {
	ifStmt.Condition.Visit(st)
	checkType(st.visitedType, boolType)
}

func (st *StaticTyping) VisitWhileStmt(whileStmt *parser.WhileStmt) {
	whileStmt.Condition.Visit(st)
	checkType(st.visitedType, boolType)
}

func (st *StaticTyping) VisitForStmt(forStmt *parser.ForStmt) {
	forStmt.Iter.Visit(st)
	iterType := st.visitedType

	iterIsString := iterType == strType
	_, iterIsList := iterType.(ListType)

	iterNameType := st.localEnv.check(forStmt.IterName, true)

	if iterIsString {
		checkAssignmentCompatible(strType, iterNameType)
	} else if iterIsList {
		elemType := iterType.(ListType).elemType
		checkAssignmentCompatible(elemType, iterNameType)
	}
}

func (st *StaticTyping) VisitPassStmt(passStmt *parser.PassStmt) {}

func (st *StaticTyping) VisitReturnStmt(returnStmt *parser.ReturnStmt) {
	returnStmt.ReturnVal.Visit(st)
	returnType := st.visitedType
	checkAssignmentCompatible(returnType, st.returnType)
}

func (st *StaticTyping) VisitAssignStmt(assignStmt *parser.AssignStmt) {
	// Case 1: Multi-assign like: a = b = c --> Assign(a, Assign(b, c))
	_, valueIsAssign := assignStmt.Value.(*parser.AssignStmt)
	if valueIsAssign {

		// Substep 1: assignOps collects every op in the assignment chain
		// ([a, b, c] for the above example)
		assignOps := []parser.Operation{assignStmt.Target}
		currentAssign := assignStmt

		for valueIsAssign {
			currentAssign = assignStmt.Value.(*parser.AssignStmt)
			assignOps = append(assignOps, currentAssign.Target)
			_, valueIsAssign = currentAssign.Value.(*parser.AssignStmt)
		}
		assignOps = append(assignOps, currentAssign.Value)

		// Substep 2: After the ops have been collected in assignOps
		// assignment compatibility of the last op is checked against every
		// op preceding it in the assignment chain
		// ( c <a b , c <a a for the above example)
		assignOps[len(assignOps)-1].Visit(st)
		lastOpType := st.visitedType

		_, lastOpIsList := lastOpType.(ListType)
		if lastOpIsList && lastOpType.(ListType).elemType == noneType {
			log.Fatalf("Semantic Error: Expected non-none list type in asssignment")
		}

		for _, assignOp := range assignOps[:len(assignOps)-1] {
			assignOp.Visit(st)
			assignOpType := st.visitedType
			checkAssignmentCompatible(lastOpType, assignOpType)
		}

		// Substep 3: Type hints are added for each op in the assignment chain
		for _, assignOp := range assignOps {
			switch assignOp := assignOp.(type) {
			case *parser.IdentExpr:
				identType := st.localEnv.check(assignOp.Identifier, true)
				assignOp.TypeHint = hintFromType(identType)
			case *parser.IndexExpr:
				assignOp.Value.Visit(st)
				valueType := st.visitedType
				assignOp.TypeHint = hintFromType(valueType)
			}
		}
	}

	switch target := assignStmt.Target.(type) {
	// Case 2: Assign to an identifier like: a = 1
	case *parser.IdentExpr:
		identName := target.Identifier
		identType := st.localEnv.check(identName, true)

		assignStmt.Value.Visit(st)
		valueType := st.visitedType

		checkAssignmentCompatible(valueType, identType)

		target.TypeHint = hintFromType(identType)

	// Case 3: Assign to a list like: a[12] = 1
	case *parser.IndexExpr:
		target.Value.Visit(st)
		targetValueType := st.visitedType
		checkListType(targetValueType)

		target.Index.Visit(st)
		targetIndexType := st.visitedType
		checkType(targetIndexType, intType)

		assignStmt.Value.Visit(st)
		valueType := st.visitedType
		checkAssignmentCompatible(valueType, targetValueType.(ListType).elemType)

		target.TypeHint = hintFromType(targetValueType)
	}
}

/* Expressions */

func (st *StaticTyping) VisitLiteralExpr(literalExpr *parser.LiteralExpr) {
	switch literalExpr.Value.(type) {
	case int:
		st.visitedType = intType
	case bool:
		st.visitedType = boolType
	case string:
		st.visitedType = strType
	default:
		st.visitedType = noneType
	}
}

func (st *StaticTyping) VisitIdentExpr(identExpr *parser.IdentExpr) {
	varType := st.localEnv.check(identExpr.Identifier, true)
	st.visitedType = varType
}

func (st *StaticTyping) VisitUnaryExpr(unaryExpr *parser.UnaryExpr) {
	unaryExpr.Value.Visit(st)

	switch unaryExpr.Op {
	case "-":
		checkType(st.visitedType, intType)
		st.visitedType = intType
	case "not":
		checkType(st.visitedType, boolType)
		st.visitedType = boolType
	}
}

func (st *StaticTyping) VisitBinaryExpr(binaryExpr *parser.BinaryExpr) {
	pretty.Println("Parsing lhs type")
	binaryExpr.Lhs.Visit(st)
	lhsType := st.visitedType
	pretty.Println("Lhs parsed type:")
	pretty.Println(lhsType)

	pretty.Println("Parsing rhs type")
	binaryExpr.Rhs.Visit(st)
	rhsType := st.visitedType
	pretty.Println("Rhs parsed type:")
	pretty.Println(rhsType)

	_, lhsIsList := lhsType.(ListType)
	_, rhsIsList := rhsType.(ListType)

	lhsIsString := lhsType == strType
	rhsIsString := rhsType == strType

	switch binaryExpr.Op {
	case "and":
		checkType(lhsType, boolType)
		checkType(rhsType, boolType)
		st.visitedType = boolType

	case "or":
		checkType(lhsType, boolType)
		checkType(rhsType, boolType)
		st.visitedType = boolType

	case "is":
		nonObjectTypes := []Type{intType, boolType, strType}
		if slices.Contains(nonObjectTypes, lhsType) ||
			slices.Contains(nonObjectTypes, rhsType) {
			log.Fatalf("Semantic Error: Expected both operands to be of object type")
		}
		st.visitedType = boolType

	case "+", "-", "*", "//", "%":
		if binaryExpr.Op == "+" && lhsIsString && rhsIsString {
			st.visitedType = strType
			return
		}
		if binaryExpr.Op == "+" && lhsIsList && rhsIsList {
			st.visitedType = ListType{elemType: join(lhsType, rhsType)}
			return
		}
		checkType(lhsType, intType)
		checkType(rhsType, intType)

	case "<", "<=", ">", ">=", "==", "!=":
		st.visitedType = boolType
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
	condType := st.visitedType

	ifExpr.IfOp.Visit(st)
	ifOpType := st.visitedType

	ifExpr.ElseOp.Visit(st)
	elseOpType := st.visitedType

	checkType(condType, boolType)
	st.visitedType = join(ifOpType, elseOpType)
}

func (st *StaticTyping) VisitListExpr(listExpr *parser.ListExpr) {
	if len(listExpr.Elements) == 0 {
		st.visitedType = emptyType
		return
	}

	elemTypes := []Type{}
	for _, elem := range listExpr.Elements {
		elem.Visit(st)
		elemTypes = append(elemTypes, st.visitedType)
	}

	joinedType := elemTypes[0]
	elemTypes = elemTypes[0:]
	for _, elemType := range elemTypes {
		joinedType = join(joinedType, elemType)
	}

	// TODO:
	// After we have set such a visitedType we do not want the StaticTyping visitor to continue traversing down the
	// AST (which is what will happen because the ListExpr has child nodes
	// that will also be traversed and which may change the st.visitedType to something else that the caller doesn't expect).
	// Therefore we may have to set some kind of a stopTraversing property on the StaticTyping
	// visitor which is checked for in each node's Visit() method and will prevent them
	// from calling the Visit() method on their child nodes as would otherwise be the case.
	// Note that this property will have to be set to false again before the next call to Visit().
	st.visitedType = ListType{elemType: joinedType}
}

func (st *StaticTyping) VisitCallExpr(callExpr *parser.CallExpr) {
	funcName := callExpr.FuncName
	funcInfo := st.localEnv.check(funcName, false)

	if len(callExpr.Arguments) != len(funcInfo.(FunctionInfo).paramNames) {
		log.Fatalf("Semantic Error: Expected %d arguments but got %d", len(funcInfo.(FunctionInfo).paramNames), len(callExpr.Arguments))
	}

	for argIdx, argument := range callExpr.Arguments {
		argument.Visit(st)
		checkAssignmentCompatible(st.visitedType, funcInfo.(FunctionInfo).funcType.paramTypes[argIdx])
	}

	st.visitedType = funcInfo.(FunctionInfo).funcType.returnType
}

func (st *StaticTyping) VisitIndexExpr(indexExpr *parser.IndexExpr) {
	indexExpr.Value.Visit(st)
	valueType := st.visitedType

	valueIsString := valueType == strType
	_, valueIsList := valueType.(ListType)

	indexExpr.Index.Visit(st)
	indexType := st.visitedType

	checkType(indexType, intType)

	if valueIsString {
		st.visitedType = strType
	} else if valueIsList {
		st.visitedType = valueType.(ListType).elemType
	}
}
