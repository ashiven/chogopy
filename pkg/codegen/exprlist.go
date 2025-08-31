package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitListExpr(listExpr *ast.ListExpr) {
	/* list alloc */
	listType := cg.attrToType(listExpr.TypeHint).(*types.PointerType).ElemType
	listPtr := cg.currentBlock.NewAlloca(listType)
	listPtr.LocalName = cg.uniqueNames.get("list_ptr")

	/* list.content alloc */
	listElemType := cg.attrToType(listExpr.TypeHint.(ast.ListAttribute).ElemType)
	listContentPtr := cg.currentBlock.NewAlloca(listElemType)
	listContentPtr.LocalName = cg.uniqueNames.get("list_content_ptr")

	/* list.content init */
	for elemIdx, elem := range listExpr.Elements {
		elem.Visit(cg)
		elemVal := cg.lastGenerated

		elemIdx := constant.NewInt(types.I32, int64(elemIdx))
		elemPtr := cg.currentBlock.NewGetElementPtr(listElemType, listContentPtr, elemIdx)

		// TODO: nested list check
		// In case the content of the list is a pointer to another list (lists have a struct type)
		// the first GEP index will select the right list struct and the second index will select the field to store into (list.content)
		if false {
			contentIdx := constant.NewInt(types.I32, 0)
			elemPtr = cg.currentBlock.NewGetElementPtr(listElemType, listContentPtr, elemIdx, contentIdx)
		}

		elemPtr.LocalName = cg.uniqueNames.get("list_content_elem_addr")
		cg.NewStore(elemVal, elemPtr)
	}

	/* list.size init */
	listSize := constant.NewInt(types.I32, int64(len(listExpr.Elements)))

	/* list init */
	zero := constant.NewInt(types.I32, 0)
	one := constant.NewInt(types.I32, 1)
	listContentAddr := cg.currentBlock.NewGetElementPtr(listType, listPtr, zero, zero)
	listContentAddr.LocalName = cg.uniqueNames.get("list_content_addr")
	listSizeAddr := cg.currentBlock.NewGetElementPtr(listType, listPtr, zero, one)
	listSizeAddr.LocalName = cg.uniqueNames.get("list_size_addr")
	cg.NewStore(listContentPtr, listContentAddr)
	cg.NewStore(listSize, listSizeAddr)

	cg.lengths[listPtr] = len(listExpr.Elements)

	cg.lastGenerated = listPtr
}
