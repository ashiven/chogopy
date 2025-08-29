package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitBinaryExpr(binaryExpr *ast.BinaryExpr) {
	binaryExpr.Lhs.Visit(cg)
	lhsValue := cg.lastGenerated
	if isIdentOrIndex(binaryExpr.Lhs) {
		lhsValue = cg.LoadVal(lhsValue)
	}

	// Short circuit AND expr if lhs is literal False
	if _, ok := binaryExpr.Lhs.(*ast.LiteralExpr); ok {
		literalVal := binaryExpr.Lhs.(*ast.LiteralExpr).Value
		if binaryExpr.Op == "and" && literalVal == false {
			cg.lastGenerated = cg.NewLiteral(false)
			return
		}
	}

	// Short circuit OR expr if lhs is literal True
	if _, ok := binaryExpr.Lhs.(*ast.LiteralExpr); ok {
		literalVal := binaryExpr.Lhs.(*ast.LiteralExpr).Value
		if binaryExpr.Op == "or" && literalVal == true {
			cg.lastGenerated = cg.NewLiteral(true)
			return
		}
	}

	binaryExpr.Rhs.Visit(cg)
	rhsValue := cg.lastGenerated
	if isIdentOrIndex(binaryExpr.Rhs) {
		rhsValue = cg.LoadVal(rhsValue)
	}

	var resVal value.Value

	switch binaryExpr.Op {
	case "and":
		resVal = cg.currentBlock.NewAnd(lhsValue, rhsValue)
	case "or":
		resVal = cg.currentBlock.NewOr(lhsValue, rhsValue)
	case "%":
		// TODO: this is broken for negative values for the same
		// reason that div below is broken
		resVal = cg.currentBlock.NewSRem(lhsValue, rhsValue)
	case "*":
		resVal = cg.currentBlock.NewMul(lhsValue, rhsValue)
	case "//":
		// TODO: implement floor div for negative values:
		// (if the div result is negative, it will just be rounded
		// to the next whole number in a positive direction while
		// we want this direction to still remain negative)
		resVal = cg.currentBlock.NewSDiv(lhsValue, rhsValue)
	case "+":
		// TODO: string/list concat and updating lengths in cg.lengths and variables.length
		if _, ok := binaryExpr.TypeHint.(ast.ListAttribute); ok {
			elemTypeAttr := binaryExpr.TypeHint.(ast.ListAttribute).ElemType
			elemType := cg.attrToType(elemTypeAttr)
			resVal = cg.concatLists(lhsValue, rhsValue, elemType)
			break
		} else if binaryExpr.TypeHint == ast.String {
			resVal = cg.concatStrings(lhsValue, rhsValue)
			break
		}
		resVal = cg.currentBlock.NewAdd(lhsValue, rhsValue)
	case "-":
		resVal = cg.currentBlock.NewSub(lhsValue, rhsValue)
	case "<":
		resVal = cg.currentBlock.NewICmp(enum.IPredSLT, lhsValue, rhsValue)
	case "<=":
		resVal = cg.currentBlock.NewICmp(enum.IPredSLE, lhsValue, rhsValue)
	case ">":
		resVal = cg.currentBlock.NewICmp(enum.IPredSGT, lhsValue, rhsValue)
	case ">=":
		resVal = cg.currentBlock.NewICmp(enum.IPredSGE, lhsValue, rhsValue)
	case "==":
		resVal = cg.currentBlock.NewICmp(enum.IPredEQ, lhsValue, rhsValue)
	case "!=":
		resVal = cg.currentBlock.NewICmp(enum.IPredNE, lhsValue, rhsValue)
	case "is":
		// TODO: this should compare the addresses of lhs and rhs but since we are loading their
		// values above (cg.Load(lhs)...) we are actually comparing values here (incorrect)
		resVal = cg.currentBlock.NewICmp(enum.IPredEQ, lhsValue, rhsValue)
	}

	cg.lastGenerated = resVal
}

func (cg *CodeGenerator) getLength(val value.Value) int {
	if _, ok := cg.lengths[val]; ok {
		return cg.lengths[val]
	}
	return cg.variables[val.Ident()[1:]].length
}

func (cg *CodeGenerator) concatStrings(lhs value.Value, rhs value.Value) value.Value {
	concatPtr := cg.currentBlock.NewAlloca(types.I8)
	lhsLen := cg.currentBlock.NewCall(cg.functions["len"], lhs)
	rhsLen := cg.currentBlock.NewCall(cg.functions["len"], rhs)
	concatLen := cg.currentBlock.NewAdd(lhsLen, rhsLen)

	concatenator := cg.NewLiteral("%s%s")
	cg.currentBlock.NewCall(cg.functions["snprintf"], concatPtr, concatLen, concatenator, lhs, rhs)
	return concatPtr
}

func (cg *CodeGenerator) concatLists(lhs value.Value, rhs value.Value, elemType types.Type) value.Value {
	concatListPtr := cg.currentBlock.NewAlloca(elemType)
	concatListPtr.LocalName = cg.uniqueNames.get("concat_list_ptr")
	concatListLength := 0

	// TODO: we need a method to get the lengths of lists/strings at runtime
	for i := range cg.getLength(lhs) {
		index := constant.NewInt(types.I64, int64(i))
		elemPtr := cg.currentBlock.NewGetElementPtr(lhs.Type().(*types.PointerType).ElemType, lhs, index)
		elem := cg.currentBlock.NewLoad(lhs.Type().(*types.PointerType).ElemType, elemPtr)
		cg.NewStore(elem, concatListPtr)
		concatListLength++
	}

	for i := range cg.getLength(rhs) {
		index := constant.NewInt(types.I64, int64(i))
		elemPtr := cg.currentBlock.NewGetElementPtr(rhs.Type().(*types.PointerType).ElemType, rhs, index)
		elem := cg.currentBlock.NewLoad(rhs.Type().(*types.PointerType).ElemType, elemPtr)
		cg.NewStore(elem, concatListPtr)
		concatListLength++
	}

	cg.lengths[concatListPtr] = concatListLength
	return concatListPtr
}
