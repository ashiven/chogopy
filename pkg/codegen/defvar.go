package codegen

import (
	"strings"

	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func (cg *CodeGenerator) VisitVarDef(varDef *ast.VarDef) {
	varName := varDef.TypedVar.(*ast.TypedVar).VarName
	literalConst := cg.getLiteralConst(varDef)

	switch cg.currentFunction {
	case cg.mainFunction:
		globalVar := cg.Module.NewGlobalDef(varName, literalConst)
		cg.setVar(
			VarInfo{name: varName, elemType: globalVar.Typ.ElemType, value: globalVar, init: literalConst},
		)

	default:
		localVar := cg.currentBlock.NewAlloca(literalConst.Type())
		cg.currentBlock.NewStore(literalConst, localVar)
		cg.setVar(
			VarInfo{name: varName, elemType: localVar.Typ.ElemType, value: localVar, init: literalConst},
		)
	}
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

		switch varType := varType.(type) {
		case *types.PointerType:
			/* None init for list* */
			if strings.Contains(varType.String(), "list") {
				listContentType := varType.ElemType.(*types.StructType).Fields[0]
				listContentNone := constant.NewNull(listContentType.(*types.PointerType))
				listLenNone := constant.NewInt(types.I32, 0)
				listInitNone := constant.NewBool(false)

				listNone := cg.Module.NewGlobalDef(
					cg.uniqueNames.get("list_none"),
					constant.NewStruct(varType.ElemType.(*types.StructType), listContentNone, listLenNone, listInitNone),
				)

				zero := constant.NewInt(types.I32, 0)
				literalConst = constant.NewGetElementPtr(listNone.Typ.ElemType, listNone, zero)

				/* None init for every other pointer type */
			} else {
				literalConst = constant.NewNull(varType)
			}

		/* None init for i32 and bool */
		case *types.IntType:
			literalConst = constant.NewInt(varType, 0)
		}

	case ast.Empty:
		literalConst = constant.NewNull(varType.(*types.PointerType))

	case ast.Object:
		literalConst = constant.NewNull(varType.(*types.PointerType))
	}

	return literalConst
}
