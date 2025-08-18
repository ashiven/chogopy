package parser

/* Types */

type Operation interface {
	Name() string
	// TODO: would be good to have but is the effort for validation even beneficial
	// or will invalid ops not already be prevented in the parser
	Validate() bool
}

type NamedType struct {
	name     string
	TypeName string
	Operation
}

func (nt *NamedType) Name() string {
	if nt.name == "" {
		nt.name = "NamedType"
	}
	return nt.name
}

type ListType struct {
	name     string
	ElemType Operation
	Operation
}

func (lt *ListType) Name() string {
	if lt.name == "" {
		lt.name = "ListType"
	}
	return lt.name
}

/* Definitions */

type Program struct {
	name        string
	Definitions []Operation
	Statements  []Operation
	Operation
}

func (p *Program) Name() string {
	if p.name == "" {
		p.name = "Program"
	}
	return p.name
}

type FuncDef struct {
	name       string
	FuncName   string
	Parameters []Operation
	FuncBody   []Operation
	ReturnType Operation
	Operation
}

func (fd *FuncDef) Name() string {
	if fd.name == "" {
		fd.name = "FuncDef"
	}
	return fd.name
}

type TypedVar struct {
	name    string
	VarName string
	VarType Operation
	Operation
}

func (tv *TypedVar) Name() string {
	if tv.name == "" {
		tv.name = "TypedVar"
	}
	return tv.name
}

type GlobalDecl struct {
	name     string
	DeclName string
	Operation
}

func (gd *GlobalDecl) Name() string {
	if gd.name == "" {
		gd.name = "GlobalDecl"
	}
	return gd.name
}

type NonLocalDecl struct {
	name     string
	DeclName string
	Operation
}

func (nl *NonLocalDecl) Name() string {
	if nl.name == "" {
		nl.name = "NonLocalDecl"
	}
	return nl.name
}

type VarDef struct {
	name     string
	TypedVar *TypedVar
	Literal  Operation
	Operation
}

func (vd *VarDef) Name() string {
	if vd.name == "" {
		vd.name = "VarDef"
	}
	return vd.name
}

/* Statements */

type IfStmt struct {
	name      string
	Condition Operation
	IfBody    []Operation
	ElseBody  []Operation
	Operation
}

func (is *IfStmt) Name() string {
	if is.name == "" {
		is.name = "IfStmt"
	}
	return is.name
}

type WhileStmt struct {
	name      string
	Condition Operation
	Body      []Operation
	Operation
}

func (ws *WhileStmt) Name() string {
	if ws.name == "" {
		ws.name = "WhileStmt"
	}
	return ws.name
}

type ForStmt struct {
	name     string
	IterName string
	Iter     Operation
	Body     []Operation
	Operation
}

func (fs *ForStmt) Name() string {
	if fs.name == "" {
		fs.name = "ForStmt"
	}
	return fs.name
}

type PassStmt struct {
	name string
	Operation
}

func (ps *PassStmt) Name() string {
	if ps.name == "" {
		ps.name = "PassStmt"
	}
	return ps.name
}

type ReturnStmt struct {
	name      string
	ReturnVal Operation
	Operation
}

func (rs *ReturnStmt) Name() string {
	if rs.name == "" {
		rs.name = "ReturnStmt"
	}
	return rs.name
}

type AssignStmt struct {
	name   string
	Target Operation
	Value  Operation
	Operation
}

func (as *AssignStmt) Name() string {
	if as.name == "" {
		as.name = "AssignStmt"
	}
	return as.name
}

/* Expressions */

type LiteralExpr struct {
	name  string
	Value any
	Operation
}

func (le *LiteralExpr) Name() string {
	if le.name == "" {
		le.name = "LiteralExpr"
	}
	return le.name
}

type IdentExpr struct {
	name       string
	Identifier string
	Operation
}

func (ie *IdentExpr) Name() string {
	if ie.name == "" {
		ie.name = "IdentExpr"
	}
	return ie.name
}

type UnaryExpr struct {
	name  string
	Op    string
	Value Operation
	Operation
}

func (ue *UnaryExpr) Name() string {
	if ue.name == "" {
		ue.name = "UnaryExpr"
	}
	return ue.name
}

type BinaryExpr struct {
	name string
	Op   string
	Lhs  Operation
	Rhs  Operation
	Operation
}

func (be *BinaryExpr) Name() string {
	if be.name == "" {
		be.name = "BinaryExpr"
	}
	return be.name
}

type IfExpr struct {
	name      string
	Condition Operation
	IfOp      Operation
	ElseOp    Operation
	Operation
}

func (ie *IfExpr) Name() string {
	if ie.name == "" {
		ie.name = "IfExpr"
	}
	return ie.name
}

type ListExpr struct {
	name     string
	Elements []Operation
	Operation
}

func (le *ListExpr) Name() string {
	if le.name == "" {
		le.name = "ListExpr"
	}
	return le.name
}

type CallExpr struct {
	name      string
	FuncName  string
	Arguments []Operation
	Operation
}

func (ce *CallExpr) Name() string {
	if ce.name == "" {
		ce.name = "CallExpr"
	}
	return ce.name
}

type IndexExpr struct {
	name  string
	Value Operation
	Index Operation
	Operation
}

func (ie *IndexExpr) Name() string {
	if ie.name == "" {
		ie.name = "IndexExpr"
	}
	return ie.name
}
