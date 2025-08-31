// Package ast implements definitions for the AST nodes and the AST visitor
package ast

type Node interface {
	Name() string
	Visit(v Visitor)

	// TODO: It would be good to have this but is the effort for validation even beneficial
	// or will invalid nodes not already be prevented in the parser and in other checks?
	Validate() bool
}

type Program struct {
	name        string
	Definitions []Node
	Statements  []Node
	Node
}

func (p *Program) Name() string {
	if p.name == "" {
		p.name = "Program"
	}
	return p.name
}

func (p *Program) Visit(v Visitor) {
	v.VisitProgram(p)
	if v.Traverse() {
		for _, definition := range p.Definitions {
			definition.Visit(v)
		}
		for _, statement := range p.Statements {
			statement.Visit(v)
		}
	}
}
