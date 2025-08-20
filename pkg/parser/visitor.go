package parser

import (
	"fmt"
)

type Visitor interface {
	Analyze(p *Program)
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

func (bv *BaseVisitor) Analyze(p *Program) {
	p.Visit(bv)
}

func (bv *BaseVisitor) VisitNamedType(nt *NamedType) {
	fmt.Println(nt.Name())
}

func (bv *BaseVisitor) VisitListType(lt *ListType) {
	fmt.Println(lt.Name())
}

func (bv *BaseVisitor) VisitProgram(p *Program) {
	fmt.Println(p.Name())
}

func (bv *BaseVisitor) VisitFuncDef(fd *FuncDef) {
	fmt.Println(fd.Name())
}

func (bv *BaseVisitor) VisitTypedVar(tv *TypedVar) {
	fmt.Println(tv.Name())
}

func (bv *BaseVisitor) VisitGlobalDecl(gd *GlobalDecl) {
	fmt.Println(gd.Name())
}

func (bv *BaseVisitor) VisitNonLocalDecl(nl *NonLocalDecl) {
	fmt.Println(nl.Name())
}

func (bv *BaseVisitor) VisitVarDef(vd *VarDef) {
	fmt.Println(vd.Name())
}

func (bv *BaseVisitor) VisitIfStmt(is *IfStmt) {
	fmt.Println(is.Name())
}

func (bv *BaseVisitor) VisitWhileStmt(ws *WhileStmt) {
	fmt.Println(ws.Name())
}

func (bv *BaseVisitor) VisitForStmt(fs *ForStmt) {
	fmt.Println(fs.Name())
}

func (bv *BaseVisitor) VisitPassStmt(ps *PassStmt) {
	fmt.Println(ps.Name())
}

func (bv *BaseVisitor) VisitReturnStmt(rs *ReturnStmt) {
	fmt.Println(rs.Name())
}

func (bv *BaseVisitor) VisitAssignStmt(as *AssignStmt) {
	fmt.Println(as.Name())
}

func (bv *BaseVisitor) VisitLiteralExpr(le *LiteralExpr) {
	fmt.Println(le.Name())
}

func (bv *BaseVisitor) VisitIdentExpr(ie *IdentExpr) {
	fmt.Println(ie.Name())
}

func (bv *BaseVisitor) VisitUnaryExpr(ue *UnaryExpr) {
	fmt.Println(ue.Name())
}

func (bv *BaseVisitor) VisitBinaryExpr(be *BinaryExpr) {
	fmt.Println(be.Name())
}

func (bv *BaseVisitor) VisitIfExpr(ie *IfExpr) {
	fmt.Println(ie.Name())
}

func (bv *BaseVisitor) VisitListExpr(le *ListExpr) {
	fmt.Println(le.Name())
}

func (bv *BaseVisitor) VisitCallExpr(ce *CallExpr) {
	fmt.Println(ce.Name())
}

func (bv *BaseVisitor) VisitIndexExpr(ie *IndexExpr) {
	fmt.Println(ie.Name())
}
