package ast

type Visitor interface {
	Traverse() bool
	VisitNamedType(nt *NamedType)
	VisitListType(lt *ListType)
	VisitProgram(p *Program)
	VisitFuncDef(fd *FuncDef)
	VisitTypedVar(tv *TypedVar)
	VisitGlobalDecl(gd *GlobalDecl)
	VisitNonLocalDecl(nl *NonLocalDecl)
	VisitVarDef(vd *VarDef)
	VisitIfStmt(is *IfStmt)
	VisitWhileStmt(ws *WhileStmt)
	VisitForStmt(fs *ForStmt)
	VisitPassStmt(ps *PassStmt)
	VisitReturnStmt(rs *ReturnStmt)
	VisitAssignStmt(as *AssignStmt)
	VisitLiteralExpr(le *LiteralExpr)
	VisitIdentExpr(ie *IdentExpr)
	VisitUnaryExpr(ue *UnaryExpr)
	VisitBinaryExpr(be *BinaryExpr)
	VisitIfExpr(ie *IfExpr)
	VisitListExpr(le *ListExpr)
	VisitCallExpr(ce *CallExpr)
	VisitIndexExpr(ie *IndexExpr)
}

type BaseVisitor struct{}

// Traverse determines whether the visitor
// should continue traversing down the tree after it has visited a node.
// Each node implements a Visit() method which
// checks whether Traverse() returns true and if it does,
// will call the Visit() method on each of its child nodes.
func (bv *BaseVisitor) Traverse() bool {
	return true
}

func (bv *BaseVisitor) VisitNamedType(nt *NamedType) {
}

func (bv *BaseVisitor) VisitListType(lt *ListType) {
}

func (bv *BaseVisitor) VisitProgram(p *Program) {
}

func (bv *BaseVisitor) VisitFuncDef(fd *FuncDef) {
}

func (bv *BaseVisitor) VisitTypedVar(tv *TypedVar) {
}

func (bv *BaseVisitor) VisitGlobalDecl(gd *GlobalDecl) {
}

func (bv *BaseVisitor) VisitNonLocalDecl(nl *NonLocalDecl) {
}

func (bv *BaseVisitor) VisitVarDef(vd *VarDef) {
}

func (bv *BaseVisitor) VisitIfStmt(is *IfStmt) {
}

func (bv *BaseVisitor) VisitWhileStmt(ws *WhileStmt) {
}

func (bv *BaseVisitor) VisitForStmt(fs *ForStmt) {
}

func (bv *BaseVisitor) VisitPassStmt(ps *PassStmt) {
}

func (bv *BaseVisitor) VisitReturnStmt(rs *ReturnStmt) {
}

func (bv *BaseVisitor) VisitAssignStmt(as *AssignStmt) {
}

func (bv *BaseVisitor) VisitLiteralExpr(le *LiteralExpr) {
}

func (bv *BaseVisitor) VisitIdentExpr(ie *IdentExpr) {
}

func (bv *BaseVisitor) VisitUnaryExpr(ue *UnaryExpr) {
}

func (bv *BaseVisitor) VisitBinaryExpr(be *BinaryExpr) {
}

func (bv *BaseVisitor) VisitIfExpr(ie *IfExpr) {
}

func (bv *BaseVisitor) VisitListExpr(le *ListExpr) {
}

func (bv *BaseVisitor) VisitCallExpr(ce *CallExpr) {
}

func (bv *BaseVisitor) VisitIndexExpr(ie *IndexExpr) {
}
