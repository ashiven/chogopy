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

	if returnType == cg.types["none"] {
		cg.currentBlock.NewRet(cg.NewLiteral(nil))
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
	varType := cg.astTypeToType(varDef.TypedVar.(*ast.TypedVar).VarType)
	literalVal := varDef.Literal.(*ast.LiteralExpr).Value

	var literalConst constant.Constant
	switch varDef.Literal.(*ast.LiteralExpr).TypeHint {
	case ast.Integer:
		literalConst = constant.NewInt(types.I32, int64(literalVal.(int)))
	case ast.Boolean:
		literalConst = constant.NewBool(literalVal.(bool))
	case ast.String:
		literalConst = constant.NewCharArrayFromString(literalVal.(string) + "\x00")
	case ast.None:
		switch varType := varType.(type) {
		case *types.PointerType:
			literalConst = constant.NewNull(varType)
		case *types.IntType:
			literalConst = constant.NewInt(varType, int64(0))
		}
	case ast.Empty:
		literalConst = constant.NewNull(varType.(*types.PointerType))
	case ast.Object:
		literalConst = constant.NewNull(varType.(*types.PointerType))
	}

	newVar := cg.Module.NewGlobalDef(varName, literalConst)
	cg.variables[varName] = VarInfo{name: varName, elemType: newVar.Typ.ElemType, value: newVar}
}
