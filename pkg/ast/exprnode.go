package ast

// TODO: We are currently simply converting Attributes to their string representation
// and setting this as the TypeHint but it may be better to set the Attribute directly and
// convert it to its string representation only when needed.

type LiteralExpr struct {
	name     string
	TypeHint TypeAttr
	Value    any
	Node
}

func (le *LiteralExpr) Name() string {
	if le.name == "" {
		le.name = "LiteralExpr"
	}
	return le.name
}

func (le *LiteralExpr) Visit(v Visitor) {
	v.VisitLiteralExpr(le)
}

type IdentExpr struct {
	name       string
	TypeHint   TypeAttr
	Identifier string
	Node
}

func (ie *IdentExpr) Name() string {
	if ie.name == "" {
		ie.name = "IdentExpr"
	}
	return ie.name
}

func (ie *IdentExpr) Visit(v Visitor) {
	v.VisitIdentExpr(ie)
	// We do not want to visit the type hint as it does not
	// belong to the AST even though it is an Node!
}

type UnaryExpr struct {
	name     string
	TypeHint TypeAttr
	Op       string
	Value    Node
	Node
}

func (ue *UnaryExpr) Name() string {
	if ue.name == "" {
		ue.name = "UnaryExpr"
	}
	return ue.name
}

func (ue *UnaryExpr) Visit(v Visitor) {
	v.VisitUnaryExpr(ue)
	if v.Traverse() {
		ue.Value.Visit(v)
	}
}

type BinaryExpr struct {
	name     string
	TypeHint TypeAttr
	Op       string
	Lhs      Node
	Rhs      Node
	Node
}

func (be *BinaryExpr) Name() string {
	if be.name == "" {
		be.name = "BinaryExpr"
	}
	return be.name
}

func (be *BinaryExpr) Visit(v Visitor) {
	v.VisitBinaryExpr(be)
	if v.Traverse() {
		be.Lhs.Visit(v)
		be.Rhs.Visit(v)
	}
}

type IfExpr struct {
	name      string
	TypeHint  TypeAttr
	Condition Node
	IfNode    Node
	ElseNode  Node
	Node
}

func (ie *IfExpr) Name() string {
	if ie.name == "" {
		ie.name = "IfExpr"
	}
	return ie.name
}

func (ie *IfExpr) Visit(v Visitor) {
	v.VisitIfExpr(ie)
	if v.Traverse() {
		ie.Condition.Visit(v)
		ie.IfNode.Visit(v)
		ie.ElseNode.Visit(v)
	}
}

type ListExpr struct {
	name     string
	TypeHint TypeAttr
	Elements []Node
	Node
}

func (le *ListExpr) Name() string {
	if le.name == "" {
		le.name = "ListExpr"
	}
	return le.name
}

func (le *ListExpr) Visit(v Visitor) {
	v.VisitListExpr(le)
	if v.Traverse() {
		for _, elem := range le.Elements {
			elem.Visit(v)
		}
	}
}

type CallExpr struct {
	name      string
	TypeHint  TypeAttr
	FuncName  string
	Arguments []Node
	Node
}

func (ce *CallExpr) Name() string {
	if ce.name == "" {
		ce.name = "CallExpr"
	}
	return ce.name
}

func (ce *CallExpr) Visit(v Visitor) {
	v.VisitCallExpr(ce)
	if v.Traverse() {
		for _, argument := range ce.Arguments {
			argument.Visit(v)
		}
	}
}

type IndexExpr struct {
	name     string
	TypeHint TypeAttr
	Value    Node
	Index    Node
	Node
}

func (ie *IndexExpr) Name() string {
	if ie.name == "" {
		ie.name = "IndexExpr"
	}
	return ie.name
}

func (ie *IndexExpr) Visit(v Visitor) {
	v.VisitIndexExpr(ie)
	if v.Traverse() {
		ie.Value.Visit(v)
		ie.Index.Visit(v)
	}
	// We do not want to visit the type hint as it does not
	// belong to the AST even though it is an Node!
}
