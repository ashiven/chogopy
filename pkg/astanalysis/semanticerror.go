package astanalysis

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

func nameSemanticError(errorKind NameScopeSemanticErrorKind, name string) {
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
	AssignTargetInvalid
)

func typeSemanticError(errorKind TypeSemanticErrorKind, t1 Type, t2 Type, defName string, funcArgs int, callArgs int) {
	switch errorKind {
	case NotAssignmentCompatible:
		fmt.Printf("Semantic Error: %s is not assignment compatible with %s\n", nameFromType(t1), nameFromType(t2))
	case UnexpectedType:
		fmt.Printf("Semantic Error: Expected %s but found %s\n", nameFromType(t1), nameFromType(t2))
	case ExpectedListType:
		fmt.Printf("Semantic Error: Expected list type but found %s\n", nameFromType(t1))
	case UnknownIdentifierUsed:
		fmt.Printf("Semantic Error: Unknown identifier used: %s\n", defName)
	case ExpectedVariableIdentifier:
		fmt.Printf("Semantic Error: Found function identifier: %s but expected variable identifier\n", defName)
	case ExpectedFunctionIdentifier:
		fmt.Printf("Semantic Error: Found variable identifier: %s but expected function identifier\n", defName)
	case ExpectedNonNoneListType:
		fmt.Printf("Semantic Error: Expected non-none list type in asssignment\n")
	case IsBinaryExpectedTwoObjectTypes:
		fmt.Printf("Semantic Error: Expected both operands to be of object type\n")
	case FunctionCallArgumentMismatch:
		fmt.Printf("Semantic Error: Expected %d arguments but got %d\n", funcArgs, callArgs)
	case AssignTargetInvalid:
		fmt.Printf("Semantic Error: Cannot assign to non-identifier or index expression\n")
	}
	os.Exit(0)
}
