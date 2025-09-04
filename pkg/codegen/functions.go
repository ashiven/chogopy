package codegen

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) registerFuncs() {
	cg.addStringConstants()
	cg.registerExternal()
	cg.registerBuiltin()
	cg.registerCustom()
}

func (cg *CodeGenerator) addStringConstants() {
	cg.strings["str_format"] = cg.globalStringDef("str_format", "%s\x00")
	cg.strings["str_format_newline"] = cg.globalStringDef("str_format_newline", "%s\n\x00")
	cg.strings["digit_format"] = cg.globalStringDef("digit_format", "%d\x00")
	cg.strings["digit_format_newline"] = cg.globalStringDef("digit_format_newline", "%d\n\x00")

	cg.strings["true_newline"] = cg.globalStringDef("true_newline", "True\n\x00")
	cg.strings["false_newline"] = cg.globalStringDef("false_newline", "False\n\x00")

	cg.strings["error_len_none"] = cg.globalStringDef("error_len_none", "TypeError: object of type 'NoneType' has no len()\n\x00")
	cg.strings["error_index_none"] = cg.globalStringDef("error_index_none", "TypeError: 'NoneType' object is not subscriptable\n\x00")
	cg.strings["error_index_neg"] = cg.globalStringDef("error_index_neg", "IndexError: list index out of range\n\x00")
	cg.strings["error_index_oob"] = cg.globalStringDef("error_index_oob", "IndexError: list index out of range\n\x00")
}

func (cg *CodeGenerator) globalStringDef(defName string, strLiteral string) *ir.Global {
	strConst := constant.NewCharArrayFromString(strLiteral)
	return cg.Module.NewGlobalDef(defName, strConst)
}

func (cg *CodeGenerator) useStringDef(block *ir.Block, defName string) *ir.InstLoad {
	zero := constant.NewInt(types.I32, 0)
	strConst := constant.NewGetElementPtr(cg.strings[defName].Typ.ElemType, cg.strings[defName], zero, zero)

	strPtr := block.NewAlloca(types.I8Ptr)
	strPtr.LocalName = cg.uniqueNames.get("str_ptr")
	block.NewStore(strConst, strPtr)

	globalStr := block.NewLoad(types.I8Ptr, strPtr)
	globalStr.LocalName = cg.uniqueNames.get("global_str")

	return globalStr
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

	memcpy := cg.Module.NewFunc(
		"memcpy",
		types.I8Ptr,
		ir.NewParam("", types.I32Ptr),
		ir.NewParam("", types.I32Ptr),
		ir.NewParam("", types.I32),
	)

	sprintf := cg.Module.NewFunc(
		"sprintf",
		types.I32,
		ir.NewParam("", types.I8Ptr),
		ir.NewParam("", types.I8Ptr),
	)
	sprintf.Sig.Variadic = true

	printf := cg.Module.NewFunc(
		"printf",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)
	printf.Sig.Variadic = true

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
	cg.functions["memcpy"] = memcpy
	cg.functions["sprintf"] = sprintf
	cg.functions["printf"] = printf
	// cg.functions["fgets"] = fgets
	// cg.functions["fdopen"] = fdopen
}

func (cg *CodeGenerator) registerBuiltin() {
	print_ := cg.Module.NewFunc(
		"print",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)

	len_ := cg.Module.NewFunc(
		"len",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)

	cg.functions["print"] = print_
	cg.functions["input"] = cg.defineInput()
	cg.functions["len"] = len_
}

// TODO: The memory this function allocates for the input string is allocated on the functions
// local stack frame, which will be freed after the function returns.
// Therefore, we are returning a pointer to unallocated memory.
// We should instead create a global string constant or use malloc to allocate memory on the heap.
// The same applies to any non-predefined functions defined by the user. If they somehow return a string
// from their function, the way we are currently generating the code, that string will be a pointer to
// unallocated memory. (allocated on the functions stack frame)
func (cg *CodeGenerator) defineInput() *ir.Func {
	input := cg.Module.NewFunc("input", types.I8Ptr)
	funcBlock := input.NewBlock(cg.uniqueNames.get("entry"))

	strFormatPtr := cg.useStringDef(funcBlock, "str_format")

	inputPtr := funcBlock.NewAlloca(types.NewArray(MaxBufferSize, types.I8))
	inputPtr.LocalName = cg.uniqueNames.get("input_ptr")
	inputCast := funcBlock.NewBitCast(inputPtr, types.I8Ptr)
	inputCast.LocalName = cg.uniqueNames.get("input_ptr_cast")

	scanRes := funcBlock.NewCall(cg.functions["scanf"], strFormatPtr, inputCast)
	scanRes.LocalName = cg.uniqueNames.get("scan_res")

	funcBlock.NewRet(inputCast)

	return input
}

