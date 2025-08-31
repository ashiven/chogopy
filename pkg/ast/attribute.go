package ast

import "fmt"

// TypeAttr has as its purpose to equip each
// expression AST node with a type hint that can later
// be used when lowering the AST into a series of
// machine instructions or a flattened IR.
type TypeAttr interface {
	String() string
}

type BasicAttribute int

const (
	Integer BasicAttribute = iota
	Boolean
	String
	None
	Empty
	Object
)

type ListAttribute struct {
	ElemType TypeAttr
}

func (ba BasicAttribute) String() string {
	switch ba {
	case Integer:
		return "Integer"
	case Boolean:
		return "Boolean"
	case String:
		return "String"
	case None:
		return "None"
	case Empty:
		return "Empty"
	case Object:
		return "Object"
	}
	return ""
}

func (la ListAttribute) String() string {
	return fmt.Sprintf("List[%s]", la.ElemType.String())
}
