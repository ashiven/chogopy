package codegen

import "chogopy/pkg/ast"

func (cg *CodeGenerator) VisitAssignStmt(assignStmt *ast.AssignStmt) {
	assignStmt.Target.Visit(cg)
	target := cg.lastGenerated

	assignStmt.Value.Visit(cg)
	value := cg.lastGenerated

	if isIdentOrIndex(assignStmt.Value) {
		value = cg.LoadVal(value)
	}

	cg.NewStore(value, target)
}
