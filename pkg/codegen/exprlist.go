package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitListExpr(listExpr *ast.ListExpr) {
	// attrToType will return something like:
	//
	// list{content: i32*, size: i32, init: i1}*
	//
	//
	// So we want to take out the elemType:
	//
	// list{content: i32*, size: i32, init: i1}
	//
	//
	// In order for the allocation (listPtr := cg.newList(...)) to have the correct type:
	//
	// list{content: i32*, size: i32, init: i1}*
	listType := cg.attrToType(listExpr.TypeHint).(*types.PointerType).ElemType

	listElems := []value.Value{}
	for _, elem := range listExpr.Elements {
		elem.Visit(cg)
		elemVal := cg.lastGenerated

		if isIdentOrIndex(elem) {
			elemVal = cg.LoadVal(elemVal)
		}

		listElems = append(listElems, elemVal)
	}

	listPtr := cg.newList(listElems, listType)
	cg.lastGenerated = listPtr
}
