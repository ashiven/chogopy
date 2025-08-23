// Package astanalysis implements multiple AST analysis passes for the chogopy compiler
package astanalysis

import (
	"chogopy/pkg/ast"
	"fmt"
	"os"
)

type AssignTargets struct {
	ast.BaseVisitor
}

func (at *AssignTargets) Analyze(program *ast.Program) {
	program.Visit(at)
}

func (at *AssignTargets) VisitAssignStmt(assignStmt *ast.AssignStmt) {
	if assignStmt.Target.Name() != "IdentExpr" && assignStmt.Target.Name() != "IndexExpr" {
		fmt.Printf("Semantic Error: Found %s as the left hand side of an assignment.\n", assignStmt.Target.Name())
		fmt.Println("Expected variable name or index expression.")
		os.Exit(0)
	}
}
