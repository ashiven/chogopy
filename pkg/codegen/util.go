package codegen

import (
	"chogopy/pkg/ast"
)

func isIdentOrIndex(astNode ast.Node) bool {
	switch astNode.(type) {
	case *ast.IdentExpr:
		return true
	case *ast.IndexExpr:
		return true
	}
	return false
}
