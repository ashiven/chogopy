package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitUnaryExpr(unaryExpr *ast.UnaryExpr) {
	unaryExpr.Value.Visit(cg)
	unaryVal := cg.lastGenerated

	if isIdentOrIndex(unaryExpr.Value) {
		unaryVal = cg.LoadVal(unaryVal)
	}

	var resVal value.Value

	switch unaryExpr.Op {
	case "not":
		true_ := cg.NewLiteral(true)
		resVal = cg.currentBlock.NewXor(true_, unaryVal)

	case "-":
		zero := cg.NewLiteral(0)
		resVal = cg.currentBlock.NewSub(zero, unaryVal)
	}

	cg.lastGenerated = resVal
}
