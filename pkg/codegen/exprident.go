package codegen

import (
	"chogopy/pkg/ast"
	"log"
)

func (cg *CodeGenerator) VisitIdentExpr(identExpr *ast.IdentExpr) {
	identName := identExpr.Identifier

	// Case 1: The identifier refers to a variable definition.

	identVarInfo, err := cg.getVar(identName)
	if err != nil {
		log.Fatalln(err.Error())
	}

	cg.lastGenerated = identVarInfo.value

	// Case 2: The identifier refers to the name of a parameter of the current function. (overwrites definitions)

	for _, param := range cg.currentFunction.Params {
		if identName == param.LocalName {
			cg.lastGenerated = param
		}
	}
}
