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

	exit := cg.Module.NewFunc(
		"exit",
		types.Void,
		ir.NewParam("", types.I32),
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
	cg.functions["exit"] = exit
	// cg.functions["fgets"] = fgets
	// cg.functions["fdopen"] = fdopen
}

func (cg *CodeGenerator) registerBuiltin() {
	// printf is really external but since
	// we are just reusing it for builtin print
	// without modification I'll just leave it here
	printf_ := cg.Module.NewFunc(
		"printf",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)
	printf_.Sig.Variadic = true

	// this is basically just a dummy function that doesn't
	// do anything because if there is an actual call to len()
	// it will be redirected to strlen or listlen based on argument type
	len_ := cg.Module.NewFunc(
		"len",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)

	cg.functions["printf"] = printf_
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
	cg.functions["printstr"] = cg.definePrintString()
	cg.functions["printint"] = cg.definePrintInt()
	cg.functions["booltostr"] = cg.defineBoolToStr()
	cg.functions["printbool"] = cg.definePrintBool()
	cg.functions["floordiv"] = cg.defineFloorDiv()
	cg.functions["newint"] = cg.defineNewInt()
	cg.functions["newbool"] = cg.defineNewBool()
	cg.functions["listinit"] = cg.defineListInit()
	cg.functions["listlen"] = cg.defineListLen()
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

func (cg *CodeGenerator) defineListInit() *ir.Func {
	list := ir.NewParam("", types.NewPointer(cg.types["list"]))
	listInitFunc := cg.Module.NewFunc("listinit", types.I1, list)
	funcBlock := listInitFunc.NewBlock(cg.uniqueNames.get("entry"))

	zero := constant.NewInt(types.I32, 0)
	initFieldIdx := constant.NewInt(types.I32, 2)
	listInitAddr := funcBlock.NewGetElementPtr(
		list.Type().(*types.PointerType).ElemType,
		list,
		zero,
		initFieldIdx,
	)
	listInitAddr.LocalName = cg.uniqueNames.get("list_init_addr")

	listInit := funcBlock.NewLoad(types.I1, listInitAddr)
	listInit.LocalName = cg.uniqueNames.get("list_init")

	funcBlock.NewRet(listInit)

	return listInitFunc
}

func (cg *CodeGenerator) defineListLen() *ir.Func {
	list := ir.NewParam("", types.NewPointer(cg.types["list"]))
	listLenFunc := cg.Module.NewFunc("listlen", types.I32, list)
	funcBlock := listLenFunc.NewBlock(cg.uniqueNames.get("entry"))

	/* Error if list is uninitialized */
	initTrueBlock := listLenFunc.NewBlock("init.true")
	initFalseBlock := listLenFunc.NewBlock("init.false")

	listInit := funcBlock.NewCall(cg.functions["listinit"], list)
	funcBlock.NewCondBr(listInit, initTrueBlock, initFalseBlock)

	/* Get list length */
	zero := constant.NewInt(types.I32, 0)
	lenFieldIdx := constant.NewInt(types.I32, 1)
	listLenAddr := initTrueBlock.NewGetElementPtr(
		list.Type().(*types.PointerType).ElemType,
		list,
		zero,
		lenFieldIdx,
	)
	listLenAddr.LocalName = cg.uniqueNames.get("list_len_addr")

	listLen := initTrueBlock.NewLoad(types.I32, listLenAddr)
	listLen.LocalName = cg.uniqueNames.get("list_len")

	initTrueBlock.NewRet(listLen)

	/* Raise runtime exception len called on uninitialized list */
	errorConst := constant.NewCharArrayFromString("TypeError: object of type 'NoneType' has no len()\n\x00")
	errorPtr := initFalseBlock.NewAlloca(errorConst.Type())
	errorPtr.LocalName = cg.uniqueNames.get("error_ptr")
	initFalseBlock.NewStore(errorConst, errorPtr)
	errorCast := initFalseBlock.NewBitCast(errorPtr, types.I8Ptr)
	errorCast.LocalName = cg.uniqueNames.get("error_cast")

	initFalseBlock.NewCall(cg.functions["printf"], errorCast)
	initFalseBlock.NewCall(cg.functions["exit"], zero)
	initFalseBlock.NewRet(zero)

	return listLenFunc
}

func (cg *CodeGenerator) definePrintString() *ir.Func {
	strArg := ir.NewParam("", types.I8Ptr)
	printString := cg.Module.NewFunc("printstr", types.I32, strArg)
	funcBlock := printString.NewBlock(cg.uniqueNames.get("entry"))

	formatConst := constant.NewCharArrayFromString("%s\n\x00")
	formatPtr := funcBlock.NewAlloca(formatConst.Type())
	formatPtr.LocalName = cg.uniqueNames.get("print_fmt_ptr")
	funcBlock.NewStore(formatConst, formatPtr)
	formatCast := funcBlock.NewBitCast(formatPtr, types.I8Ptr)
	formatCast.LocalName = cg.uniqueNames.get("print_fmt_cast")

	printRes := funcBlock.NewCall(cg.functions["printf"], formatCast, strArg)

	funcBlock.NewRet(printRes)

	return printString
}

func (cg *CodeGenerator) definePrintInt() *ir.Func {
	intArg := ir.NewParam("", types.I32)
	printInt := cg.Module.NewFunc("printint", types.I32, intArg)
	funcBlock := printInt.NewBlock(cg.uniqueNames.get("entry"))

	formatConst := constant.NewCharArrayFromString("%d\n\x00")
	formatPtr := funcBlock.NewAlloca(formatConst.Type())
	formatPtr.LocalName = cg.uniqueNames.get("print_fmt_ptr")
	funcBlock.NewStore(formatConst, formatPtr)
	formatCast := funcBlock.NewBitCast(formatPtr, types.I8Ptr)
	formatCast.LocalName = cg.uniqueNames.get("print_fmt_cast")

	printRes := funcBlock.NewCall(cg.functions["printf"], formatCast, intArg)

	funcBlock.NewRet(printRes)

	return printInt
}

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

func (cg *CodeGenerator) definePrintBool() *ir.Func {
	boolArg := ir.NewParam("", types.I1)
	printBool := cg.Module.NewFunc("printbool", types.I32, boolArg)
	funcBlock := printBool.NewBlock(cg.uniqueNames.get("entry"))

	boolStr := funcBlock.NewCall(cg.functions["booltostr"], boolArg)
	printRes := funcBlock.NewCall(cg.functions["printf"], boolStr)

	funcBlock.NewRet(printRes)

	return printBool
}
