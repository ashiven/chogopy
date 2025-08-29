package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitCallExpr(callExpr *ast.CallExpr) {
	args := []value.Value{}
	for _, arg := range callExpr.Arguments {
		arg.Visit(cg)
		args = append(args, cg.lastGenerated)
	}

	switch callExpr.FuncName {
	case "print":
		args = cg.convertPrintArgs(args)
	case "input":
		cg.readString()
		return
	}

	callee := cg.functions[callExpr.FuncName]
	callRes := cg.currentBlock.NewCall(callee, args...)
	callRes.LocalName = cg.uniqueNames.get("call_res")

	cg.lastGenerated = callRes
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
	inputPtr := cg.currentBlock.NewAlloca(types.NewArray(10000, types.I8))
	inputPtr.LocalName = cg.uniqueNames.get("input_ptr")

	scanRes := cg.currentBlock.NewCall(cg.functions["scanf"], format, inputPtr)
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
		if hasType(arg, types.I32) || isPtrTo(arg, types.I32) {

			/* Integer print */
			digitStr := cg.NewLiteral("%d\n")
			argVal := cg.LoadVal(arg)
			printArgs = append(printArgs, digitStr)
			printArgs = append(printArgs, argVal)
		} else if hasType(arg, types.I1) || isPtrTo(arg, types.I1) {

			/* Boolean print
			> Uses a ternary to determine whether to print "True" or "False" based on the given I1 argument */
			ifBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("boolprint.then"))
			elseBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("boolprint.else"))
			exitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("boolprint.exit"))

			resAlloc := cg.currentBlock.NewAlloca(types.I8Ptr)
			resAlloc.LocalName = cg.uniqueNames.get("boolprint_res_ptr")

			cond := arg
			cond = cg.LoadVal(cond)
			cg.currentBlock.NewCondBr(cond, ifBlock, elseBlock)

			cg.currentBlock = ifBlock
			ifBlockRes := cg.NewLiteral("True\n")
			cg.NewStore(ifBlockRes, resAlloc)
			cg.currentBlock.NewBr(exitBlock)

			cg.currentBlock = elseBlock
			elseBlockRes := cg.NewLiteral("False\n")
			cg.NewStore(elseBlockRes, resAlloc)
			cg.currentBlock.NewBr(exitBlock)

			cg.currentBlock = exitBlock

			argVal := cg.LoadVal(resAlloc)
			printArgs = append(printArgs, argVal)

		} else {

			/* String print */
			arg = cg.LoadVal(arg)
			newline := cg.NewLiteral("\n")
			arg = cg.concatStrings(arg, newline)
			printArgs = append(printArgs, arg)
		}
	}
	return printArgs
}
