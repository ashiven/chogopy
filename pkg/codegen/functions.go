package codegen

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) registerFuncs() {
	cg.registerExternal()
	cg.registerBuiltin()
	cg.registerCustom()
}

func (cg *CodeGenerator) registerExternal() {
	strcat := cg.Module.NewFunc(
		"strcat",
		types.I8Ptr,
		ir.NewParam("", types.I8Ptr),
		ir.NewParam("", types.I8Ptr),
	)

	scanf := cg.Module.NewFunc(
		"scanf",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)
	scanf.Sig.Variadic = true

	strcpy := cg.Module.NewFunc(
		"strcpy",
		types.I8Ptr,
		ir.NewParam("", types.I8Ptr),
		ir.NewParam("", types.I8Ptr),
	)

	strcmp := cg.Module.NewFunc(
		"strcmp",
		types.I32,
		ir.NewParam("", types.I8Ptr),
		ir.NewParam("", types.I8Ptr),
	)

	strlen := cg.Module.NewFunc(
		"strlen",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)
	//fgets := cg.Module.NewFunc(
	//	"fgets",
	//	types.I8Ptr,
	//	ir.NewParam("", types.I8Ptr),
	//	ir.NewParam("", types.I32),
	//	ir.NewParam("", types.I8Ptr),
	//)
	//fdopen := cg.Module.NewFunc(
	//	"fdopen",
	//	cg.types["FILE"],
	//	ir.NewParam("", types.I32),
	//	ir.NewParam("", types.I8Ptr),
	//)

	cg.functions["strcat"] = strcat
	cg.functions["scanf"] = scanf
	cg.functions["strcpy"] = strcpy
	cg.functions["strcmp"] = strcmp
	cg.functions["strlen"] = strlen
	// cg.functions["fgets"] = fgets
	// cg.functions["fdopen"] = fdopen
}

func (cg *CodeGenerator) registerBuiltin() {
	// printf is really external but since
	// we are just reusing it for builtin print
	// without modification I'll just leave it here
	print_ := cg.Module.NewFunc(
		"printf",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)
	print_.Sig.Variadic = true

	// TODO: define
	len_ := cg.Module.NewFunc(
		"len",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)

	cg.functions["print"] = print_
	cg.functions["input"] = cg.defineInput()
	cg.functions["len"] = len_
}

func (cg *CodeGenerator) defineInput() *ir.Func {
	input := cg.Module.NewFunc("input", types.I8Ptr)
	funcBlock := input.NewBlock(cg.uniqueNames.get("entry"))

	strFormatConst := constant.NewCharArrayFromString("%s\x00")
	strFormatPtr := funcBlock.NewAlloca(strFormatConst.Type())
	strFormatPtr.LocalName = cg.uniqueNames.get("str_format_ptr")
	funcBlock.NewStore(strFormatConst, strFormatPtr)
	strFormatCast := funcBlock.NewBitCast(strFormatPtr, types.I8Ptr)
	strFormatCast.LocalName = cg.uniqueNames.get("str_format_ptr_cast")

	inputPtr := funcBlock.NewAlloca(types.NewArray(MaxBufferSize, types.I8))
	inputPtr.LocalName = cg.uniqueNames.get("input_ptr")
	inputCast := funcBlock.NewBitCast(inputPtr, types.I8Ptr)
	inputCast.LocalName = cg.uniqueNames.get("input_ptr_cast")

	scanRes := funcBlock.NewCall(cg.functions["scanf"], strFormatCast, inputCast)
	scanRes.LocalName = cg.uniqueNames.get("scan_res")

	funcBlock.NewRet(inputCast)

	return input
}

func (cg *CodeGenerator) registerCustom() {
	cg.functions["booltostr"] = cg.defineBoolToStr()
	cg.functions["floordiv"] = cg.defineFloorDiv()
	cg.functions["newint"] = cg.defineNewInt()
	cg.functions["newbool"] = cg.defineNewBool()
}

