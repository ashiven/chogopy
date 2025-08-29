package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitIfStmt(ifStmt *ast.IfStmt) {
	ifBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("if.then"))
	elseBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("if.else"))
	exitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("if.exit"))

	ifStmt.Condition.Visit(cg)
	cond := cg.lastGenerated
	cond = cg.LoadVal(cond)
	cg.currentBlock.NewCondBr(cond, ifBlock, elseBlock)

	cg.currentBlock = ifBlock
	for _, ifBodyNode := range ifStmt.IfBody {
		ifBodyNode.Visit(cg)
	}
	cg.currentBlock.NewBr(exitBlock)

	cg.currentBlock = elseBlock
	for _, elseBodyNode := range ifStmt.ElseBody {
		elseBodyNode.Visit(cg)
	}
	cg.currentBlock.NewBr(exitBlock)

	cg.currentBlock = exitBlock
}

func (cg *CodeGenerator) VisitWhileStmt(whileStmt *ast.WhileStmt) {
	whileCondBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("while.cond"))
	whileBodyBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("while.body"))
	whileExitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("while.exit"))

	cg.currentBlock.NewBr(whileCondBlock)

	/* Condition block */
	cg.currentBlock = whileCondBlock
	whileStmt.Condition.Visit(cg)
	cond := cg.lastGenerated
	if isIdentOrIndex(whileStmt.Condition) {
		cond = cg.LoadVal(cond)
	}
	cg.currentBlock.NewCondBr(cond, whileBodyBlock, whileExitBlock)

	/* Body block */
	cg.currentBlock = whileBodyBlock
	for _, bodyOp := range whileStmt.Body {
		bodyOp.Visit(cg)
	}
	cg.currentBlock.NewBr(whileCondBlock)

	/* Exit block */
	cg.currentBlock = whileExitBlock
}

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

func (cg *CodeGenerator) VisitPassStmt(passStmt *ast.PassStmt) {
	/* no op */
}

func (cg *CodeGenerator) VisitReturnStmt(returnStmt *ast.ReturnStmt) {
	var returnVal value.Value
	if returnStmt.ReturnVal != nil {
		returnStmt.ReturnVal.Visit(cg)
		returnVal = cg.lastGenerated
	} else {
		returnVal = nil
	}

	cg.currentBlock.NewRet(returnVal)
}

func (cg *CodeGenerator) VisitAssignStmt(assignStmt *ast.AssignStmt) {
	assignStmt.Target.Visit(cg)
	target := cg.lastGenerated

	assignStmt.Value.Visit(cg)
	value := cg.lastGenerated

	if isIdentOrIndex(assignStmt.Value) {
		value = cg.LoadVal(value)
	}

	cg.NewStore(value, target)
}
