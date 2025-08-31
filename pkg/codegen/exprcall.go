package codegen

import (
	"chogopy/pkg/ast"
	"log"

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
	case "len":
		lenRes := cg.getLen(args[0])
		cg.lastGenerated = lenRes
		return
	case "print":
		args = cg.convertPrintArgs(args)
	}

	callee := cg.functions[callExpr.FuncName]
	callRes := cg.currentBlock.NewCall(callee, args...)
	callRes.LocalName = cg.uniqueNames.get("call_res")

	cg.lastGenerated = callRes
}

func (cg *CodeGenerator) getLen(arg value.Value) value.Value {
	if isString(arg) {
		strLen := cg.currentBlock.NewCall(cg.functions["strlen"], arg)
		strLen.LocalName = cg.uniqueNames.get("str_len")
		return strLen

	} else if isPtrTo(arg, cg.types["list"]) {
		listLen := cg.currentBlock.NewCall(cg.functions["listlen"], arg)
		listLen.LocalName = cg.uniqueNames.get("list_len")
		return listLen
	}

	log.Fatalln("Code Generation: len() expected an argument of type str or list")
	return nil
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