// defineBoolPrint converts an i1 to its string representation "True" or "False" */
func (cg *CodeGenerator) defineBoolToStr() *ir.Func {
	arg := ir.NewParam("", types.I1)
	boolToStr := cg.Module.NewFunc("booltostr", types.I8Ptr, arg)

	entry := boolToStr.NewBlock(cg.uniqueNames.get("entry"))
	ifBlock := boolToStr.NewBlock(cg.uniqueNames.get("booltostr.then"))
	elseBlock := boolToStr.NewBlock(cg.uniqueNames.get("booltostr.else"))
	exitBlock := boolToStr.NewBlock(cg.uniqueNames.get("booltostr.exit"))

	resPtr := entry.NewAlloca(types.I8Ptr)
	resPtr.LocalName = cg.uniqueNames.get("booltostr_res_ptr")
	entry.NewCondBr(arg, ifBlock, elseBlock)

	trueConst := constant.NewCharArrayFromString("True\n\x00")
	truePtr := ifBlock.NewAlloca(trueConst.Type())
	truePtr.LocalName = cg.uniqueNames.get("true_ptr")
	ifBlock.NewStore(trueConst, truePtr)
	truePtrCast := ifBlock.NewBitCast(truePtr, types.I8Ptr)
	truePtrCast.LocalName = cg.uniqueNames.get("true_ptr_cast")
	ifBlock.NewStore(truePtrCast, resPtr)
	ifBlock.NewBr(exitBlock)

	falseConst := constant.NewCharArrayFromString("False\n\x00")
	falsePtr := elseBlock.NewAlloca(falseConst.Type())
	falsePtr.LocalName = cg.uniqueNames.get("false_ptr")
	elseBlock.NewStore(falseConst, falsePtr)
	falsePtrCast := elseBlock.NewBitCast(falsePtr, types.I8Ptr)
	falsePtrCast.LocalName = cg.uniqueNames.get("false_ptr_cast")
	elseBlock.NewStore(falsePtrCast, resPtr)
	elseBlock.NewBr(exitBlock)

	resLoad := exitBlock.NewLoad(types.I8Ptr, resPtr)
	resLoad.LocalName = cg.uniqueNames.get("res_load")
	exitBlock.NewRet(resLoad)

	return boolToStr
}

func (cg *CodeGenerator) defineFloorDiv() *ir.Func {
	lhs := ir.NewParam("", types.I32)
	rhs := ir.NewParam("", types.I32)
	floorDiv := cg.Module.NewFunc("floordiv", types.I32, lhs, rhs)
	funcBlock := floorDiv.NewBlock(cg.uniqueNames.get("entry"))

	lhsFloat := funcBlock.NewSIToFP(lhs, types.Float)
	lhsFloat.LocalName = cg.uniqueNames.get("div_lhs_fp")
	rhsFloat := funcBlock.NewSIToFP(rhs, types.Float)
	rhsFloat.LocalName = cg.uniqueNames.get("div_rhs_fp")
	floatDiv := funcBlock.NewFDiv(lhsFloat, rhsFloat)
	floatDiv.LocalName = cg.uniqueNames.get("div_res_fp")

	truncDiv := funcBlock.NewFPToSI(floatDiv, types.I32)
	truncDiv.LocalName = cg.uniqueNames.get("div_res_trunc")
	truncDivFloat := funcBlock.NewSIToFP(truncDiv, types.Float)
	truncDivFloat.LocalName = cg.uniqueNames.get("div_res_trunc_fp")

	// floor(x) = trunc(x) - ((trunc(x) > x) as I32)
	subtractOne := funcBlock.NewFCmp(enum.FPredOGT, truncDivFloat, floatDiv)
	subtractOne.LocalName = cg.uniqueNames.get("trunc_gt_div_res")
	subtractOneInt := funcBlock.NewZExt(subtractOne, types.I32)
	subtractOneInt.LocalName = cg.uniqueNames.get("trunc_gt_div_res_int")
	floorRes := funcBlock.NewSub(truncDiv, subtractOneInt)
	floorRes.LocalName = cg.uniqueNames.get("floor_res")

	funcBlock.NewRet(floorRes)

	return floorDiv
}

func (cg *CodeGenerator) defineNewInt() *ir.Func {
	intLiteral := ir.NewParam("", types.I32)
	newInt := cg.Module.NewFunc("newint", types.I32, intLiteral)
	funcBlock := newInt.NewBlock(cg.uniqueNames.get("entry"))

	intPtr := funcBlock.NewAlloca(types.I32)
	intPtr.LocalName = cg.uniqueNames.get("int_ptr")

	funcBlock.NewStore(intLiteral, intPtr)

	intVal := funcBlock.NewLoad(types.I32, intPtr)
	intVal.LocalName = cg.uniqueNames.get("int_val")

	funcBlock.NewRet(intVal)

	return newInt
}

func (cg *CodeGenerator) defineNewBool() *ir.Func {
	boolLiteral := ir.NewParam("", types.I1)
	newBool := cg.Module.NewFunc("newbool", types.I1, boolLiteral)
	funcBlock := newBool.NewBlock(cg.uniqueNames.get("entry"))

	boolPtr := funcBlock.NewAlloca(types.I1)
	boolPtr.LocalName = cg.uniqueNames.get("bool_ptr")

	funcBlock.NewStore(boolLiteral, boolPtr)

	boolVal := funcBlock.NewLoad(types.I1, boolPtr)
	boolVal.LocalName = cg.uniqueNames.get("bool_val")

	funcBlock.NewRet(boolVal)

	return newBool
}
