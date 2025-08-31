package typechecking

import (
	"chogopy/pkg/ast"
	"fmt"
	"log"
)

type Type any

type BasicType struct {
	typeName string
	Type
}

type ListType struct {
	elemType Type
	Type
}

type BottomType struct {
	typeName string
	Type
}

type ObjectType struct {
	typeName string
	Type
}

type FunctionType struct {
	paramTypes []Type
	returnType Type
}

var (
	intType    = BasicType{typeName: "int"}
	boolType   = BasicType{typeName: "bool"}
	strType    = BasicType{typeName: "str"}
	noneType   = BasicType{typeName: "<None>"}
	emptyType  = BasicType{typeName: "<Empty>"}
	bottomType = BottomType{typeName: "bottom"}
	objectType = ObjectType{typeName: "object"}
)

func typeFromNode(node ast.Node) Type {
	switch node := node.(type) {
	case *ast.NamedType:
		switch node.TypeName {
		case "int":
			return intType
		case "bool":
			return boolType
		case "str":
			return strType
		case "<None>":
			return noneType
		case "<Empty>":
			return emptyType
		case "object":
			return objectType
		}

	case *ast.ListType:
		elemType := typeFromNode(node.ElemType)
		return ListType{elemType: elemType}
	}

	log.Fatalf("Expected Node but found %# v", node)
	return nil
}

func attrFromType(nodeType Type) ast.TypeAttr {
	switch nodeType {
	case intType:
		return ast.Integer
	case boolType:
		return ast.Boolean
	case strType:
		return ast.String
	case noneType:
		return ast.None
	case emptyType:
		return ast.Empty
	case objectType:
		return ast.Object
	}

	_, isListType := nodeType.(ListType)
	if isListType {
		elemType := attrFromType(nodeType.(ListType).elemType)
		return ast.ListAttribute{ElemType: elemType}
	}

	log.Fatalf("Expected Type but found %# v", nodeType)
	return nil
}

func nameFromType(nodeType Type) string {
	switch nodeType {
	case intType:
		return intType.typeName
	case boolType:
		return boolType.typeName
	case strType:
		return strType.typeName
	case noneType:
		return noneType.typeName
	case emptyType:
		return emptyType.typeName
	case objectType:
		return objectType.typeName
	case bottomType:
		return bottomType.typeName
	}

	_, isListType := nodeType.(ListType)
	if isListType {
		elemType := nameFromType(nodeType.(ListType).elemType)
		return fmt.Sprintf("List[%s]", elemType)
	}

	log.Fatalf("Expected Type but found %# v", nodeType)
	return ""
}

func join(t1 Type, t2 Type) Type {
	if isAssignmentCompatible(t1, t2) {
		return t2
	}
	if isAssignmentCompatible(t2, t1) {
		return t1
	}
	return objectType
}

func isSubType(t1 Type, t2 Type) bool {
	_, t1IsList := t1.(ListType)

	switch {
	case t1 == t2:
		return true
	case (t1 == intType || t1 == boolType || t1 == strType || t1IsList) && t2 == objectType:
		return true
	case t1 == noneType && t2 == objectType:
		return true
	case t1 == emptyType && t2 == objectType:
		return true
	case t1 == bottomType:
		return true
	}
	return false
}

func isAssignmentCompatible(t1 Type, t2 Type) bool {
	_, t1IsList := t1.(ListType)
	_, t2IsList := t2.(ListType)

	switch {
	case isSubType(t1, t2) && t1 != bottomType:
		return true
	case t1 == noneType && t2 != intType && t2 != boolType && t2 != strType && t2 != bottomType:
		return true
	case t1 == emptyType && t2IsList:
		return true
	case t1IsList && t2IsList && t1.(ListType).elemType == noneType && isAssignmentCompatible(noneType, t2.(ListType).elemType):
		return true
	case t1IsList && t2IsList && t1.(ListType).elemType == t2.(ListType).elemType:
		return true
	}
	return false
}

func checkAssignmentCompatible(t1 Type, t2 Type) {
	if !isAssignmentCompatible(t1, t2) {
		typeSemanticError(NotAssignmentCompatible, t1, t2, "", 0, 0)
	}
}

func checkType(found Type, expected Type) {
	if found != expected {
		typeSemanticError(UnexpectedType, expected, found, "", 0, 0)
	}
}

func checkListType(found Type) {
	_, foundIsList := found.(ListType)
	if !foundIsList {
		typeSemanticError(ExpectedListType, found, nil, "", 0, 0)
	}
}
