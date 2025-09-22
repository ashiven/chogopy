package typechecks

import (
	"fmt"
	"os"
)

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

func semanticError(errorKind TypeSemanticErrorKind, t1 Type, t2 Type, defName string, funcArgs int, callArgs int) {
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
