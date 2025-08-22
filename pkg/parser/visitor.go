package parser

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
}

func (bv *BaseVisitor) VisitListType(lt *ListType) {
	lt.ElemType.Visit(bv)
}

func (bv *BaseVisitor) VisitProgram(p *Program) {
	for _, definition := range p.Definitions {
		definition.Visit(bv)
	}
	for _, statement := range p.Statements {
		statement.Visit(bv)
	}
}

func (bv *BaseVisitor) VisitFuncDef(fd *FuncDef) {
	for _, param := range fd.Parameters {
		param.Visit(bv)
	}
	for _, bodyOp := range fd.FuncBody {
		bodyOp.Visit(bv)
	}
	fd.ReturnType.Visit(bv)
}

func (bv *BaseVisitor) VisitTypedVar(tv *TypedVar) {
	tv.VarType.Visit(bv)
}

func (bv *BaseVisitor) VisitGlobalDecl(gd *GlobalDecl) {
}

func (bv *BaseVisitor) VisitNonLocalDecl(nl *NonLocalDecl) {
}

func (bv *BaseVisitor) VisitVarDef(vd *VarDef) {
	vd.TypedVar.Visit(bv)
	vd.Literal.Visit(bv)
}

func (bv *BaseVisitor) VisitIfStmt(is *IfStmt) {
	is.Condition.Visit(bv)
	for _, ifBodyOp := range is.IfBody {
		ifBodyOp.Visit(bv)
	}
	for _, elseBodyOp := range is.ElseBody {
		elseBodyOp.Visit(bv)
	}
}

func (bv *BaseVisitor) VisitWhileStmt(ws *WhileStmt) {
	for _, bodyOp := range ws.Body {
		bodyOp.Visit(bv)
	}
}

func (bv *BaseVisitor) VisitForStmt(fs *ForStmt) {
	fs.Iter.Visit(bv)
	for _, bodyOp := range fs.Body {
		bodyOp.Visit(bv)
	}
}

func (bv *BaseVisitor) VisitPassStmt(ps *PassStmt) {
}

func (bv *BaseVisitor) VisitReturnStmt(rs *ReturnStmt) {
	rs.ReturnVal.Visit(bv)
}

func (bv *BaseVisitor) VisitAssignStmt(as *AssignStmt) {
	as.Target.Visit(bv)
	as.Value.Visit(bv)
}

func (bv *BaseVisitor) VisitLiteralExpr(le *LiteralExpr) {
}

func (bv *BaseVisitor) VisitIdentExpr(ie *IdentExpr) {
}

func (bv *BaseVisitor) VisitUnaryExpr(ue *UnaryExpr) {
	ue.Value.Visit(bv)
}

func (bv *BaseVisitor) VisitBinaryExpr(be *BinaryExpr) {
	be.Lhs.Visit(bv)
	be.Rhs.Visit(bv)
}

func (bv *BaseVisitor) VisitIfExpr(ie *IfExpr) {
	ie.Condition.Visit(bv)
	ie.IfOp.Visit(bv)
	ie.ElseOp.Visit(bv)
}

func (bv *BaseVisitor) VisitListExpr(le *ListExpr) {
	for _, elem := range le.Elements {
		elem.Visit(bv)
	}
}

func (bv *BaseVisitor) VisitCallExpr(ce *CallExpr) {
	for _, argument := range ce.Arguments {
		argument.Visit(bv)
	}
}

func (bv *BaseVisitor) VisitIndexExpr(ie *IndexExpr) {
	ie.Value.Visit(bv)
	ie.Index.Visit(bv)
}
