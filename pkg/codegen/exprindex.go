package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitIndexExpr(indexExpr *ast.IndexExpr) {
	// No need for isIdentOrIndex check before LoadVal because that
	// will always be true for indexExpr after type checking.
	indexExpr.Value.Visit(cg)
	val := cg.LoadVal(cg.lastGenerated)

	indexExpr.Index.Visit(cg)
	index := cg.lastGenerated

	currentAddr := cg.currentBlock.NewGetElementPtr(val.Type().(*types.PointerType).ElemType, val, index)
	currentAddr.LocalName = cg.uniqueNames.get("index_addr")

	cg.lastGenerated = currentAddr
}
