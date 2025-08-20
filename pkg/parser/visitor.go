package parser

import (
	"fmt"
	"reflect"
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

type BaseVisitor struct {
	ChildVisitor Visitor
}

func (bv *BaseVisitor) Analyze(p *Program) {
	bv.VisitProgram(p)
}

func (bv *BaseVisitor) ChildImplements(visitMethodName string) bool {
	st := reflect.TypeOf(bv.ChildVisitor)
	_, ok := st.MethodByName(visitMethodName)
	return ok
}

// VisitNext will call the next visit method of the visitor according to
// the type of the given operation.
// If the ChildVisitor that extends the BaseVisitor implements an "overriding"
// visit method, it will be called instead of the one defined on the BaseVisitor.
func (bv *BaseVisitor) VisitNext(operation Operation) {
	switch operation := operation.(type) {
	case *NamedType:
		if bv.ChildImplements("VisitNamedType") {
			bv.ChildVisitor.VisitNamedType(operation)
		} else {
			bv.VisitNamedType(operation)
		}
	case *ListType:
		if bv.ChildImplements("VisitListType") {
			bv.ChildVisitor.VisitListType(operation)
		} else {
			bv.VisitListType(operation)
		}
	case *FuncDef:
		if bv.ChildImplements("VisitFuncDef") {
			bv.ChildVisitor.VisitFuncDef(operation)
		} else {
			bv.VisitFuncDef(operation)
		}
	case *TypedVar:
		if bv.ChildImplements("VisitTypedVar") {
			bv.ChildVisitor.VisitTypedVar(operation)
		} else {
			bv.VisitTypedVar(operation)
		}
	case *GlobalDecl:
		if bv.ChildImplements("VisitGlobalDecl") {
			bv.ChildVisitor.VisitGlobalDecl(operation)
		} else {
			bv.VisitGlobalDecl(operation)
		}
	case *NonLocalDecl:
		if bv.ChildImplements("VisitNonLocalDecl") {
			bv.ChildVisitor.VisitNonLocalDecl(operation)
		} else {
			bv.VisitNonLocalDecl(operation)
		}
	case *VarDef:
		if bv.ChildImplements("VisitVarDef") {
			bv.ChildVisitor.VisitVarDef(operation)
		} else {
			bv.VisitVarDef(operation)
		}
	case *IfStmt:
		if bv.ChildImplements("VisitIfStmt") {
			bv.ChildVisitor.VisitIfStmt(operation)
		} else {
			bv.VisitIfStmt(operation)
		}
	case *WhileStmt:
		if bv.ChildImplements("VisitWhileStmt") {
			bv.ChildVisitor.VisitWhileStmt(operation)
		} else {
			bv.VisitWhileStmt(operation)
		}
	case *ForStmt:
		if bv.ChildImplements("VisitForStmt") {
			bv.ChildVisitor.VisitForStmt(operation)
		} else {
			bv.VisitForStmt(operation)
		}
	case *PassStmt:
		if bv.ChildImplements("VisitPassStmt") {
			bv.ChildVisitor.VisitPassStmt(operation)
		} else {
			bv.VisitPassStmt(operation)
		}
	case *ReturnStmt:
		if bv.ChildImplements("VisitReturnStmt") {
			bv.ChildVisitor.VisitReturnStmt(operation)
		} else {
			bv.VisitReturnStmt(operation)
		}
	case *AssignStmt:
		if bv.ChildImplements("VisitAssignStmt") {
			bv.ChildVisitor.VisitAssignStmt(operation)
		} else {
			bv.VisitAssignStmt(operation)
		}
	case *LiteralExpr:
		if bv.ChildImplements("VisitLiteralExpr") {
			bv.ChildVisitor.VisitLiteralExpr(operation)
		} else {
			bv.VisitLiteralExpr(operation)
		}
	case *IdentExpr:
		if bv.ChildImplements("VisitIdentExpr") {
			bv.ChildVisitor.VisitIdentExpr(operation)
		} else {
			bv.VisitIdentExpr(operation)
		}
	case *UnaryExpr:
		if bv.ChildImplements("VisitUnaryExpr") {
			bv.ChildVisitor.VisitUnaryExpr(operation)
		} else {
			bv.VisitUnaryExpr(operation)
		}
	case *BinaryExpr:
		if bv.ChildImplements("VisitBinaryExpr") {
			bv.ChildVisitor.VisitBinaryExpr(operation)
		} else {
			bv.VisitBinaryExpr(operation)
		}
	case *IfExpr:
		if bv.ChildImplements("VisitIfExpr") {
			bv.ChildVisitor.VisitIfExpr(operation)
		} else {
			bv.VisitIfExpr(operation)
		}
	case *ListExpr:
		if bv.ChildImplements("VisitListExpr") {
			bv.ChildVisitor.VisitListExpr(operation)
		} else {
			bv.VisitListExpr(operation)
		}
	case *CallExpr:
		if bv.ChildImplements("VisitCallExpr") {
			bv.ChildVisitor.VisitCallExpr(operation)
		} else {
			bv.VisitCallExpr(operation)
		}
	case *IndexExpr:
		if bv.ChildImplements("VisitIndexExpr") {
			bv.ChildVisitor.VisitIndexExpr(operation)
		} else {
			bv.VisitIndexExpr(operation)
		}
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
		bv.VisitNext(definition)
	}
	for _, statement := range p.Statements {
		bv.VisitNext(statement)
	}
}

func (bv *BaseVisitor) VisitFuncDef(fd *FuncDef) {
	fmt.Println(fd.Name())
	for _, param := range fd.Parameters {
		bv.VisitNext(param)
	}
	for _, bodyOp := range fd.FuncBody {
		bv.VisitNext(bodyOp)
	}
	bv.VisitNext(fd.ReturnType)
}

func (bv *BaseVisitor) VisitTypedVar(tv *TypedVar) {
	fmt.Println(tv.Name())
	bv.VisitNext(tv.VarType)
}

func (bv *BaseVisitor) VisitGlobalDecl(gd *GlobalDecl) {
	fmt.Println(gd.Name())
}

func (bv *BaseVisitor) VisitNonLocalDecl(nl *NonLocalDecl) {
	fmt.Println(nl.Name())
}

func (bv *BaseVisitor) VisitVarDef(vd *VarDef) {
	fmt.Println(vd.Name())
	bv.VisitNext(vd.TypedVar)
	bv.VisitNext(vd.Literal)
}

func (bv *BaseVisitor) VisitIfStmt(is *IfStmt) {
	fmt.Println(is.Name())
	bv.VisitNext(is.Condition)
	for _, ifBodyOp := range is.IfBody {
		bv.VisitNext(ifBodyOp)
	}
	for _, elseBodyOp := range is.ElseBody {
		bv.VisitNext(elseBodyOp)
	}
}

func (bv *BaseVisitor) VisitWhileStmt(ws *WhileStmt) {
	fmt.Println(ws.Name())
	bv.VisitNext(ws.Condition)
	for _, bodyOp := range ws.Body {
		bv.VisitNext(bodyOp)
	}
}

func (bv *BaseVisitor) VisitForStmt(fs *ForStmt) {
	fmt.Println(fs.Name())
	bv.VisitNext(fs.Iter)
	for _, bodyOp := range fs.Body {
		bv.VisitNext(bodyOp)
	}
}

func (bv *BaseVisitor) VisitPassStmt(ps *PassStmt) {
	fmt.Println(ps.Name())
}

func (bv *BaseVisitor) VisitReturnStmt(rs *ReturnStmt) {
	fmt.Println(rs.Name())
	bv.VisitNext(rs.ReturnVal)
}

func (bv *BaseVisitor) VisitAssignStmt(as *AssignStmt) {
	fmt.Println(as.Name())
	bv.VisitNext(as.Target)
	bv.VisitNext(as.Value)
}

func (bv *BaseVisitor) VisitLiteralExpr(le *LiteralExpr) {
	fmt.Println(le.Name())
}

func (bv *BaseVisitor) VisitIdentExpr(ie *IdentExpr) {
	fmt.Println(ie.Name())
}

func (bv *BaseVisitor) VisitUnaryExpr(ue *UnaryExpr) {
	fmt.Println(ue.Name())
	bv.VisitNext(ue.Value)
}

func (bv *BaseVisitor) VisitBinaryExpr(be *BinaryExpr) {
	fmt.Println(be.Name())
	bv.VisitNext(be.Lhs)
	bv.VisitNext(be.Rhs)
}

func (bv *BaseVisitor) VisitIfExpr(ie *IfExpr) {
	fmt.Println(ie.Name())
	bv.VisitNext(ie.Condition)
	bv.VisitNext(ie.IfOp)
	bv.VisitNext(ie.ElseOp)
}

func (bv *BaseVisitor) VisitListExpr(le *ListExpr) {
	fmt.Println(le.Name())
	for _, elem := range le.Elements {
		bv.VisitNext(elem)
	}
}

func (bv *BaseVisitor) VisitCallExpr(ce *CallExpr) {
	fmt.Println(ce.Name())
	for _, argument := range ce.Arguments {
		bv.VisitNext(argument)
	}
}

func (bv *BaseVisitor) VisitIndexExpr(ie *IndexExpr) {
	fmt.Println(ie.Name())
	bv.VisitNext(ie.Value)
	bv.VisitNext(ie.Index)
}
