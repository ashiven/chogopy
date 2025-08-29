package codegen

import "chogopy/pkg/ast"

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
