package scopes

import (
	"fmt"
	"os"
)

type NameScopeSemanticErrorKind int

const (
	IdentifierAlreadyDefined NameScopeSemanticErrorKind = iota
	IdentifierUndefined
	AssignTargetOutOfScope
	IdentifierNotInParentScope
	IdentifierNotInGlobalScope
)

func semanticError(errorKind NameScopeSemanticErrorKind, name string) {
	switch errorKind {
	case IdentifierAlreadyDefined:
		fmt.Printf("Semantic Error: Identifier %s already defined in the current context.\n", name)
	case IdentifierUndefined:
		fmt.Printf("Semantic Error: Identifier %s used that was not previously defined.\n", name)
	case AssignTargetOutOfScope:
		fmt.Printf("Semantic Error: Cannot assign to variable %s that was not declared in the current scope.\n", name)
	case IdentifierNotInParentScope:
		fmt.Printf("Semantic Error: Identifier %s not declared in valid parent scope.\n", name)
	case IdentifierNotInGlobalScope:
		fmt.Printf("Semantic Error: Identifier %s not declared in the global scope.\n", name)
	}
	os.Exit(0)
}
