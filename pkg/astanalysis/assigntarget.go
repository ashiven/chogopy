// Package astanalysis implements multiple AST analysis passes for the chogopy compiler
package astanalysis

import (
	"chogopy/pkg/parser"
	"fmt"
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
	fmt.Println("hello from the assign target visitor")
	fmt.Println(as.Name())
	fmt.Println("end communication")
	av.VisitNext(as.Target)
	av.VisitNext(as.Value)
}
