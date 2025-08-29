package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitListExpr(listExpr *ast.ListExpr) {
	listElemType := cg.attrToType(listExpr.TypeHint.(ast.ListAttribute).ElemType)
	listPtr := cg.currentBlock.NewAlloca(listElemType)
	listPtr.LocalName = cg.uniqueNames.get("list_ptr")

	for elemIdx, elem := range listExpr.Elements {
		elem.Visit(cg)
		elemVal := cg.lastGenerated

		elemIdxConst := constant.NewInt(types.I32, int64(elemIdx))
		elemPtr := cg.currentBlock.NewGetElementPtr(listElemType, listPtr, elemIdxConst)
		elemPtr.LocalName = cg.uniqueNames.get("list_elem_ptr")

		cg.NewStore(elemVal, elemPtr)
	}

	cg.lengths[listPtr] = len(listExpr.Elements)
	cg.lastGenerated = listPtr
}
