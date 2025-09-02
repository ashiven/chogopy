package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitIndexExpr(indexExpr *ast.IndexExpr) {
	indexExpr.Value.Visit(cg)
	val := cg.lastGenerated

	if isIdentOrIndex(indexExpr.Value) {
		val = cg.LoadVal(cg.lastGenerated)
	}

	indexExpr.Index.Visit(cg)
	index := cg.lastGenerated

	if isIdentOrIndex(indexExpr.Index) {
		index = cg.LoadVal(index)
	}

	var currentAddr value.Value
	if isList(val) {
		currentAddr = cg.getListElemPtr(val, index)
	} else {
		currentAddr = cg.currentBlock.NewGetElementPtr(val.Type().(*types.PointerType).ElemType, val, index)
		currentAddr.(*ir.InstGetElementPtr).LocalName = cg.uniqueNames.get("index_addr")
	}

	// NOTE: The below does not work if our language allows strings assignments like "test"[2] = "c".
	// Something like "test"[1] should not return the whole remaining string "est"
	// but rather be clamped to size 1 so the return will be "e" instead.
	if isString(val) {
		cg.lastGenerated = cg.clampString(currentAddr)
		return
	}

	// NOTE:
	// An index expression can appear both on the left hand side and on the right hand side of
	// an assign statement. Therefore, it's result can both be used to store and to load a value.
	// For this reason, the return of this NEEDS TO BE AN ADDRESS!
	// We don't do anything related to loading or storing in here and leave that decision to the
	// caller who has to decide based on context whether to use the resulting address for storing
	// a value at it or loading a value from it.
	cg.lastGenerated = currentAddr
}
