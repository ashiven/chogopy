package astanalysis

import "chogopy/pkg/parser"

type Names struct {
	parser.BaseVisitor
}

func (n *Names) Analyze(program *parser.Program) {
	program.Visit(n)
}
