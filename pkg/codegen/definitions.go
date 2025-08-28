package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

/* Definitions */

func (cg *CodeGenerator) VisitFuncDef(funcDef *ast.FuncDef) {
	params := []*ir.Param{}
	for _, paramNode := range funcDef.Parameters {
		paramName := paramNode.(*ast.TypedVar).VarName
		paramType := astTypeToType(paramNode.(*ast.TypedVar).VarType)
		param := ir.NewParam(paramName, paramType)
		params = append(params, param)
	}

	returnType := astTypeToType(funcDef.ReturnType)

	newFunction := cg.Module.NewFunc(funcDef.FuncName, returnType, params...)
	newBlock := newFunction.NewBlock(cg.uniqueNames.get("entry"))

	cg.funcDefs[funcDef.FuncName] = newFunction
	cg.currentFunction = newFunction
	cg.currentBlock = newBlock

	for _, bodyNode := range funcDef.FuncBody {
		bodyNode.Visit(cg)
	}

	if returnType == types.Void {
		cg.currentBlock.NewRet(nil)
	}
}

func (cg *CodeGenerator) VisitTypedVar(typedVar *ast.TypedVar) {
	/* no op */
}

func (cg *CodeGenerator) VisitGlobalDecl(globalDecl *ast.GlobalDecl) {
	/* no op */
}

func (cg *CodeGenerator) VisitNonLocalDecl(nonLocalDecl *ast.NonLocalDecl) {
	/* no op */
}

func (cg *CodeGenerator) VisitVarDef(varDef *ast.VarDef) {
	varName := varDef.TypedVar.(*ast.TypedVar).VarName
	varType := astTypeToType(varDef.TypedVar.(*ast.TypedVar).VarType)
	literalVal := varDef.Literal.(*ast.LiteralExpr).Value

	var literalConst constant.Constant
	switch varDef.Literal.(*ast.LiteralExpr).TypeHint {
	case ast.Integer:
		literalConst = constant.NewInt(types.I32, int64(literalVal.(int)))
	case ast.Boolean:
		literalConst = constant.NewBool(literalVal.(bool))
	case ast.String:
		literalConst = constant.NewCharArrayFromString(literalVal.(string))
		// TODO: we want strings to represented internally as I8*
		// but LLVM will create a type like [4 x I8] which we need to bitcast into I8*
		// For the same reason, the newly created var will have type [4 x I8]* with elemType I8.
		// We need to adjust this so the variable has type I8** with elemType I8*
	case ast.None:
		switch varType := varType.(type) {
		case *types.PointerType:
			literalConst = constant.NewNull(varType)
		case *types.IntType:
			literalConst = constant.NewInt(varType, int64(0))
		}
	}

	newVar := cg.Module.NewGlobalDef(varName, literalConst)
	cg.varDefs[varName] = VarDef{name: varName, elemType: newVar.Typ.ElemType, value: newVar}
}