func (cg *CodeGenerator) registerCustom() {
	cg.functions["printstr"] = cg.definePrintString()
	cg.functions["printint"] = cg.definePrintInt()
	cg.functions["printbool"] = cg.definePrintBool()
	cg.functions["floordiv"] = cg.defineFloorDiv()
	cg.functions["newint"] = cg.defineNewInt()
	cg.functions["newbool"] = cg.defineNewBool()

	/* function defs for each list type */
	for _, listType := range cg.types {
		if isListType(listType) {
			listElemType := getListElemTypeFromListType(listType)

			initName := listType.Name() + "_init"
			cg.functions[initName] = cg.defineListInit(initName, listType)

			lenName := listType.Name() + "_len"
			cg.functions[lenName] = cg.defineListLen(lenName, listType)

			elemPtrName := listType.Name() + "_elemptr"
			cg.functions[elemPtrName] = cg.defineListElemPtr(elemPtrName, listType, listElemType)
		}
	}
}

func (cg *CodeGenerator) exceptionHelper(block *ir.Block, exception string) {
	returnCode := constant.NewInt(types.I32, 0)
	errorStr := cg.useStringDef(block, exception)

	block.NewCall(cg.functions["printf"], errorStr)
	block.NewCall(cg.functions["exit"], returnCode)
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

func (cg *CodeGenerator) defineListInit(funcName string, listType types.Type) *ir.Func {
	list := ir.NewParam("", types.NewPointer(listType))
	listInitFunc := cg.Module.NewFunc(funcName, types.I1, list)
	funcBlock := listInitFunc.NewBlock(cg.uniqueNames.get("entry"))

	zero := constant.NewInt(types.I32, 0)
	initFieldIdx := constant.NewInt(types.I32, 2)
	listInitAddr := funcBlock.NewGetElementPtr(listType, list, zero, initFieldIdx)
	listInitAddr.LocalName = cg.uniqueNames.get("list_init_addr")

	listInit := funcBlock.NewLoad(types.I1, listInitAddr)
	listInit.LocalName = cg.uniqueNames.get("list_init")

	funcBlock.NewRet(listInit)

	return listInitFunc
}

func (cg *CodeGenerator) defineListLen(funcName string, listType types.Type) *ir.Func {
	list := ir.NewParam("", types.NewPointer(listType))
	listLenFunc := cg.Module.NewFunc(funcName, types.I32, list)
	funcBlock := listLenFunc.NewBlock(cg.uniqueNames.get("entry"))

	/* Error if list is uninitialized */
	initTrueBlock := listLenFunc.NewBlock("init.true")
	initFalseBlock := listLenFunc.NewBlock("init.false")

	initFuncName := listType.Name() + "_init"
	listInit := funcBlock.NewCall(cg.functions[initFuncName], list)
	funcBlock.NewCondBr(listInit, initTrueBlock, initFalseBlock)

	/* Get list length */
	zero := constant.NewInt(types.I32, 0)
	lenFieldIdx := constant.NewInt(types.I32, 1)
	listLenAddr := initTrueBlock.NewGetElementPtr(listType, list, zero, lenFieldIdx)
	listLenAddr.LocalName = cg.uniqueNames.get("list_len_addr")

	listLen := initTrueBlock.NewLoad(types.I32, listLenAddr)
	listLen.LocalName = cg.uniqueNames.get("list_len")
	initTrueBlock.NewRet(listLen)

	/* Raise runtime exception len called on uninitialized list */
	cg.exceptionHelper(initFalseBlock, "error_len_none")
	initFalseBlock.NewRet(zero)

	return listLenFunc
}

func (cg *CodeGenerator) defineListElemPtr(funcName string, listType types.Type, listElemType types.Type) *ir.Func {
	list := ir.NewParam("", types.NewPointer(listType))
	index := ir.NewParam("", types.I32)
	listLenFunc := cg.Module.NewFunc(funcName, types.NewPointer(listElemType), list, index)
	funcBlock := listLenFunc.NewBlock(cg.uniqueNames.get("entry"))

	/* Error if list is uninitialized, index is negative, or index is out of bounds */
	initTrueBlock := listLenFunc.NewBlock("init.true")
	initFalseBlock := listLenFunc.NewBlock("init.false")
	indexPosBlock := listLenFunc.NewBlock("index.positive")
	indexNegBlock := listLenFunc.NewBlock("index.negative")
	indexIBBlock := listLenFunc.NewBlock("index.inbounds")
	indexOOBBlock := listLenFunc.NewBlock("index.outofbounds")

	/* Check that list is not None */
	initFuncName := listType.Name() + "_init"
	listInit := funcBlock.NewCall(cg.functions[initFuncName], list)
	funcBlock.NewCondBr(listInit, initTrueBlock, initFalseBlock)

	/* Check that index is greater equal zero */
	zero := constant.NewInt(types.I32, 0)
	indexPositive := initTrueBlock.NewICmp(enum.IPredSGE, index, zero)
	initTrueBlock.NewCondBr(indexPositive, indexPosBlock, indexNegBlock)

	/* Check that index is not greater than list len */
	lenFuncName := listType.Name() + "_len"
	listLen := indexPosBlock.NewCall(cg.functions[lenFuncName], list)
	indexInBounds := indexPosBlock.NewICmp(enum.IPredSLT, index, listLen)
	indexPosBlock.NewCondBr(indexInBounds, indexIBBlock, indexOOBBlock)

	/* Get list element pointer at index */
	listContentAddr := indexIBBlock.NewGetElementPtr(listType, list, zero, zero)
	listContentAddr.LocalName = cg.uniqueNames.get("list_content_addr")

	listContentType := types.NewPointer(listElemType)
	listContentPtr := indexIBBlock.NewLoad(listContentType, listContentAddr)
	listContentPtr.LocalName = cg.uniqueNames.get("list_content_ptr")

	// In case the content of the list is a pointer to another list (lists have a struct type)
	// the first GEP index will select the list struct and the second index will select the field to store into (list.content)
	// Think about it like this:
	//
	// - You have a pointer to a struct list: list*
	// - This points to a contiguous location in memory at which one or more list structs reside
	//
	// Memory(starting at list*):
	//
	// 							0:	list{content: i32*, size: i32}
	// elemIdx ->		1:	list{content: i32*, size: i32}
	// 							2:	list{content: i32*, size: i32}
	// 							3:	list{content: i32*, size: i32}
	//													^
	//													|
	//										contentIdx
	//
	// - Now GEP will first need to know which of these lists to address (elemIdx)
	// - Then GEP will want to know which field of the list to address (contentIdx)
	var elemPtr *ir.InstGetElementPtr
	if isList(listContentPtr) {
		contentIdx := constant.NewInt(types.I32, 0)
		elemPtr = indexIBBlock.NewGetElementPtr(listElemType, listContentPtr, index, contentIdx)
	} else {
		elemPtr = indexIBBlock.NewGetElementPtr(listElemType, listContentPtr, index)
	}
	elemPtr.LocalName = cg.uniqueNames.get("list_elem_ptr")
	indexIBBlock.NewRet(elemPtr)

	/* Raise runtime exception indexing uninitialized list */
	cg.exceptionHelper(initFalseBlock, "error_index_none")
	initFalseBlock.NewRet(constant.NewNull(types.NewPointer(listElemType)))

	/* Raise runtime exception negative index */
	cg.exceptionHelper(indexNegBlock, "error_index_neg")
	indexNegBlock.NewRet(constant.NewNull(types.NewPointer(listElemType)))

	/* Raise runtime exception index out of bounds */
	cg.exceptionHelper(indexOOBBlock, "error_index_oob")
	indexOOBBlock.NewRet(constant.NewNull(types.NewPointer(listElemType)))

	return listLenFunc
}

func (cg *CodeGenerator) definePrintString() *ir.Func {
	strArg := ir.NewParam("", types.I8Ptr)
	printString := cg.Module.NewFunc("printstr", types.I32, strArg)
	funcBlock := printString.NewBlock(cg.uniqueNames.get("entry"))

	formatStr := cg.useStringDef(funcBlock, "str_format_newline")
	printRes := funcBlock.NewCall(cg.functions["printf"], formatStr, strArg)
	funcBlock.NewRet(printRes)

	return printString
}

func (cg *CodeGenerator) definePrintInt() *ir.Func {
	intArg := ir.NewParam("", types.I32)
	printInt := cg.Module.NewFunc("printint", types.I32, intArg)
	funcBlock := printInt.NewBlock(cg.uniqueNames.get("entry"))

	formatStr := cg.useStringDef(funcBlock, "digit_format_newline")
	printRes := funcBlock.NewCall(cg.functions["printf"], formatStr, intArg)
	funcBlock.NewRet(printRes)

	return printInt
}

func (cg *CodeGenerator) definePrintBool() *ir.Func {
	arg := ir.NewParam("", types.I1)
	boolToStr := cg.Module.NewFunc("booltostr", types.I32, arg)

	entry := boolToStr.NewBlock(cg.uniqueNames.get("entry"))
	ifBlock := boolToStr.NewBlock(cg.uniqueNames.get("booltostr.then"))
	elseBlock := boolToStr.NewBlock(cg.uniqueNames.get("booltostr.else"))

	entry.NewCondBr(arg, ifBlock, elseBlock)

	trueStr := cg.useStringDef(ifBlock, "true_newline")
	truePrint := ifBlock.NewCall(cg.functions["printf"], trueStr)
	ifBlock.NewRet(truePrint)

	falseStr := cg.useStringDef(elseBlock, "false_newline")
	falsePrint := elseBlock.NewCall(cg.functions["printf"], falseStr)
	elseBlock.NewRet(falsePrint)

	return boolToStr
}

// NOTE: We can't use the below function because it uses alloca to
// allocate memory for a new string (alloca allocates this memory on
// the call stack of the function which gets freed after the function returns).
// The return of the function is then a pointer to unallocated memory.
/*
func (cg *CodeGenerator) defineStringConcat() *ir.Func {
	lhs := ir.NewParam("", types.I8Ptr)
	rhs := ir.NewParam("", types.I8Ptr)
	stringConcat := cg.Module.NewFunc("strconcat", types.I8Ptr, lhs, rhs)
	funcBlock := stringConcat.NewBlock(cg.uniqueNames.get("entry"))

	lhsLen := funcBlock.NewCall(cg.functions["strlen"], lhs)
	lhsLen.LocalName = cg.uniqueNames.get("lhs_len")
	rhsLen := funcBlock.NewCall(cg.functions["strlen"], rhs)
	rhsLen.LocalName = cg.uniqueNames.get("rhs_len")
	concatLen := funcBlock.NewAdd(lhsLen, rhsLen)
	concatLen.LocalName = cg.uniqueNames.get("concat_len")
	concatLen = funcBlock.NewAdd(concatLen, constant.NewInt(types.I32, 1))

	concatStr := &ir.InstAlloca{ElemType: types.I8, NElems: concatLen}
	concatStr.Type()
	funcBlock.Insts = append(funcBlock.Insts, concatStr)
	concatStr.LocalName = cg.uniqueNames.get("concat_str")

	copyRes := funcBlock.NewCall(cg.functions["strcpy"], concatStr, lhs)
	copyRes.LocalName = cg.uniqueNames.get("concat_copy_res")
	concatRes := funcBlock.NewCall(cg.functions["strcat"], concatStr, rhs)
	concatRes.LocalName = cg.uniqueNames.get("concat_append_res")

	funcBlock.NewRet(concatRes)

	// formatStr := cg.stringHelper(funcBlock, "%s%s\x00")
	// funcBlock.NewCall(cg.functions["sprintf"], concatStr, formatStr, str1, str2)

	return stringConcat
}
*/
