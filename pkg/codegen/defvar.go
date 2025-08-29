package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitVarDef(varDef *ast.VarDef) {
	varName := varDef.TypedVar.(*ast.TypedVar).VarName
	varType := cg.astTypeToType(varDef.TypedVar.(*ast.TypedVar).VarType)
	literalVal := varDef.Literal.(*ast.LiteralExpr).Value

	literalLength := 1
	var literalConst constant.Constant
	switch varDef.Literal.(*ast.LiteralExpr).TypeHint {
	case ast.Integer:
		literalConst = constant.NewInt(types.I32, int64(literalVal.(int)))
	case ast.Boolean:
		literalConst = constant.NewBool(literalVal.(bool))
	case ast.String:
		literalConst = constant.NewCharArrayFromString(literalVal.(string) + "\x00")
		literalLength = len(literalVal.(string)) + 1
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
	cg.variables[varName] = VarInfo{name: varName, elemType: newVar.Typ.ElemType, value: newVar, length: literalLength}
}
