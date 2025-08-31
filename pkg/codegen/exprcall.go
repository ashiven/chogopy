package codegen

import (
	"chogopy/pkg/ast"
	"log"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitCallExpr(callExpr *ast.CallExpr) {
	args := []value.Value{}
	for _, arg := range callExpr.Arguments {
		arg.Visit(cg)
		argVal := cg.lastGenerated
		if isIdentOrIndex(arg) {
			argVal = cg.LoadVal(cg.lastGenerated)
		}
		args = append(args, argVal)
	}

	switch callExpr.FuncName {
	case "print":
		args = cg.convertPrintArgs(args)
	case "input":
		cg.readString()
		return
	case "len":
		cg.getLen(args)
		return
	}

	callee := cg.functions[callExpr.FuncName]
	callRes := cg.currentBlock.NewCall(callee, args...)
	callRes.LocalName = cg.uniqueNames.get("call_res")

	cg.lastGenerated = callRes
}

func (cg *CodeGenerator) getLen(args []value.Value) {
	for _, arg := range args {
		if isString(arg) {
			strLen := cg.currentBlock.NewCall(cg.functions["strlen"], arg)
			strLen.LocalName = cg.uniqueNames.get("str_len")
			cg.lastGenerated = strLen

			// TODO: for now we just assume if it isn't a string, it's a list but we need a more explicit check
		} else {
			zero := constant.NewInt(types.I32, 0)
			lenFieldIdx := constant.NewInt(types.I32, 1)
			listLenAddr := cg.currentBlock.NewGetElementPtr(
				arg.Type().(*types.PointerType).ElemType,
				arg,
				zero,
				lenFieldIdx,
			)
			listLenAddr.LocalName = cg.uniqueNames.get("list_len_addr")
			listLen := cg.currentBlock.NewLoad(types.I32, listLenAddr)
			listLen.LocalName = cg.uniqueNames.get("list_len")
			cg.lastGenerated = listLen
		}
	}
}

func (cg *CodeGenerator) readString() {
	/* i8* fgets(buf i8*, size i32, fd FILE*) */

	// fileDes := constant.NewInt(types.I32, 0)
	// readMode := cg.NewLiteral("r")
	// stdin := cg.currentBlock.NewCall(cg.functions["fdopen"], fileDes, readMode)
	// stdin.LocalName = cg.uniqueNames.get("stdin")

	// inputPtr := cg.currentBlock.NewAlloca(types.NewArray(10000, types.I8))
	// inputPtr.LocalName = cg.uniqueNames.get("input_ptr")
	// inputPtrSize := cg.NewLiteral(10000)

	// scanRes := cg.currentBlock.NewCall(cg.functions["fgets"], inputPtr, inputPtrSize, stdin)
	// scanRes.LocalName = cg.uniqueNames.get("fgets_res")
	// cg.lastGenerated = cg.LoadVal(inputPtr)

	/* int scanf(format i8*, buf i8*)  */

	format := cg.NewLiteral("%s")
	inputPtr := cg.currentBlock.NewAlloca(types.NewArray(MaxBufferSize, types.I8))
	inputPtr.LocalName = cg.uniqueNames.get("input_ptr")
	inputPtrCast := cg.toString(inputPtr)

	scanRes := cg.currentBlock.NewCall(cg.functions["scanf"], format, inputPtrCast)
	scanRes.LocalName = cg.uniqueNames.get("scan_res")

	cg.lastGenerated = cg.LoadVal(inputPtr)
}

// convertPrintArgs converts a list of argument values serving as input
// to a call of the print function so that they are printed correctly.
// For example, given an arg of type bool (I1), it converts it into the string "True"
// if its value is 1, or "False" if its value is 0.
func (cg *CodeGenerator) convertPrintArgs(args []value.Value) []value.Value {
	printArgs := []value.Value{}
	for _, arg := range args {
		if hasType(arg, types.I32) || hasType(arg, types.I32Ptr) {
			/* Integer print */
			digitStr := cg.NewLiteral("%d\n")
			argVal := cg.LoadVal(arg)
			printArgs = append(printArgs, digitStr)
			printArgs = append(printArgs, argVal)

		} else if hasType(arg, types.I1) || hasType(arg, types.I1Ptr) {
			/* Boolean print */
			argVal := cg.LoadVal(arg)
			argVal = cg.currentBlock.NewCall(cg.functions["booltostr"], argVal)
			printArgs = append(printArgs, argVal)

		} else if isString(arg) {
			/* String print */
			arg = cg.LoadVal(arg)
			newline := cg.NewLiteral("\n")
			arg = cg.concatStrings(arg, newline)
			printArgs = append(printArgs, arg)

		} else {
			log.Fatalln("Code Generation: print() expected an argument of type int, bool, or str")
		}
	}
	return printArgs
}
