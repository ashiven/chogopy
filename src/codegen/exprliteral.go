package codegen

import "chogopy/src/ast"

func (cg *CodeGenerator) VisitLiteralExpr(literalExpr *ast.LiteralExpr) {
	literal := cg.NewLiteral(literalExpr.Value)
	cg.lastGenerated = literal
}
