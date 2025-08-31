package namescopes

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
	switch assignStmt.Target.(type) {
	case *ast.IdentExpr:
		return
	case *ast.IndexExpr:
		return
	}

	fmt.Printf("Semantic Error: Found %s as the left hand side of an assignment.\n", assignStmt.Target.Name())
	fmt.Println("Expected variable name or index expression.")
	os.Exit(0)
}
