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

func NewAssignTargetVisitor() *AssignTargetVisitor {
	assignTargetVisitor := &AssignTargetVisitor{}
	// We have to define the ChildVisitor attribute for the BaseVisitor so it
	// will be able to call visit methods that are overridden in this class.
	// All of this happens inside of BaseVisitor.VisitNext(operation Operation)
	// When the ChildVisitor has an overriding visit method defined, this method is called
	// instead of the method originally defined on the BaseVisitor.
	assignTargetVisitor.ChildVisitor = assignTargetVisitor

	return assignTargetVisitor
}

func (av *AssignTargetVisitor) VisitAssignStmt(as *parser.AssignStmt) {
	if as.Target.Name() != "IdentExpr" && as.Target.Name() != "IndexExpr" {
		fmt.Printf("Semantic Error: Found %s as the left hand side of an assignment.\n", as.Target.Name())
		fmt.Println("Expected variable name or index expression.")
		os.Exit(0)
	}

	av.VisitNext(as.Target)
	av.VisitNext(as.Value)
}
