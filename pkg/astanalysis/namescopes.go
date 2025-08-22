package astanalysis

import "chogopy/pkg/parser"

type NameScopes struct {
	parser.BaseVisitor
}

func (ns *NameScopes) Analyze(program *parser.Program) {
	program.Visit(ns)

	for _, definition := range program.Definitions {
		definition.Visit(ns)
	}
	for _, statement := range program.Statements {
		statement.Visit(ns)
	}
}
