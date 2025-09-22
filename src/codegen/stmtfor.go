package codegen

import (
	"chogopy/src/ast"
	"log"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitForStmt(forStmt *ast.ForStmt) {
	forCondBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.cond"))
	forBodyBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.body"))
	forIncBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.inc"))
	forExitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.exit"))

	iterNameInfo, err := cg.getVar(forStmt.IterName)
	if err != nil {
		log.Fatalln(err.Error())
	}

	iterName := iterNameInfo.value

	forStmt.Iter.Visit(cg)
	iterVal := cg.lastGenerated
	if isIdentOrIndex(forStmt.Iter) {
		iterVal = cg.LoadVal(iterVal)
	}

	iterLen := cg.getLen(iterVal)

	zero := constant.NewInt(types.I32, 0)
	one := constant.NewInt(types.I32, 1)

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
	cg.currentBlock.NewBr(forIncBlock)

	var currentVal value.Value
	if isList(iterVal) {
		currentVal = cg.getListElem(iterVal, index)
	} else {
		currentVal = cg.getStringElem(iterVal, index)
	}

	cg.NewStore(currentVal, iterName)

	for _, bodyOp := range forStmt.Body {
		bodyOp.Visit(cg)
	}

	/* Increment block */
	cg.currentBlock = forIncBlock
	cg.currentBlock.NewBr(forCondBlock)

	incremented := cg.currentBlock.NewAdd(index, one)
	incremented.LocalName = cg.uniqueNames.get("inc")
	cg.NewStore(incremented, indexAlloc)

	/* Exit block */
	cg.currentBlock = forExitBlock
}
