package codegen

import "chogopy/pkg/ast"

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
	cg.currentBlock.NewBr(whileCondBlock)

	for _, bodyOp := range whileStmt.Body {
		bodyOp.Visit(cg)
	}

	/* Exit block */
	cg.currentBlock = whileExitBlock
}
