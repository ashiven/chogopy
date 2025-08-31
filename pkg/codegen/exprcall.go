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
		printRes := cg.printGeneric(args[0])
		cg.lastGenerated = printRes
		return
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

func (cg *CodeGenerator) printGeneric(arg value.Value) value.Value {
	if hasType(arg, types.I32) {
		return cg.currentBlock.NewCall(cg.functions["printint"], arg)
	} else if hasType(arg, types.I1) {
		return cg.currentBlock.NewCall(cg.functions["printbool"], arg)
	} else if isString(arg) {
		return cg.currentBlock.NewCall(cg.functions["printstr"], arg)
	}
	log.Fatalln("Code Generation: print() expected an argument of type int, bool, or str")
	return nil
}
