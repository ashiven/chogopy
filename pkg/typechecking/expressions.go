package typechecking

import (
	"chogopy/pkg/ast"
	"slices"
)

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
			semanticError(IsBinaryExpectedTwoObjectTypes, nil, nil, "", 0, 0)
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
		semanticError(FunctionCallArgumentMismatch, nil, nil, "", len(funcInfo.(FunctionInfo).paramNames), len(callExpr.Arguments))
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
