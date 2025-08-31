package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitVarDef(varDef *ast.VarDef) {
	varName := varDef.TypedVar.(*ast.TypedVar).VarName
	literalLength := cg.getLiteralLength(varDef)
	literalConst := cg.getLiteralConst(varDef)

	switch cg.currentFunction {
	case cg.mainFunction:
		globalVar := cg.Module.NewGlobalDef(varName, literalConst)
		cg.setVar(
			VarInfo{name: varName, elemType: globalVar.Typ.ElemType, value: globalVar, length: literalLength},
		)

	default:
		localVar := cg.currentBlock.NewAlloca(literalConst.Type())
		cg.currentBlock.NewStore(literalConst, localVar)
		cg.setVar(
			VarInfo{name: varName, elemType: localVar.Typ.ElemType, value: localVar, length: literalLength},
		)
	}
}

func (cg *CodeGenerator) getLiteralLength(varDef *ast.VarDef) int {
	if varDef.Literal.(*ast.LiteralExpr).TypeHint == ast.String {
		return len(varDef.Literal.(*ast.LiteralExpr).Value.(string)) + 1
	}
	return 1
}

func (cg *CodeGenerator) getLiteralConst(varDef *ast.VarDef) constant.Constant {
	varType := cg.astTypeToType(varDef.TypedVar.(*ast.TypedVar).VarType)
	literalVal := varDef.Literal.(*ast.LiteralExpr).Value

	var literalConst constant.Constant
	switch varDef.Literal.(*ast.LiteralExpr).TypeHint {
	case ast.Integer:
		literalConst = constant.NewInt(types.I32, int64(literalVal.(int)))

	case ast.Boolean:
		literalConst = constant.NewBool(literalVal.(bool))

	case ast.String:
		strConst := constant.NewCharArrayFromString(literalVal.(string) + "\x00")
		strDef := cg.Module.NewGlobalDef(cg.uniqueNames.get("str"), strConst)

		zero := constant.NewInt(types.I32, 0)
		literalConst = constant.NewGetElementPtr(strDef.Typ.ElemType, strDef, zero, zero)

	case ast.None:
		literalConst = constant.NewZeroInitializer(varType)
		// switch varType := varType.(type) {
		// case *types.StructType:
		// 	literalConst = constant.NewZeroInitializer(varType)
		// case *types.PointerType:
		// 	literalConst = constant.NewNull(varType)
		// case *types.IntType:
		// 	literalConst = constant.NewInt(varType, int64(0))
		// }

	case ast.Empty:
		literalConst = constant.NewNull(varType.(*types.PointerType))

	case ast.Object:
		literalConst = constant.NewNull(varType.(*types.PointerType))
	}

	return literalConst
}
