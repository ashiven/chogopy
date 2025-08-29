package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitForStmt(forStmt *ast.ForStmt) {
	forCondBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.cond"))
	forBodyBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.body"))
	forIncBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.inc"))
	forExitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.exit"))

	// NOTE: We are using iterName to iterate over a string/list, so we should reset its value to an empty string/0 before assigning to it.
	iterName := cg.variables[forStmt.IterName].value
	iterNameType := cg.variables[forStmt.IterName].elemType

	forStmt.Iter.Visit(cg)
	iterVal := cg.lastGenerated
	iterLength := cg.lengths[iterVal]
	if isIdentOrIndex(forStmt.Iter) {
		iterLength = cg.variables[iterVal.Ident()[1:]].length
		iterVal = cg.LoadVal(iterVal)
	}

	// Some constants for convenience
	zero := constant.NewInt(types.I32, 0)
	one := constant.NewInt(types.I32, 1)
	iterLen := constant.NewInt(types.I32, int64(iterLength))

	// Initialize iteration index
	indexAlloc := cg.currentBlock.NewAlloca(types.I32)
	indexAlloc.LocalName = cg.uniqueNames.get("index_ptr")
	cg.NewStore(zero, indexAlloc)
	cg.currentBlock.NewBr(forCondBlock)

	/* Condition block */
	cg.currentBlock = forCondBlock
	index := cg.currentBlock.NewLoad(types.I32, indexAlloc)
	index.LocalName = cg.uniqueNames.get("index")
	continueLoop := cg.currentBlock.NewICmp(enum.IPredSLT, index, iterLen)
	continueLoop.LocalName = cg.uniqueNames.get("continue")
	cg.currentBlock.NewCondBr(continueLoop, forBodyBlock, forExitBlock)

	/* Body block */
	cg.currentBlock = forBodyBlock
	currentAddress := cg.currentBlock.NewGetElementPtr(iterNameType, iterVal, index)
	currentAddress.LocalName = cg.uniqueNames.get("curr_addr")
	currentVal := cg.currentBlock.NewLoad(iterNameType, currentAddress)
	currentVal.LocalName = cg.uniqueNames.get("curr_val")
	cg.NewStore(currentVal, iterName)
	for _, bodyOp := range forStmt.Body {
		bodyOp.Visit(cg)
	}
	cg.currentBlock.NewBr(forIncBlock)

	/* Increment block */
	cg.currentBlock = forIncBlock
	incremented := cg.currentBlock.NewAdd(index, one)
	incremented.LocalName = cg.uniqueNames.get("inc")
	cg.NewStore(incremented, indexAlloc)
	cg.currentBlock.NewBr(forCondBlock)

	/* Exit block */
	cg.currentBlock = forExitBlock
}
