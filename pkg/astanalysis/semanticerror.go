package astanalysis

import (
	"fmt"
	"os"
)

type NameScopeSemanticErrorKind int

const (
	IdentifierAlreadyDefined NameScopeSemanticErrorKind = iota
	IdentifierUndefined
	IdentifierUndeclaredParentScope
	IdentifierUndeclaredGlobalScope
	AssignToUndeclaredVariable
)

func nameSemanticError(errorKind NameScopeSemanticErrorKind, name string) {
}

type TypeSemanticErrorKind int

const (
	NotAssignmentCompatible TypeSemanticErrorKind = iota
	UnexpectedType
	ExpectedListType
	UnknownIdentifierUsed
	ExpectedVariableIdentifier
	ExpectedFunctionIdentifier
	ExpectedNonNoneListType
	IsBinaryExpectedTwoObjectTypes
	FunctionCallArgumentMismatch
)

func typeSemanticError(errorKind TypeSemanticErrorKind, t1 Type, t2 Type, defName string, funcArgs int, callArgs int) {
	switch errorKind {
	case NotAssignmentCompatible:
		fmt.Printf("Semantic Error: %s is not assignment compatible with %s", nameFromType(t1), nameFromType(t2))
	case UnexpectedType:
		fmt.Printf("Semantic Error: Expected %s but found %s", nameFromType(t1), nameFromType(t2))
	case ExpectedListType:
		fmt.Printf("Semantic Error: Expected list type but found %s", nameFromType(t1))
	case UnknownIdentifierUsed:
		fmt.Printf("Semantic Error: Unknown identifier used: %s", defName)
	case ExpectedVariableIdentifier:
		fmt.Printf("Semantic Error: Found function identifier: %s but expected variable identifier", defName)
	case ExpectedFunctionIdentifier:
		fmt.Printf("Semantic Error: Found variable identifier: %s but expected function identifier", defName)
	case ExpectedNonNoneListType:
		fmt.Printf("Semantic Error: Expected non-none list type in asssignment")
	case IsBinaryExpectedTwoObjectTypes:
		fmt.Printf("Semantic Error: Expected both operands to be of object type")
	case FunctionCallArgumentMismatch:
		fmt.Printf("Semantic Error: Expected %d arguments but got %d\n", funcArgs, callArgs)
	}
	os.Exit(0)
}
