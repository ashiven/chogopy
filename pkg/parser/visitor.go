package parser

import "fmt"

type Visitor interface {
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

func (bv *BaseVisitor) visitNext(operation Operation) {
	switch operation := operation.(type) {
	case *NamedType:
		bv.VisitNamedType(operation)
	case *ListType:
		bv.VisitListType(operation)
	case *FuncDef:
		bv.VisitFuncDef(operation)
	case *TypedVar:
		bv.VisitTypedVar(operation)
	case *GlobalDecl:
		bv.VisitGlobalDecl(operation)
	case *NonLocalDecl:
		bv.VisitNonLocalDecl(operation)
	case *VarDef:
		bv.VisitVarDef(operation)
	case *IfStmt:
		bv.VisitIfStmt(operation)
	case *WhileStmt:
		bv.VisitWhileStmt(operation)
	case *ForStmt:
		bv.VisitForStmt(operation)
	case *PassStmt:
		bv.VisitPassStmt(operation)
	case *ReturnStmt:
		bv.VisitReturnStmt(operation)
	case *AssignStmt:
		bv.VisitAssignStmt(operation)
	case *LiteralExpr:
		bv.VisitLiteralExpr(operation)
	case *IdentExpr:
		bv.VisitIdentExpr(operation)
	case *UnaryExpr:
		bv.VisitUnaryExpr(operation)
	case *BinaryExpr:
		bv.VisitBinaryExpr(operation)
	case *IfExpr:
		bv.VisitIfExpr(operation)
	case *ListExpr:
		bv.VisitListExpr(operation)
	case *CallExpr:
		bv.VisitCallExpr(operation)
	case *IndexExpr:
		bv.VisitIndexExpr(operation)
	}
}

func (bv *BaseVisitor) VisitNamedType(nt *NamedType) {
	fmt.Println(nt.Name())
}

func (bv *BaseVisitor) VisitListType(lt *ListType) {
	fmt.Println(lt.Name())
}

func (bv *BaseVisitor) VisitProgram(p *Program) {
	for _, definition := range p.Definitions {
		bv.visitNext(definition)
	}
	for _, statement := range p.Statements {
		bv.visitNext(statement)
	}
}

func (bv *BaseVisitor) VisitFuncDef(fd *FuncDef) {
	fmt.Println(fd.Name())
	for _, param := range fd.Parameters {
		bv.visitNext(param)
	}
	for _, bodyOp := range fd.FuncBody {
		bv.visitNext(bodyOp)
	}
	bv.visitNext(fd.ReturnType)
}

func (bv *BaseVisitor) VisitTypedVar(tv *TypedVar) {
	fmt.Println(tv.Name())
	bv.visitNext(tv.VarType)
}

func (bv *BaseVisitor) VisitGlobalDecl(gd *GlobalDecl) {
	fmt.Println(gd.Name())
}

func (bv *BaseVisitor) VisitNonLocalDecl(nl *NonLocalDecl) {
	fmt.Println(nl.Name())
}

func (bv *BaseVisitor) VisitVarDef(vd *VarDef) {
	fmt.Println(vd.Name())
	bv.visitNext(vd.TypedVar)
	bv.visitNext(vd.Literal)
}

func (bv *BaseVisitor) VisitIfStmt(is *IfStmt) {
	fmt.Println(is.Name())
	bv.visitNext(is.Condition)
	for _, ifBodyOp := range is.IfBody {
		bv.visitNext(ifBodyOp)
	}
	for _, elseBodyOp := range is.ElseBody {
		bv.visitNext(elseBodyOp)
	}
}

func (bv *BaseVisitor) VisitWhileStmt(ws *WhileStmt) {
	fmt.Println(ws.Name())
	bv.visitNext(ws.Condition)
	for _, bodyOp := range ws.Body {
		bv.visitNext(bodyOp)
	}
}

func (bv *BaseVisitor) VisitForStmt(fs *ForStmt) {
	fmt.Println(fs.Name())
	bv.visitNext(fs.Iter)
	for _, bodyOp := range fs.Body {
		bv.visitNext(bodyOp)
	}
}

func (bv *BaseVisitor) VisitPassStmt(ps *PassStmt) {
	fmt.Println(ps.Name())
}

func (bv *BaseVisitor) VisitReturnStmt(rs *ReturnStmt) {
	fmt.Println(rs.Name())
	bv.visitNext(rs.ReturnVal)
}

func (bv *BaseVisitor) VisitAssignStmt(as *AssignStmt) {
	fmt.Println(as.Name())
	bv.visitNext(as.Target)
	bv.visitNext(as.Value)
}

func (bv *BaseVisitor) VisitLiteralExpr(le *LiteralExpr) {
	fmt.Println(le.Name())
}

func (bv *BaseVisitor) VisitIdentExpr(ie *IdentExpr) {
	fmt.Println(ie.Name())
}

func (bv *BaseVisitor) VisitUnaryExpr(ue *UnaryExpr) {
	fmt.Println(ue.Name())
	bv.visitNext(ue.Value)
}

func (bv *BaseVisitor) VisitBinaryExpr(be *BinaryExpr) {
	fmt.Println(be.Name())
	bv.visitNext(be.Lhs)
	bv.visitNext(be.Rhs)
}

func (bv *BaseVisitor) VisitIfExpr(ie *IfExpr) {
	fmt.Println(ie.Name())
	bv.visitNext(ie.Condition)
	bv.visitNext(ie.IfOp)
	bv.visitNext(ie.ElseOp)
}

func (bv *BaseVisitor) VisitListExpr(le *ListExpr) {
	fmt.Println(le.Name())
	for _, elem := range le.Elements {
		bv.visitNext(elem)
	}
}

func (bv *BaseVisitor) VisitCallExpr(ce *CallExpr) {
	fmt.Println(ce.Name())
	for _, argument := range ce.Arguments {
		bv.visitNext(argument)
	}
}

func (bv *BaseVisitor) VisitIndexExpr(ie *IndexExpr) {
	fmt.Println(ie.Name())
	bv.visitNext(ie.Value)
	bv.visitNext(ie.Index)
}
