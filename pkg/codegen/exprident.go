package codegen

import "chogopy/pkg/ast"

func (cg *CodeGenerator) VisitIdentExpr(identExpr *ast.IdentExpr) {
	identName := identExpr.Identifier

	// Case 1: The identifier refers to a global var def.
	if variable, ok := cg.variables[identName]; ok {
		cg.lastGenerated = variable.value
	}

	// Case 2: The identifier refers to the name of a parameter of the current function. (overwrites global def)
	for _, param := range cg.currentFunction.Params {
		if identName == param.LocalName {
			cg.lastGenerated = param
		}
	}
}
