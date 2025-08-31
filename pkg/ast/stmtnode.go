package ast

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
