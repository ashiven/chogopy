// Package astanalysis implements multiple AST analysis passes for the chogopy compiler
package astanalysis

import (
	"chogopy/pkg/parser"
	"fmt"
	"os"
)

type AssignTargets struct {
	parser.BaseVisitor
}

func (at *AssignTargets) Analyze(program *parser.Program) {
	program.Visit(at)

	for _, definition := range program.Definitions {
		definition.Visit(at)
	}
	for _, statement := range program.Statements {
		statement.Visit(at)
	}
}

func (at *AssignTargets) VisitAssignStmt(assignStmt *parser.AssignStmt) {
	if assignStmt.Target.Name() != "IdentExpr" && assignStmt.Target.Name() != "IndexExpr" {
		fmt.Printf("Semantic Error: Found %s as the left hand side of an assignment.\n", assignStmt.Target.Name())
		fmt.Println("Expected variable name or index expression.")
		os.Exit(0)
	}

	assignStmt.Value.Visit(at)
}
