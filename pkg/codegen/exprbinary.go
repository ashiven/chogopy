package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitBinaryExpr(binaryExpr *ast.BinaryExpr) {
	binaryExpr.Lhs.Visit(cg)
	lhsValue := cg.lastGenerated
	if isIdentOrIndex(binaryExpr.Lhs) {
		lhsValue = cg.LoadVal(lhsValue)
	}

	if cg.shortCircuit(binaryExpr) {
		return
	}

	binaryExpr.Rhs.Visit(cg)
	rhsValue := cg.lastGenerated
	if isIdentOrIndex(binaryExpr.Rhs) {
		rhsValue = cg.LoadVal(rhsValue)
	}

	var resVal value.Value

	switch binaryExpr.Op {
	case "and":
		resVal = cg.currentBlock.NewAnd(lhsValue, rhsValue)
	case "or":
		resVal = cg.currentBlock.NewOr(lhsValue, rhsValue)
	case "%":
		resVal = cg.floorRem(lhsValue, rhsValue)
	case "*":
		resVal = cg.currentBlock.NewMul(lhsValue, rhsValue)
	case "//":
		resVal = cg.floorDiv(lhsValue, rhsValue)
	case "+":
		if cg.concat(binaryExpr, lhsValue, rhsValue) {
			return
		}
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
		if cg.stringEQ(lhsValue, rhsValue) {
			return
		}
		resVal = cg.currentBlock.NewICmp(enum.IPredEQ, lhsValue, rhsValue)
	case "!=":
		if cg.stringNE(lhsValue, rhsValue) {
			return
		}
		resVal = cg.currentBlock.NewICmp(enum.IPredNE, lhsValue, rhsValue)
	case "is":
		// TODO: this should compare the addresses of lhs and rhs but since we are loading their
		// values above (cg.Load(lhs)...) we are actually comparing values here (incorrect)
		resVal = cg.currentBlock.NewICmp(enum.IPredEQ, lhsValue, rhsValue)
	}

	cg.lastGenerated = resVal
}

func (cg *CodeGenerator) floorDiv(lhs value.Value, rhs value.Value) value.Value {
	lhsFloat := cg.currentBlock.NewSIToFP(lhs, types.FP128)
	lhsFloat.LocalName = cg.uniqueNames.get("div_lhs_fp")
	rhsFloat := cg.currentBlock.NewSIToFP(rhs, types.FP128)
	rhsFloat.LocalName = cg.uniqueNames.get("div_rhs_fp")
	floatDiv := cg.currentBlock.NewFDiv(lhsFloat, rhsFloat)
	floatDiv.LocalName = cg.uniqueNames.get("div_res_fp")

	// NOTE: FP128 is a double and FPToSI is apparently undefined if it contains a value that doesn't fit into the result type I32.
	// This may cause problems if someone were to perform calculations with huge numbers but I'll leave it like this for now.
	truncDiv := cg.currentBlock.NewFPToSI(floatDiv, types.I32)
	truncDiv.LocalName = cg.uniqueNames.get("div_res_trunc")
	truncDivFloat := cg.currentBlock.NewSIToFP(truncDiv, types.FP128)
	truncDivFloat.LocalName = cg.uniqueNames.get("div_res_trunc_fp")

	// floor(x) = trunc(x) - ((trunc(x) > x) as I32)
	subtractOne := cg.currentBlock.NewFCmp(enum.FPredOGT, truncDivFloat, floatDiv)
	subtractOne.LocalName = cg.uniqueNames.get("trunc_gt_div_res")
	subtractOneInt := cg.currentBlock.NewZExt(subtractOne, types.I32)
	subtractOneInt.LocalName = cg.uniqueNames.get("trunc_gt_div_res_int")
	floorRes := cg.currentBlock.NewSub(truncDiv, subtractOneInt)
	floorRes.LocalName = cg.uniqueNames.get("floor_res")

	return floorRes

	// floatRem := cg.currentBlock.NewFRem(lhsFloat, rhsFloat)
}

func (cg *CodeGenerator) floorRem(lhs value.Value, rhs value.Value) value.Value {
	/* rem = lhs - rhs * floorDiv(lhs, rhs) */

	floorDiv := cg.floorDiv(lhs, rhs)
	rhsMult := cg.currentBlock.NewMul(rhs, floorDiv)
	floorRem := cg.currentBlock.NewSub(lhs, rhsMult)

	return floorRem
}

//func (cg *CodeGenerator) defineFloorDiv() *ir.Func {
//	lhs := ir.NewParam("", types.I32)
//	rhs := ir.NewParam("", types.I32)
//	floorDiv := cg.Module.NewFunc("floordiv", types.I32, lhs, rhs)
//	funcBlock := floorDiv.NewBlock(cg.uniqueNames.get("entry"))
//
//	lhsFloat := funcBlock.NewSIToFP(lhs, types.FP128)
//	lhsFloat.LocalName = cg.uniqueNames.get("div_lhs_fp")
//	rhsFloat := funcBlock.NewSIToFP(rhs, types.FP128)
//	rhsFloat.LocalName = cg.uniqueNames.get("div_rhs_fp")
//	floatDiv := funcBlock.NewFDiv(lhsFloat, rhsFloat)
//	floatDiv.LocalName = cg.uniqueNames.get("div_res_fp")
//
//	// NOTE: FP128 is a double and FPToSI is apparently undefined if it contains a value that doesn't fit into the result type I32.
//	// This may cause problems if someone were to perform calculations with huge numbers but I'll leave it like this for now.
//	truncDiv := funcBlock.NewFPToSI(floatDiv, types.I32)
//	truncDiv.LocalName = cg.uniqueNames.get("div_res_trunc")
//	truncDivFloat := funcBlock.NewSIToFP(truncDiv, types.FP128)
//	truncDivFloat.LocalName = cg.uniqueNames.get("div_res_trunc_fp")
//
//	// floor(x) = trunc(x) - ((trunc(x) > x) as I32)
//	subtractOne := funcBlock.NewFCmp(enum.FPredOGT, truncDivFloat, floatDiv)
//	subtractOne.LocalName = cg.uniqueNames.get("trunc_gt_div_res")
//	subtractOneInt := funcBlock.NewZExt(subtractOne, types.I32)
//	subtractOneInt.LocalName = cg.uniqueNames.get("trunc_gt_div_res_int")
//	floorRes := funcBlock.NewSub(truncDiv, subtractOneInt)
//	floorRes.LocalName = cg.uniqueNames.get("floor_res")
//
//	funcBlock.NewRet(floorRes)
//
//	return floorDiv
//	// floatRem := cg.currentBlock.NewFRem(lhsFloat, rhsFloat)
//}

