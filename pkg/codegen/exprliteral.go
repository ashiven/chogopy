package codegen

import "chogopy/pkg/ast"

func (cg *CodeGenerator) VisitLiteralExpr(literalExpr *ast.LiteralExpr) {
	literal := cg.NewLiteral(literalExpr.Value)
	cg.lastGenerated = literal
}
