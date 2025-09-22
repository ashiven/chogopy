package codegen

import (
	"chogopy/src/ast"
	"log"

	"github.com/llir/llvm/ir"
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
		lenRes.(*ir.InstCall).LocalName = cg.uniqueNames.get("call_res")
		cg.lastGenerated = lenRes
		return
	case "print":
		printRes := cg.printGeneric(args[0])
		printRes.(*ir.InstCall).LocalName = cg.uniqueNames.get("call_res")
		cg.lastGenerated = printRes
		return
	}

	callee := cg.functions[callExpr.FuncName]
	callRes := cg.currentBlock.NewCall(callee, args...)
	callRes.LocalName = cg.uniqueNames.get("call_res")

	if _, ok := callRes.Type().(*types.PointerType); ok {
		cg.heapAllocs = append(cg.heapAllocs, callRes)
	}

	cg.lastGenerated = callRes
}

func (cg *CodeGenerator) getLen(arg value.Value) value.Value {
	if isString(arg) {
		return cg.getStringLen(arg)
	} else if isList(arg) {
		return cg.getListLen(arg)
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
	log.Fatalln("Code Generation: print() expected an argument of type int, bool, or str but got:", arg.Type().String())
	return nil
}
