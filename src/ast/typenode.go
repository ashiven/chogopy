package ast

type NamedType struct {
	name     string
	TypeName string
	Node
}

func (nt *NamedType) Name() string {
	if nt.name == "" {
		nt.name = "NamedType"
	}
	return nt.name
}

func (nt *NamedType) Visit(v Visitor) {
	v.VisitNamedType(nt)
}

type ListType struct {
	name     string
	ElemType Node
	Node
}

func (lt *ListType) Name() string {
	if lt.name == "" {
		lt.name = "ListType"
	}
	return lt.name
}

func (lt *ListType) Visit(v Visitor) {
	v.VisitListType(lt)

	if v.Traverse() {
		lt.ElemType.Visit(v)
	}
}
