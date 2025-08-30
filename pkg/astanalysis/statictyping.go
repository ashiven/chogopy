package astanalysis

import (
	"chogopy/pkg/ast"
	"maps"
	"slices"
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
		typeSemanticError(UnknownIdentifierUsed, nil, nil, defName, 0, 0)
	}

	_, isFuncInfo := defType.(FunctionInfo)
	if isFuncInfo && expectVarDef {
		typeSemanticError(ExpectedVariableIdentifier, nil, nil, defName, 0, 0)
	}

	if !isFuncInfo && !expectVarDef {
		typeSemanticError(ExpectedFunctionIdentifier, nil, nil, defName, 0, 0)
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
	envBuilder := EnvironmentBuilder{}
	envBuilder.Build(program)

	st.localEnv = envBuilder.LocalEnv
	st.returnType = bottomType

	for _, definition := range program.Definitions {
		definition.Visit(st)
	}
	for _, statement := range program.Statements {
		statement.Visit(st)
	}
}

// Traverse determines whether the StaticTyping visitor should traverse further down the AST after a call to any nodes' Visit() method.
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

/* Statements */

func (st *StaticTyping) VisitIfStmt(ifStmt *ast.IfStmt) {
	ifStmt.Condition.Visit(st)
	checkType(st.visitedType, boolType)

	for _, ifBodyNode := range ifStmt.IfBody {
		ifBodyNode.Visit(st)
	}
	for _, elseBodyNode := range ifStmt.ElseBody {
		elseBodyNode.Visit(st)
	}
}

