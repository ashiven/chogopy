package codegen

import (
	"chogopy/src/ast"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitFuncDef(funcDef *ast.FuncDef) {
	params := []*ir.Param{}
	for _, paramNode := range funcDef.Parameters {
		paramName := paramNode.(*ast.TypedVar).VarName
		paramType := cg.astTypeToType(paramNode.(*ast.TypedVar).VarType)
		param := ir.NewParam(paramName, paramType)
		params = append(params, param)
	}

	returnType := cg.astTypeToType(funcDef.ReturnType)

	newFunction := cg.Module.NewFunc(funcDef.FuncName, returnType, params...)
	newBlock := newFunction.NewBlock(cg.uniqueNames.get("entry"))

	cg.functions[funcDef.FuncName] = newFunction
	cg.currentFunction = newFunction
	cg.currentBlock = newBlock

	for _, bodyNode := range funcDef.FuncBody {
		bodyNode.Visit(cg)
	}

	if returnType.Equal(types.NewPointer(cg.types["none"])) {
		cg.currentBlock.NewRet(cg.NewLiteral(nil))
	}
}
