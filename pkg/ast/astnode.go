// Package ast implements definitions for the AST nodes and the AST visitor
package ast

import "fmt"

/* Types */

type Node interface {
	Name() string
	Visit(v Visitor)
	// TODO: It would be good to have but is the effort for validation even beneficial
	// or will invalid nodes not already be prevented in the parser?
	Validate() bool
}

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
	Length   int
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

/* Definitions */

type Program struct {
	name        string
	Definitions []Node
	Statements  []Node
	Node
}

func (p *Program) Name() string {
	if p.name == "" {
		p.name = "Program"
	}
	return p.name
}

func (p *Program) Visit(v Visitor) {
	v.VisitProgram(p)
	if v.Traverse() {
		for _, definition := range p.Definitions {
			definition.Visit(v)
		}
		for _, statement := range p.Statements {
			statement.Visit(v)
		}
	}
}

type FuncDef struct {
	name       string
	FuncName   string
	Parameters []Node
	FuncBody   []Node
	ReturnType Node
	Node
}

func (fd *FuncDef) Name() string {
	if fd.name == "" {
		fd.name = "FuncDef"
	}
	return fd.name
}

func (fd *FuncDef) Visit(v Visitor) {
	v.VisitFuncDef(fd)
	if v.Traverse() {
		for _, param := range fd.Parameters {
			param.Visit(v)
		}
		for _, bodyNode := range fd.FuncBody {
			bodyNode.Visit(v)
		}
		fd.ReturnType.Visit(v)
	}
}

type TypedVar struct {
	name    string
	VarName string
	VarType Node
	Node
}

func (tv *TypedVar) Name() string {
	if tv.name == "" {
		tv.name = "TypedVar"
	}
	return tv.name
}

func (tv *TypedVar) Visit(v Visitor) {
	v.VisitTypedVar(tv)
	if v.Traverse() {
		tv.VarType.Visit(v)
	}
}

type GlobalDecl struct {
	name     string
	DeclName string
	Node
}

func (gd *GlobalDecl) Name() string {
	if gd.name == "" {
		gd.name = "GlobalDecl"
	}
	return gd.name
}

func (gd *GlobalDecl) Visit(v Visitor) {
	v.VisitGlobalDecl(gd)
}

type NonLocalDecl struct {
	name     string
	DeclName string
	Node
}

func (nl *NonLocalDecl) Name() string {
	if nl.name == "" {
		nl.name = "NonLocalDecl"
	}
	return nl.name
}

func (nl *NonLocalDecl) Visit(v Visitor) {
	v.VisitNonLocalDecl(nl)
}

type VarDef struct {
	name     string
	TypedVar Node
	Literal  Node
	Node
}

func (vd *VarDef) Name() string {
	if vd.name == "" {
		vd.name = "VarDef"
	}
	return vd.name
}

func (vd *VarDef) Visit(v Visitor) {
	v.VisitVarDef(vd)
	if v.Traverse() {
		vd.TypedVar.Visit(v)
		vd.Literal.Visit(v)
	}
}

/* Statements */

type IfStmt struct {
	name      string
	Condition Node
	IfBody    []Node
	ElseBody  []Node
	Node
}

func (is *IfStmt) Name() string {
	if is.name == "" {
		is.name = "IfStmt"
	}
	return is.name
}

func (is *IfStmt) Visit(v Visitor) {
	v.VisitIfStmt(is)
	if v.Traverse() {
		is.Condition.Visit(v)
		for _, ifBodyNode := range is.IfBody {
			ifBodyNode.Visit(v)
		}
		for _, elseBodyNode := range is.ElseBody {
			elseBodyNode.Visit(v)
		}
	}
}

type WhileStmt struct {
	name      string
	Condition Node
	Body      []Node
	Node
}

func (ws *WhileStmt) Name() string {
	if ws.name == "" {
		ws.name = "WhileStmt"
	}
	return ws.name
}

func (ws *WhileStmt) Visit(v Visitor) {
	v.VisitWhileStmt(ws)
	if v.Traverse() {
		for _, bodyNode := range ws.Body {
			bodyNode.Visit(v)
		}
	}
}

type ForStmt struct {
	name     string
	IterName string
	Iter     Node
	Body     []Node
	Node
}

func (fs *ForStmt) Name() string {
	if fs.name == "" {
		fs.name = "ForStmt"
	}
	return fs.name
}

func (fs *ForStmt) Visit(v Visitor) {
	v.VisitForStmt(fs)
	if v.Traverse() {
		fs.Iter.Visit(v)
		for _, bodyNode := range fs.Body {
			bodyNode.Visit(v)
		}
	}
}

type PassStmt struct {
	name string
	Node
}

func (ps *PassStmt) Name() string {
	if ps.name == "" {
		ps.name = "PassStmt"
	}
	return ps.name
}

func (ps *PassStmt) Visit(v Visitor) {
	v.VisitPassStmt(ps)
}

type ReturnStmt struct {
	name      string
	ReturnVal Node
	Node
}

func (rs *ReturnStmt) Name() string {
	if rs.name == "" {
		rs.name = "ReturnStmt"
	}
	return rs.name
}

func (rs *ReturnStmt) Visit(v Visitor) {
	v.VisitReturnStmt(rs)
	if v.Traverse() && rs.ReturnVal != nil {
		rs.ReturnVal.Visit(v)
	}
}

type AssignStmt struct {
	name   string
	Target Node
	Value  Node
	Node
}

func (as *AssignStmt) Name() string {
	if as.name == "" {
		as.name = "AssignStmt"
	}
	return as.name
}

func (as *AssignStmt) Visit(v Visitor) {
	v.VisitAssignStmt(as)
	if v.Traverse() {
		as.Target.Visit(v)
		as.Value.Visit(v)
	}
}

// TODO: We are currently simply converting Attributes to their string representation
// and setting this as the TypeHint but it may be better to set the Attribute directly and
// convert it to its string representation only when needed.

/* Expressions */

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
