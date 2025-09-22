package codegen

import (
	"chogopy/src/ast"

	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitReturnStmt(returnStmt *ast.ReturnStmt) {
	var returnVal value.Value

	if returnStmt.ReturnVal != nil {
		returnStmt.ReturnVal.Visit(cg)
		returnVal = cg.lastGenerated

		if isIdentOrIndex(returnStmt.ReturnVal) {
			returnVal = cg.LoadVal(returnVal)
		}
	}

	cg.currentBlock.NewRet(returnVal)
}
