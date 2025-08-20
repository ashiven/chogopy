// Package astanalysis implements multiple AST analysis passes for the chogopy compiler
package astanalysis

import (
	"chogopy/pkg/parser"
	"fmt"
	"os"
)

type AssignTargetVisitor struct {
	parser.BaseVisitor
}

func (av *AssignTargetVisitor) Analyze(p *parser.Program) {
	p.Visit(av)
}

func (av *AssignTargetVisitor) VisitAssignStmt(as *parser.AssignStmt) {
	if as.Target.Name() != "IdentExpr" && as.Target.Name() != "IndexExpr" {
		fmt.Printf("Semantic Error: Found %s as the left hand side of an assignment.\n", as.Target.Name())
		fmt.Println("Expected variable name or index expression.")
		os.Exit(0)
	}
}
