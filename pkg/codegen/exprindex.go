package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitIndexExpr(indexExpr *ast.IndexExpr) {
	indexExpr.Value.Visit(cg)
	// We have to do this to get the value out of the global variable pointer
	// returned after visiting identExpr but it will break stuff if this is done again inside a nested indexExpr
	val := cg.LoadVal(cg.lastGenerated)

	indexExpr.Index.Visit(cg)
	index := cg.lastGenerated

	currentAddr := cg.currentBlock.NewGetElementPtr(val.Type().(*types.PointerType).ElemType, val, index)
	currentAddr.LocalName = cg.uniqueNames.get("index_addr")

	cg.lastGenerated = cg.LoadVal(currentAddr)

	// TODO: this works for simple index exprs but will go horribly
	// wrong for anything more complicated like nested index exprs
	// or indexing into a string literal etc.
}
