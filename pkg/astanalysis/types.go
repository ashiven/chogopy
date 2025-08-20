package astanalysis

import "chogopy/pkg/parser"

type Types struct {
	parser.BaseVisitor
}

func (t *Types) Analyze(program *parser.Program) {
	program.Visit(t)
}
