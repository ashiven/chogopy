package codegen

import "chogopy/pkg/ast"

func (cg *CodeGenerator) VisitIfExpr(ifExpr *ast.IfExpr) {
	ifBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("ifexpr.then"))
	elseBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("ifexpr.else"))
	exitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("ifexpr.exit"))

	resAlloc := cg.currentBlock.NewAlloca(cg.attrToType(ifExpr.TypeHint))
	resAlloc.LocalName = cg.uniqueNames.get("ifexpr_res_ptr")

	ifExpr.Condition.Visit(cg)
	cond := cg.lastGenerated
	cond = cg.LoadVal(cond)
	cg.currentBlock.NewCondBr(cond, ifBlock, elseBlock)

	/* If Block */
	cg.currentBlock = ifBlock
	ifExpr.IfNode.Visit(cg)

	ifBlockRes := cg.lastGenerated
	if isIdentOrIndex(ifExpr.IfNode) {
		ifBlockRes = cg.LoadVal(ifBlockRes)
	}

	cg.NewStore(ifBlockRes, resAlloc)
	cg.currentBlock.NewBr(exitBlock)

	/* Else Block */
	cg.currentBlock = elseBlock
	ifExpr.ElseNode.Visit(cg)

	elseBlockRes := cg.lastGenerated
	if isIdentOrIndex(ifExpr.ElseNode) {
		elseBlockRes = cg.LoadVal(elseBlockRes)
	}

	cg.NewStore(elseBlockRes, resAlloc)
	cg.currentBlock.NewBr(exitBlock)

	/* Exit Block */
	cg.currentBlock = exitBlock
	cg.lastGenerated = resAlloc
}
