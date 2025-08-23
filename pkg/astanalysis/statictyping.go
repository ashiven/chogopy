package astanalysis

import (
	"chogopy/pkg/ast"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
)

type SemanticErrorKind int

const (
	NotAssignmentCompatible SemanticErrorKind = iota
	UnexpectedType
	ExpectedListType
	UnknownIdentifierUsed
	ExpectedVariableIdentifier
	ExpectedFunctionIdentifier
	ExpectedNonNoneListType
	IsBinaryExpectedTwoObjectTypes
	FunctionCallArgumentMismatch
)

func semanticError(errorKind SemanticErrorKind, t1 Type, t2 Type, defName string, funcArgs int, callArgs int) {
	switch errorKind {
	case NotAssignmentCompatible:
		fmt.Printf("Semantic Error: %s is not assignment compatible with %s", nameFromType(t1), nameFromType(t2))
	case UnexpectedType:
		fmt.Printf("Semantic Error: Expected %s but found %s", nameFromType(t1), nameFromType(t2))
	case ExpectedListType:
		fmt.Printf("Semantic Error: Expected list type but found %s", nameFromType(t1))
	case UnknownIdentifierUsed:
		fmt.Printf("Semantic Error: Unknown identifier used: %s", defName)
	case ExpectedVariableIdentifier:
		fmt.Printf("Semantic Error: Found function identifier: %s but expected variable identifier", defName)
	case ExpectedFunctionIdentifier:
		fmt.Printf("Semantic Error: Found variable identifier: %s but expected function identifier", defName)
	case ExpectedNonNoneListType:
		fmt.Printf("Semantic Error: Expected non-none list type in asssignment")
	case IsBinaryExpectedTwoObjectTypes:
		fmt.Printf("Semantic Error: Expected both operands to be of object type")
	case FunctionCallArgumentMismatch:
		fmt.Printf("Semantic Error: Expected %d arguments but got %d\n", funcArgs, callArgs)
	}
	os.Exit(0)
}

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
	typeName string
	Type
}

type ObjectType struct {
	typeName string
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
	bottomType = BottomType{typeName: "bottom"}
	objectType = ObjectType{typeName: "object"}
)

func typeFromOp(op ast.Operation) Type {
	switch op := op.(type) {
	case *ast.NamedType:
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

	case *ast.ListType:
		elemType := typeFromOp(op.ElemType)
		return ListType{elemType: elemType}
	}

	log.Fatalf("Expected Operation but found %# v", op)
	return nil
}

func hintFromType(opType Type) ast.Operation {
	switch opType {
	case intType:
		return &ast.NamedType{TypeName: "int"}
	case boolType:
		return &ast.NamedType{TypeName: "bool"}
	case strType:
		return &ast.NamedType{TypeName: "str"}
	case noneType:
		return &ast.NamedType{TypeName: "<None>"}
	case emptyType:
		return &ast.NamedType{TypeName: "<Empty>"}
	case objectType:
		return &ast.NamedType{TypeName: "object"}
	}

	_, isListType := opType.(ListType)
	if isListType {
		elemType := hintFromType(opType.(ListType).elemType)
		return &ast.ListType{ElemType: elemType}
	}

	log.Fatalf("Expected Type but found %# v", opType)
	return nil
}

func nameFromType(opType Type) string {
	switch opType {
	case intType:
		return intType.typeName
	case boolType:
		return boolType.typeName
	case strType:
		return strType.typeName
	case noneType:
		return noneType.typeName
	case emptyType:
		return emptyType.typeName
	case objectType:
		return objectType.typeName
	case bottomType:
		return bottomType.typeName
	}

	_, isListType := opType.(ListType)
	if isListType {
		elemType := nameFromType(opType.(ListType).elemType)
		return fmt.Sprintf("List[%s]", elemType)
	}

	log.Fatalf("Expected Type but found %# v", opType)
	return ""
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
		semanticError(NotAssignmentCompatible, t1, t2, "", 0, 0)
	}
}

func checkType(found Type, expected Type) {
	if found != expected {
		semanticError(UnexpectedType, expected, found, "", 0, 0)
	}
}

func checkListType(found Type) {
	_, foundIsList := found.(ListType)
	if !foundIsList {
		semanticError(ExpectedListType, found, nil, "", 0, 0)
	}
}

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
	LocalEnvironment LocalEnvironment
	ast.BaseVisitor
}

