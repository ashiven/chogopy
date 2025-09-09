package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitListExpr(listExpr *ast.ListExpr) {
	listPtr := cg.newConstantList(listExpr)

	cg.lastGenerated = listPtr
}

func (cg *CodeGenerator) newConstantList(listExpr *ast.ListExpr) constant.Constant {
	listLenU64 := uint64(len(listExpr.Elements))

	// attrToType will return something like:
	// list{content: i32*, size: i32, init: i1}*
	//
	// So we want to take out the elemType:
	// list{content: i32*, size: i32, init: i1}
	listType := cg.attrToType(listExpr.TypeHint).(*types.PointerType).ElemType.(*types.StructType)

	// For the example above, listElemType will be: i32  (derived from content: i32*)
	listElemType := getListElemTypeFromListType(listType)

	/* list.size and list.init */
	listLen := constant.NewInt(types.I32, int64(listLenU64))
	listInit := constant.NewBool(true)

	listElems := cg.getListElems(listExpr)

	zero := constant.NewInt(types.I32, 0)

	listContent := constant.NewArray(types.NewArray(listLenU64, listElemType), listElems...)
	listContentDef := cg.Module.NewGlobalDef(cg.uniqueNames.get("list_content"), listContent)

	listContentPtr := constant.NewGetElementPtr(listContentDef.Typ.ElemType, listContentDef, zero, zero)

	listDef := cg.Module.NewGlobalDef(
		cg.uniqueNames.get("list_literal"),
		constant.NewStruct(listType, listContentPtr, listLen, listInit),
	)
	listPtr := constant.NewGetElementPtr(listType, listDef, zero)

	return listPtr
}

func (cg *CodeGenerator) getListElems(listExpr *ast.ListExpr) []constant.Constant {
	// NOTE: We are assuming that all list literals will contain nothing but
	// literal expressions like: [1,2,3] or ["a","b","c"] as opposed to: [var1,var2,var3] or [v[0],v[1],v[3]]
	// This is because we will allocate list literals statically at compile time (using global definitions)
	// instead of dynamically at runtime on the stack or the heap.
	// This limits the expressiveness of this already limited language even more but I am currently
	// unsure how to go about returning either a copy of the call stack which the caller may be able to
	// use to create a copy of the list or how to allocate memory on the heap that will be
	// freed again without the programmer explicitly having to call a freeing function.
	listElems := []constant.Constant{}

	for _, elem := range listExpr.Elements {
		// According to our assumption above, an elem can either be another list literal or a literal expression
		switch elem := elem.(type) {
		case *ast.ListExpr:
			listConstant := cg.newConstantList(elem)
			listElems = append(listElems, listConstant)

		case *ast.LiteralExpr:
			elemType := cg.attrToType(elem.TypeHint)
			switch elemType {
			case types.I1:
				listElems = append(listElems, constant.NewBool(elem.Value.(bool)))
			case types.I32:
				listElems = append(listElems, constant.NewInt(types.I32, int64(elem.Value.(int))))
			case types.I8Ptr:
				strConst := constant.NewCharArrayFromString(elem.Value.(string) + "\x00")
				strDef := cg.Module.NewGlobalDef(cg.uniqueNames.get("str"), strConst)

				zero := constant.NewInt(types.I32, 0)
				strPtr := constant.NewGetElementPtr(strDef.Typ.ElemType, strDef, zero, zero)

				listElems = append(listElems, strPtr)
			}
		}
	}

	return listElems
}