func (cg *CodeGenerator) shortCircuit(binaryExpr *ast.BinaryExpr) bool {
	if _, ok := binaryExpr.Lhs.(*ast.LiteralExpr); ok {
		literalVal := binaryExpr.Lhs.(*ast.LiteralExpr).Value
		if binaryExpr.Op == "and" && literalVal == false {
			cg.lastGenerated = cg.NewLiteral(false)
			return true
		}
	}

	if _, ok := binaryExpr.Lhs.(*ast.LiteralExpr); ok {
		literalVal := binaryExpr.Lhs.(*ast.LiteralExpr).Value
		if binaryExpr.Op == "or" && literalVal == true {
			cg.lastGenerated = cg.NewLiteral(true)
			return true
		}
	}

	return false
}

func (cg *CodeGenerator) getLength(val value.Value) int {
	if _, ok := cg.lengths[val]; ok {
		return cg.lengths[val]
	}
	return cg.variables[val.Ident()[1:]].length
}

// TODO: list concat
func (cg *CodeGenerator) concat(binaryExpr *ast.BinaryExpr, lhs value.Value, rhs value.Value) bool {
	if _, ok := binaryExpr.TypeHint.(ast.ListAttribute); ok {
		elemTypeAttr := binaryExpr.TypeHint.(ast.ListAttribute).ElemType
		elemType := cg.attrToType(elemTypeAttr)
		cg.lastGenerated = cg.concatLists(lhs, rhs, elemType)
		return true

	} else if binaryExpr.TypeHint == ast.String {
		cg.lastGenerated = cg.concatStrings(lhs, rhs)
		return true
	}

	return false
}

func (cg *CodeGenerator) concatStrings(lhs value.Value, rhs value.Value) value.Value {
	// 1) Allocate a destination buffer of size: char[BUFFER_SIZE] (needs extra space for stuff to be appended)
	destBuffer := cg.currentBlock.NewAlloca(types.NewArray(MaxBufferSize, types.I8))
	destBuffer.LocalName = cg.uniqueNames.get("concat_buffer_ptr")

	// 2) Copy the string that should be appended to into that buffer
	copyRes := cg.currentBlock.NewCall(cg.functions["strcpy"], destBuffer, lhs)
	copyRes.LocalName = cg.uniqueNames.get("concat_copy_res")

	// 3) Append the other string via strcat
	concatRes := cg.currentBlock.NewCall(cg.functions["strcat"], destBuffer, rhs)
	concatRes.LocalName = cg.uniqueNames.get("concat_append_res")

	return concatRes
}

func (cg *CodeGenerator) concatLists(lhs value.Value, rhs value.Value, elemType types.Type) value.Value {
	concatListPtr := cg.currentBlock.NewAlloca(elemType)
	concatListPtr.LocalName = cg.uniqueNames.get("concat_list_ptr")
	concatListLength := 0

	// TODO: we need a method to get the lengths of lists at runtime
	for i := range cg.getLength(lhs) {
		index := constant.NewInt(types.I64, int64(i))
		elemPtr := cg.currentBlock.NewGetElementPtr(lhs.Type().(*types.PointerType).ElemType, lhs, index)
		elem := cg.currentBlock.NewLoad(lhs.Type().(*types.PointerType).ElemType, elemPtr)
		cg.NewStore(elem, concatListPtr)
		concatListLength++
	}

	for i := range cg.getLength(rhs) {
		index := constant.NewInt(types.I64, int64(i))
		elemPtr := cg.currentBlock.NewGetElementPtr(rhs.Type().(*types.PointerType).ElemType, rhs, index)
		elem := cg.currentBlock.NewLoad(rhs.Type().(*types.PointerType).ElemType, elemPtr)
		cg.NewStore(elem, concatListPtr)
		concatListLength++
	}

	cg.lengths[concatListPtr] = concatListLength
	return concatListPtr
}

func (cg *CodeGenerator) stringEQ(lhs value.Value, rhs value.Value) bool {
	if isString(lhs) && isString(rhs) {
		cmpResInt := cg.currentBlock.NewCall(cg.functions["strcmp"], lhs, rhs)
		cmpRes := cg.currentBlock.NewICmp(enum.IPredEQ, cmpResInt, constant.NewInt(types.I32, 0))
		cg.lastGenerated = cmpRes
		return true
	}
	return false
}

func (cg *CodeGenerator) stringNE(lhs value.Value, rhs value.Value) bool {
	if isString(lhs) && isString(rhs) {
		cmpResInt := cg.currentBlock.NewCall(cg.functions["strcmp"], lhs, rhs)
		cmpRes := cg.currentBlock.NewICmp(enum.IPredEQ, cmpResInt, constant.NewInt(types.I32, 1))
		cg.lastGenerated = cmpRes
		return true
	}
	return false
}
