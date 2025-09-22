package codegen

import (
	"chogopy/src/ast"
)

func (cg *CodeGenerator) VisitAssignStmt(assignStmt *ast.AssignStmt) {
	/* In the case of a multi assign stmt like: a = b = 3
	 * we first collect all nodes in a slice and then perform the assignment as follows:
	 *
	 * 1) visit the node that should be assigned and generate its value (
	 *  visit 3
	 *
	 * 2) assign the value to each target sequentially via iteration over assignTargets:
	 *  a = 3 -> b = 3
	 *
	 *  */
	assignNodes := flattenAssign(assignStmt)
	assignTargets := assignNodes[:len(assignNodes)-1]
	assignValue := assignNodes[len(assignNodes)-1]

	assignValue.Visit(cg)
	value := cg.lastGenerated

	if isIdentOrIndex(assignValue) {
		value = cg.LoadVal(value)
	}

	for _, assignTarget := range assignTargets {
		assignTarget.Visit(cg)
		target := cg.lastGenerated

		cg.NewStore(value, target)
	}
}

func flattenAssign(assignStmt *ast.AssignStmt) []ast.Node {
	assignNodes := []ast.Node{assignStmt.Target}
	currentAssign := assignStmt

	if _, valueIsAssign := assignStmt.Value.(*ast.AssignStmt); valueIsAssign {
		for valueIsAssign {
			currentAssign = assignStmt.Value.(*ast.AssignStmt)
			assignNodes = append(assignNodes, currentAssign.Target)
			_, valueIsAssign = currentAssign.Value.(*ast.AssignStmt)
		}
	}
	assignNodes = append(assignNodes, currentAssign.Value)

	return assignNodes
}
