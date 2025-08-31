package ast

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
