package typechecking

import "chogopy/pkg/ast"

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
