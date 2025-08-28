package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
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
	if varDef, ok := cg.varDefs[identName]; ok {
		cg.lastGenerated = varDef.value
	}

	// Case 2: The identifier refers to the name of a parameter of the current function. (overwrites global def)
	for _, param := range cg.currentFunction.Params {
		if identName == param.LocalName {
			cg.lastGenerated = param
		}
	}
}

func (cg *CodeGenerator) VisitUnaryExpr(unaryExpr *ast.UnaryExpr) {
	// TODO: implement
}

func (cg *CodeGenerator) VisitBinaryExpr(binaryExpr *ast.BinaryExpr) {
	binaryExpr.Lhs.Visit(cg)
	lhsValue := cg.lastGenerated

	binaryExpr.Rhs.Visit(cg)
	rhsValue := cg.lastGenerated

	// TODO: implement
	switch binaryExpr.Op {
	case "+":
		newAdd := cg.currentBlock.NewAdd(lhsValue, rhsValue)
		cg.lastGenerated = newAdd
	}
}

func (cg *CodeGenerator) VisitIfExpr(ifExpr *ast.IfExpr) {
	// TODO: implement
}

func (cg *CodeGenerator) VisitListExpr(listExpr *ast.ListExpr) {
	listElemType := attrToType(listExpr.TypeHint.(ast.ListAttribute).ElemType)
	listAlloc := cg.currentBlock.NewAlloca(listElemType)
	listAlloc.LocalName = cg.uniqueNames.get("list_ptr")

	for elemIdx, elem := range listExpr.Elements {
		elem.Visit(cg)
		elemVal := cg.lastGenerated

		elemIdxConst := constant.NewInt(types.I32, int64(elemIdx))
		elemPtr := cg.currentBlock.NewGetElementPtr(listElemType, listAlloc, elemIdxConst)
		elemPtr.LocalName = cg.uniqueNames.get("list_elem_ptr")

		cg.currentBlock.NewStore(elemVal, elemPtr)
	}

	cg.lastGenerated = listAlloc
}

func (cg *CodeGenerator) VisitCallExpr(callExpr *ast.CallExpr) {
	args := []value.Value{}
	for _, arg := range callExpr.Arguments {
		arg.Visit(cg)
		args = append(args, cg.lastGenerated)
	}

	switch callExpr.FuncName {
	case "print":
		printArgs := []value.Value{}
		for _, arg := range args {
			if hasType(arg, types.I32) || isPtrTo(arg, types.I32) {
				/* Integer print */
				digitStr := cg.NewLiteral("%d")
				argVal := cg.LoadVal(arg)
				printArgs = append(printArgs, digitStr)
				printArgs = append(printArgs, argVal)
			} else if hasType(arg, types.I1) || isPtrTo(arg, types.I1) {
				/* Boolean print */
				// TODO: use something like a ternary expr to print "True" if val is 1 else "False"
			} else {
				/* String print */
				bitCast := cg.currentBlock.NewBitCast(arg, types.I8Ptr)
				bitCast.LocalName = cg.uniqueNames.get("print_arg_cast")
				printArgs = append(printArgs, bitCast)
			}
		}
		args = printArgs
	}

	callee := cg.funcDefs[callExpr.FuncName]
	newCall := cg.currentBlock.NewCall(callee, args...)
	newCall.LocalName = cg.uniqueNames.get("call")
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