func (eb *EnvironmentBuilder) Analyze(program *ast.Program) {
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

func (eb *EnvironmentBuilder) VisitTypedVar(typedVar *ast.TypedVar) {
	varName := typedVar.VarName
	varType := typeFromOp(typedVar.VarType)
	eb.LocalEnvironment[varName] = varType
}

func (eb *EnvironmentBuilder) VisitFuncDef(funcDef *ast.FuncDef) {
	funcName := funcDef.FuncName

	paramNames := []string{}
	paramTypes := []Type{}
	for _, param := range funcDef.Parameters {
		paramName := param.(*ast.TypedVar).VarName
		paramType := typeFromOp(param.(*ast.TypedVar).VarType)
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
	ast.BaseVisitor
}

// Analyze performs static type checking according to the rules defined in
// chapter 5 of the chocopy language reference:
//
// https://chocopy.org/chocopy_language_reference.pdf
func (st *StaticTyping) Analyze(program *ast.Program) {
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
	for _, funcBodyOp := range funcDef.FuncBody {
		funcBodyOp.Visit(funcBodyVisitor)
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

/* Statements */

func (st *StaticTyping) VisitIfStmt(ifStmt *ast.IfStmt) {
	ifStmt.Condition.Visit(st)
	checkType(st.visitedType, boolType)

	for _, ifBodyOp := range ifStmt.IfBody {
		ifBodyOp.Visit(st)
	}
	for _, elseBodyOp := range ifStmt.ElseBody {
		elseBodyOp.Visit(st)
	}
}

func (st *StaticTyping) VisitWhileStmt(whileStmt *ast.WhileStmt) {
	whileStmt.Condition.Visit(st)
	checkType(st.visitedType, boolType)

	for _, bodyOp := range whileStmt.Body {
		bodyOp.Visit(st)
	}
}

func (st *StaticTyping) VisitForStmt(forStmt *ast.ForStmt) {
	forStmt.Iter.Visit(st)
	iterType := st.visitedType
	iterNameType := st.localEnv.check(forStmt.IterName, true)

	if iterType == strType {
		checkAssignmentCompatible(strType, iterNameType)
	} else {
		checkListType(iterType)
		elemType := iterType.(ListType).elemType
		checkAssignmentCompatible(elemType, iterNameType)
	}

	for _, bodyOp := range forStmt.Body {
		bodyOp.Visit(st)
	}
}

func (st *StaticTyping) VisitPassStmt(passStmt *ast.PassStmt) {}

func (st *StaticTyping) VisitReturnStmt(returnStmt *ast.ReturnStmt) {
	returnStmt.ReturnVal.Visit(st)
	returnType := st.visitedType
	checkAssignmentCompatible(returnType, st.returnType)
}

func (st *StaticTyping) VisitAssignStmt(assignStmt *ast.AssignStmt) {
	// Case 1: Multi-assign like: a = b = c --> Assign(a, Assign(b, c))
	_, valueIsAssign := assignStmt.Value.(*ast.AssignStmt)
	if valueIsAssign {

		// Substep 1: assignOps collects every op in the assignment chain
		// ([a, b, c] for the above example)
		assignOps := []ast.Operation{assignStmt.Target}
		currentAssign := assignStmt

		for valueIsAssign {
			currentAssign = assignStmt.Value.(*ast.AssignStmt)
			assignOps = append(assignOps, currentAssign.Target)
			_, valueIsAssign = currentAssign.Value.(*ast.AssignStmt)
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
			semanticError(ExpectedNonNoneListType, nil, nil, "", 0, 0)
		}

		for _, assignOp := range assignOps[:len(assignOps)-1] {
			assignOp.Visit(st)
			assignOpType := st.visitedType
			checkAssignmentCompatible(lastOpType, assignOpType)
		}

		// Substep 3: Type hints are added for each op in the assignment chain
		for _, assignOp := range assignOps {
			switch assignOp := assignOp.(type) {
			case *ast.IdentExpr:
				identType := st.localEnv.check(assignOp.Identifier, true)
				assignOp.TypeHint = hintFromType(identType)
			case *ast.IndexExpr:
				assignOp.Value.Visit(st)
				valueType := st.visitedType
				assignOp.TypeHint = hintFromType(valueType)
			}
		}
	}

	switch target := assignStmt.Target.(type) {
	// Case 2: Assign to an identifier like: a = 1
	case *ast.IdentExpr:
		identName := target.Identifier
		identType := st.localEnv.check(identName, true)

		assignStmt.Value.Visit(st)
		valueType := st.visitedType

		checkAssignmentCompatible(valueType, identType)

		target.TypeHint = hintFromType(identType)

	// Case 3: Assign to a list like: a[12] = 1
	case *ast.IndexExpr:
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

// TODO: all expressions should be equipped with a type hint
/* Expressions */

func (st *StaticTyping) VisitLiteralExpr(literalExpr *ast.LiteralExpr) {
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

func (st *StaticTyping) VisitIdentExpr(identExpr *ast.IdentExpr) {
	varType := st.localEnv.check(identExpr.Identifier, true)
	st.visitedType = varType
}

func (st *StaticTyping) VisitUnaryExpr(unaryExpr *ast.UnaryExpr) {
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

func (st *StaticTyping) VisitBinaryExpr(binaryExpr *ast.BinaryExpr) {
	binaryExpr.Lhs.Visit(st)
	lhsType := st.visitedType

	binaryExpr.Rhs.Visit(st)
	rhsType := st.visitedType

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
			semanticError(IsBinaryExpectedTwoObjectTypes, nil, nil, "", 0, 0)
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

func (st *StaticTyping) VisitIfExpr(ifExpr *ast.IfExpr) {
	ifExpr.Condition.Visit(st)
	condType := st.visitedType

	ifExpr.IfOp.Visit(st)
	ifOpType := st.visitedType

	ifExpr.ElseOp.Visit(st)
	elseOpType := st.visitedType

	checkType(condType, boolType)
	st.visitedType = join(ifOpType, elseOpType)
}

func (st *StaticTyping) VisitListExpr(listExpr *ast.ListExpr) {
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

	st.visitedType = ListType{elemType: joinedType}
}

func (st *StaticTyping) VisitCallExpr(callExpr *ast.CallExpr) {
	funcName := callExpr.FuncName
	funcInfo := st.localEnv.check(funcName, false)

	if len(callExpr.Arguments) != len(funcInfo.(FunctionInfo).paramNames) {
		semanticError(FunctionCallArgumentMismatch, nil, nil, "", len(funcInfo.(FunctionInfo).paramNames), len(callExpr.Arguments))
	}

	for argIdx, argument := range callExpr.Arguments {
		argument.Visit(st)
		checkAssignmentCompatible(st.visitedType, funcInfo.(FunctionInfo).funcType.paramTypes[argIdx])
	}

	st.visitedType = funcInfo.(FunctionInfo).funcType.returnType
}

func (st *StaticTyping) VisitIndexExpr(indexExpr *ast.IndexExpr) {
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
