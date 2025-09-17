package codegen

import (
	"chogopy/pkg/ast"

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

	if cg.shortCircuit(binaryExpr) {
		return
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
		resVal = cg.floorRem(lhsValue, rhsValue)
	case "*":
		resVal = cg.currentBlock.NewMul(lhsValue, rhsValue)
	case "//":
		resVal = cg.currentBlock.NewCall(cg.functions["floordiv"], lhsValue, rhsValue)
	case "+":
		if cg.concat(binaryExpr, lhsValue, rhsValue) {
			return
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
		if cg.stringEqual(lhsValue, rhsValue) {
			return
		}
		resVal = cg.currentBlock.NewICmp(enum.IPredEQ, lhsValue, rhsValue)
	case "!=":
		if cg.stringNotEqual(lhsValue, rhsValue) {
			return
		}
		resVal = cg.currentBlock.NewICmp(enum.IPredNE, lhsValue, rhsValue)
	case "is":
		_, lhsIsPtr := lhsValue.Type().(*types.PointerType)
		_, rhsIsPtr := rhsValue.Type().(*types.PointerType)

		if lhsIsPtr && rhsIsPtr && !lhsValue.Type().Equal(rhsValue.Type()) {
			lhsCast := cg.currentBlock.NewBitCast(lhsValue, rhsValue.Type())
			resVal = cg.currentBlock.NewICmp(enum.IPredEQ, lhsCast, rhsValue)

		} else {
			resVal = cg.currentBlock.NewICmp(enum.IPredEQ, lhsValue, rhsValue)
		}
	}

	cg.lastGenerated = resVal
}

func (cg *CodeGenerator) floorRem(lhs value.Value, rhs value.Value) value.Value {
	/* rem = lhs - rhs * floorDiv(lhs, rhs) */

	floorDiv := cg.currentBlock.NewCall(cg.functions["floordiv"], lhs, rhs)
	rhsMult := cg.currentBlock.NewMul(rhs, floorDiv)
	floorRem := cg.currentBlock.NewSub(lhs, rhsMult)

	return floorRem
}

func (cg *CodeGenerator) shortCircuit(binaryExpr *ast.BinaryExpr) bool {
	if _, ok := binaryExpr.Lhs.(*ast.LiteralExpr); ok {
		literalVal := binaryExpr.Lhs.(*ast.LiteralExpr).Value
		if binaryExpr.Op == "and" && literalVal == false {
			cg.lastGenerated = cg.NewLiteral(false)
			return true
		}
	}

	if _, ok := binaryExpr.Lhs.(*ast.LiteralExpr); ok {
		literalVal := binaryExpr.Lhs.(*ast.LiteralExpr).Value
		if binaryExpr.Op == "or" && literalVal == true {
			cg.lastGenerated = cg.NewLiteral(true)
			return true
		}
	}

	return false
}

func (cg *CodeGenerator) concat(binaryExpr *ast.BinaryExpr, lhs value.Value, rhs value.Value) bool {
	if _, ok := binaryExpr.TypeHint.(ast.ListAttribute); ok {
		listType := cg.attrToType(binaryExpr.TypeHint).(*types.PointerType).ElemType
		cg.lastGenerated = cg.concatLists(lhs, rhs, listType)
		return true

	} else if binaryExpr.TypeHint == ast.String {
		cg.lastGenerated = cg.concatStrings(lhs, rhs)
		return true
	}

	return false
}

// func (cg CodeGenerator) getStrLiteral(node ast.Node) string {
// 	if isIdentOrIndex(node) {
// 		varInfo, _ := cg.getVar(node.(*ast.IdentExpr).Identifier)
// 		initConst := varInfo.init.(*constant.ExprGetElementPtr)
// 		charArr := initConst.Src.(*ir.Global).Init.(*constant.CharArray).X
// 		strLiteral := string(charArr[:len(charArr)-1]) // Remove '/0' from the char array
// 		return strLiteral
//
// 	} else {
// 		return node.(*ast.LiteralExpr).Value.(string)
// 	}
// }

// 		lhsString := cg.getStrLiteral(binaryExpr.Lhs)
// 		rhsString := cg.getStrLiteral(binaryExpr.Rhs)
//
// 		cg.lastGenerated = cg.concatStrings(lhsString, rhsString)
