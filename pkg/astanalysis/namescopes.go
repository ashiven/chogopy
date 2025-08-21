package astanalysis

import "chogopy/pkg/parser"

type NameScopes struct {
	parser.BaseVisitor
}

func (ns *NameScopes) Analyze(program *parser.Program) {
	program.Visit(ns)
}
