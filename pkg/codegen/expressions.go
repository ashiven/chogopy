package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitLiteralExpr(literalExpr *ast.LiteralExpr) {
	literal := cg.NewLiteral(literalExpr.Value)
	cg.lastGenerated = literal
}

func (cg *CodeGenerator) VisitIdentExpr(identExpr *ast.IdentExpr) {
	identName := identExpr.Identifier

	// Case 1: The identifier refers to a global var def.
	if variable, ok := cg.variables[identName]; ok {
		cg.lastGenerated = variable.value
	}

	// Case 2: The identifier refers to the name of a parameter of the current function. (overwrites global def)
	for _, param := range cg.currentFunction.Params {
		if identName == param.LocalName {
			cg.lastGenerated = param
		}
	}
}

func (cg *CodeGenerator) VisitUnaryExpr(unaryExpr *ast.UnaryExpr) {
	unaryExpr.Value.Visit(cg)
	unaryVal := cg.lastGenerated

	var resVal value.Value

	switch unaryExpr.Op {
	case "not":
		true_ := cg.NewLiteral(true)
		resVal = cg.currentBlock.NewXor(true_, unaryVal)

	case "-":
		zero := cg.NewLiteral(0)
		resVal = cg.currentBlock.NewSub(zero, unaryVal)
	}

	cg.lastGenerated = resVal
}

func (cg *CodeGenerator) VisitBinaryExpr(binaryExpr *ast.BinaryExpr) {
	binaryExpr.Lhs.Visit(cg)
	lhsValue := cg.lastGenerated

	// Short circuit AND expr if lhs is literal False
	if _, ok := binaryExpr.Lhs.(*ast.LiteralExpr); ok {
		literalVal := binaryExpr.Lhs.(*ast.LiteralExpr).Value
		if binaryExpr.Op == "and" && literalVal == false {
			cg.lastGenerated = cg.NewLiteral(false)
			return
		}
	}

	// Short circuit OR expr if lhs is literal True
	if _, ok := binaryExpr.Lhs.(*ast.LiteralExpr); ok {
		literalVal := binaryExpr.Lhs.(*ast.LiteralExpr).Value
		if binaryExpr.Op == "or" && literalVal == true {
			cg.lastGenerated = cg.NewLiteral(true)
			return
		}
	}

	binaryExpr.Rhs.Visit(cg)
	rhsValue := cg.lastGenerated

	var resVal value.Value

	switch binaryExpr.Op {
	case "and":
		resVal = cg.currentBlock.NewAnd(lhsValue, rhsValue)
	case "or":
		resVal = cg.currentBlock.NewOr(lhsValue, rhsValue)
	case "%":
		// TODO: this is broken for negative values for the same
		// reason that div below is broken
		resVal = cg.currentBlock.NewSRem(lhsValue, rhsValue)
	case "*":
		resVal = cg.currentBlock.NewMul(lhsValue, rhsValue)
	case "//":
		// TODO: implement floor div for negative values:
		// (if the div result is negative, it will just be rounded
		// to the next whole number in a positive direction while
		// we want this direction to still remain negative)
		resVal = cg.currentBlock.NewSDiv(lhsValue, rhsValue)
	case "+":
		// TODO: string/list concat and updating lengths in cg.lengths and variables.length
		resVal = cg.currentBlock.NewAdd(lhsValue, rhsValue)
	case "-":
		resVal = cg.currentBlock.NewSub(lhsValue, rhsValue)
	case "<":
		resVal = cg.currentBlock.NewICmp(enum.IPredSLT, lhsValue, rhsValue)
	case "<=":
		resVal = cg.currentBlock.NewICmp(enum.IPredSLE, lhsValue, rhsValue)
	case ">":
		resVal = cg.currentBlock.NewICmp(enum.IPredSGT, lhsValue, rhsValue)
	case ">=":
		resVal = cg.currentBlock.NewICmp(enum.IPredSGE, lhsValue, rhsValue)
	case "==":
		resVal = cg.currentBlock.NewICmp(enum.IPredEQ, lhsValue, rhsValue)
	case "!=":
		resVal = cg.currentBlock.NewICmp(enum.IPredNE, lhsValue, rhsValue)
	case "is":
		resVal = cg.currentBlock.NewICmp(enum.IPredEQ, lhsValue, rhsValue)
	}

	cg.lastGenerated = resVal
}

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

	cg.currentBlock = ifBlock
	ifExpr.IfNode.Visit(cg)
	ifBlockRes := cg.lastGenerated
	cg.NewStore(ifBlockRes, resAlloc)
	cg.currentBlock.NewBr(exitBlock)

	cg.currentBlock = elseBlock
	ifExpr.ElseNode.Visit(cg)
	elseBlockRes := cg.lastGenerated
	cg.NewStore(elseBlockRes, resAlloc)
	cg.currentBlock.NewBr(exitBlock)

	cg.currentBlock = exitBlock
	cg.lastGenerated = resAlloc
}

func (cg *CodeGenerator) VisitListExpr(listExpr *ast.ListExpr) {
	listElemType := cg.attrToType(listExpr.TypeHint.(ast.ListAttribute).ElemType)
	listPtr := cg.currentBlock.NewAlloca(listElemType)
	listPtr.LocalName = cg.uniqueNames.get("list_ptr")

	for elemIdx, elem := range listExpr.Elements {
		elem.Visit(cg)
		elemVal := cg.lastGenerated

		elemIdxConst := constant.NewInt(types.I32, int64(elemIdx))
		elemPtr := cg.currentBlock.NewGetElementPtr(listElemType, listPtr, elemIdxConst)
		elemPtr.LocalName = cg.uniqueNames.get("list_elem_ptr")

		cg.NewStore(elemVal, elemPtr)
	}

	cg.lengths[listPtr] = len(listExpr.Elements)
	cg.lastGenerated = listPtr
}

func (cg *CodeGenerator) VisitCallExpr(callExpr *ast.CallExpr) {
	args := []value.Value{}
	for _, arg := range callExpr.Arguments {
		arg.Visit(cg)
		args = append(args, cg.lastGenerated)
	}

	switch callExpr.FuncName {
	case "print":
		args = cg.convertPrintArgs(args)
	}

	callee := cg.functions[callExpr.FuncName]
	callRes := cg.currentBlock.NewCall(callee, args...)
	callRes.LocalName = cg.uniqueNames.get("call_res")

	cg.lastGenerated = callRes
}

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
