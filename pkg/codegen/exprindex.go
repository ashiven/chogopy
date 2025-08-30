package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitIndexExpr(indexExpr *ast.IndexExpr) {
	indexExpr.Value.Visit(cg)
	val := cg.lastGenerated

	if isIdentOrIndex(indexExpr.Value) {
		val = cg.LoadVal(cg.lastGenerated)
	}

	indexExpr.Index.Visit(cg)
	index := cg.lastGenerated

	currentAddr := cg.currentBlock.NewGetElementPtr(val.Type().(*types.PointerType).ElemType, val, index)
	currentAddr.LocalName = cg.uniqueNames.get("index_addr")

	// Something like "test"[1] should not return the whole remaining string "est"
	// but rather be clamped to size 1 so the return will be "e" instead.
	if isString(val) {
		cg.lastGenerated = cg.clampStrSize(currentAddr)
		return
	}

	cg.lastGenerated = currentAddr
}
