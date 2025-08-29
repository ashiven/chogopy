package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitReturnStmt(returnStmt *ast.ReturnStmt) {
	var returnVal value.Value
	if returnStmt.ReturnVal != nil {
		returnStmt.ReturnVal.Visit(cg)
		returnVal = cg.lastGenerated
	}

	cg.currentBlock.NewRet(returnVal)
}
