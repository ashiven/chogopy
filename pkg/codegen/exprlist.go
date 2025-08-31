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
		// Think about it like this:
		//
		// - You have a pointer to a struct list: list*
		// - This points to a contiguous location in memory at which one or more list structs reside
		//
		// Memory(starting at list*):
		//
		// 							0:	list{content: i32*, size: i32}
		// elemIdx ->		1:	list{content: i32*, size: i32}
		// 							2:	list{content: i32*, size: i32}
		// 							3:	list{content: i32*, size: i32}
		//													^
		//													|
		//										contentIdx
		//
		// - Now you GEP will first need to know which of these lists to address (elemIdx)
		// - Then GEP will want to know which field of the list to address (contentIdx)
		//
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
