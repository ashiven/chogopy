package astanalysis

import (
	"chogopy/pkg/ast"
)

type NameScopes struct {
	ast.BaseVisitor
}

func (ns *NameScopes) Analyze(program *ast.Program) {
	program.Visit(ns)
}