func (st *StaticTyping) VisitWhileStmt(whileStmt *ast.WhileStmt) {
	whileStmt.Condition.Visit(st)
	checkType(st.visitedType, boolType)

	for _, bodyNode := range whileStmt.Body {
		bodyNode.Visit(st)
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

	for _, bodyNode := range forStmt.Body {
		bodyNode.Visit(st)
	}
}

func (st *StaticTyping) VisitPassStmt(passStmt *ast.PassStmt) {}

func (st *StaticTyping) VisitReturnStmt(returnStmt *ast.ReturnStmt) {
	var returnType Type
	if returnStmt.ReturnVal != nil {
		returnStmt.ReturnVal.Visit(st)
		returnType = st.visitedType
	} else {
		returnType = noneType
	}
	checkAssignmentCompatible(returnType, st.returnType)
}

func (st *StaticTyping) VisitAssignStmt(assignStmt *ast.AssignStmt) {
	// Case 1: Multi-assign like: a = b = c --> Assign(a, Assign(b, c))
	_, valueIsAssign := assignStmt.Value.(*ast.AssignStmt)
	if valueIsAssign {

		// Substep 1: assignNodes collects every node in the assignment chain
		// ([a, b, c] for the above example)
		assignNodes := []ast.Node{assignStmt.Target}
		currentAssign := assignStmt

		for valueIsAssign {
			_, targetIsIdent := currentAssign.Target.(*ast.IdentExpr)
			_, targetIsIndex := currentAssign.Target.(*ast.IndexExpr)
			if !targetIsIdent && !targetIsIndex {
				typeSemanticError(AssignTargetInvalid, nil, nil, "", 0, 0)
			}

			currentAssign = assignStmt.Value.(*ast.AssignStmt)
			assignNodes = append(assignNodes, currentAssign.Target)
			_, valueIsAssign = currentAssign.Value.(*ast.AssignStmt)
		}
		assignNodes = append(assignNodes, currentAssign.Value)

		// Substep 2: After the nodes have been collected in assignNodes
		// assignment compatibility of the last node is checked against every
		// node preceding it in the assignment chain
		// ( c <a b , c <a a for the above example)
		assignNodes[len(assignNodes)-1].Visit(st)
		lastNodeType := st.visitedType

		_, lastNodeIsList := lastNodeType.(ListType)
		if lastNodeIsList && lastNodeType.(ListType).elemType == noneType {
			typeSemanticError(ExpectedNonNoneListType, nil, nil, "", 0, 0)
		}

		for _, assignNode := range assignNodes[:len(assignNodes)-1] {
			assignNode.Visit(st)
			assignNodeType := st.visitedType
			checkAssignmentCompatible(lastNodeType, assignNodeType)
		}

		// Substep 3: Type hints are added for each node in the assignment chain
		for _, assignNode := range assignNodes {
			switch assignNode := assignNode.(type) {
			case *ast.IdentExpr:
				identType := st.localEnv.check(assignNode.Identifier, true)
				assignNode.TypeHint = attrFromType(identType)
			case *ast.IndexExpr:
				assignNode.Value.Visit(st)
				valueType := st.visitedType
				assignNode.TypeHint = attrFromType(valueType)
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

		target.TypeHint = attrFromType(identType)

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

		target.TypeHint = attrFromType(targetValueType)

	// Assigning to anything that doesn't represent an identifier / index expression is illegal
	default:
		typeSemanticError(AssignTargetInvalid, nil, nil, "", 0, 0)
	}
}

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
	literalExpr.TypeHint = attrFromType(st.visitedType)
}

func (st *StaticTyping) VisitIdentExpr(identExpr *ast.IdentExpr) {
	varType := st.localEnv.check(identExpr.Identifier, true)
	st.visitedType = varType
	identExpr.TypeHint = attrFromType(st.visitedType)
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
	unaryExpr.TypeHint = attrFromType(st.visitedType)
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
		binaryExpr.TypeHint = attrFromType(st.visitedType)

	case "or":
		checkType(lhsType, boolType)
		checkType(rhsType, boolType)
		st.visitedType = boolType
		binaryExpr.TypeHint = attrFromType(st.visitedType)

	case "is":
		nonObjectTypes := []Type{intType, boolType, strType}
		if slices.Contains(nonObjectTypes, lhsType) ||
			slices.Contains(nonObjectTypes, rhsType) {
			typeSemanticError(IsBinaryExpectedTwoObjectTypes, nil, nil, "", 0, 0)
		}
		st.visitedType = boolType
		binaryExpr.TypeHint = attrFromType(st.visitedType)

	case "+", "-", "*", "//", "%":
		if binaryExpr.Op == "+" && lhsIsString && rhsIsString {
			st.visitedType = strType
			binaryExpr.TypeHint = attrFromType(st.visitedType)
			return
		}
		if binaryExpr.Op == "+" && lhsIsList && rhsIsList {
			st.visitedType = ListType{
				elemType: join(lhsType.(ListType).elemType, rhsType.(ListType).elemType),
			}
			binaryExpr.TypeHint = attrFromType(st.visitedType)
			return
		}
		checkType(lhsType, intType)
		checkType(rhsType, intType)
		st.visitedType = intType
		binaryExpr.TypeHint = attrFromType(st.visitedType)

	case "<", "<=", ">", ">=", "==", "!=":
		if lhsIsString && rhsIsString {
			st.visitedType = boolType
			binaryExpr.TypeHint = attrFromType(st.visitedType)
			return
		}
		if lhsIsList && rhsIsList {
			st.visitedType = boolType
			binaryExpr.TypeHint = attrFromType(st.visitedType)
			return
		}
		if lhsType == boolType && rhsType == boolType {
			st.visitedType = boolType
			binaryExpr.TypeHint = attrFromType(st.visitedType)
			return
		}
		checkType(lhsType, intType)
		checkType(rhsType, intType)
		st.visitedType = boolType
		binaryExpr.TypeHint = attrFromType(st.visitedType)
	}
}

func (st *StaticTyping) VisitIfExpr(ifExpr *ast.IfExpr) {
	ifExpr.Condition.Visit(st)
	condType := st.visitedType

	ifExpr.IfNode.Visit(st)
	ifNodeType := st.visitedType

	ifExpr.ElseNode.Visit(st)
	elseNodeType := st.visitedType

	checkType(condType, boolType)
	st.visitedType = join(ifNodeType, elseNodeType)
	ifExpr.TypeHint = attrFromType(st.visitedType)
}

func (st *StaticTyping) VisitListExpr(listExpr *ast.ListExpr) {
	if len(listExpr.Elements) == 0 {
		st.visitedType = emptyType
		listExpr.TypeHint = attrFromType(st.visitedType)
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
	listExpr.TypeHint = attrFromType(st.visitedType)
}

func (st *StaticTyping) VisitCallExpr(callExpr *ast.CallExpr) {
	funcName := callExpr.FuncName
	funcInfo := st.localEnv.check(funcName, false)

	if len(callExpr.Arguments) != len(funcInfo.(FunctionInfo).paramNames) {
		typeSemanticError(FunctionCallArgumentMismatch, nil, nil, "", len(funcInfo.(FunctionInfo).paramNames), len(callExpr.Arguments))
	}

	for argIdx, argument := range callExpr.Arguments {
		argument.Visit(st)
		checkAssignmentCompatible(st.visitedType, funcInfo.(FunctionInfo).funcType.paramTypes[argIdx])
	}

	st.visitedType = funcInfo.(FunctionInfo).funcType.returnType
	callExpr.TypeHint = attrFromType(st.visitedType)
}

func (st *StaticTyping) VisitIndexExpr(indexExpr *ast.IndexExpr) {
	indexExpr.Value.Visit(st)
	valueType := st.visitedType

	indexExpr.Index.Visit(st)
	indexType := st.visitedType

	checkType(indexType, intType)

	if valueType == strType {
		st.visitedType = strType
		indexExpr.TypeHint = attrFromType(st.visitedType)
	} else {
		checkListType(valueType)
		st.visitedType = valueType.(ListType).elemType
		indexExpr.TypeHint = attrFromType(st.visitedType)
	}
}
