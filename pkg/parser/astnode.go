package parser

/* Types */

type Operation any

type namedType struct {
	TypeName string
	Operation
}

type listType struct {
	ElemType Operation
	Operation
}

/* Definitions */

type Program struct {
	Definitions []Operation
	Statements  []Operation
	Operation
}

type funcDef struct {
	FuncName   string
	Parameters []Operation
	FuncBody   []Operation
	ReturnType Operation
	Operation
}

type typedVar struct {
	VarName string
	VarType Operation
	Operation
}

type globalDecl struct {
	DeclName string
	Operation
}

type nonLocalDecl struct {
	DeclName string
	Operation
}

type varDef struct {
	TypedVar *typedVar
	Literal  Operation
	Operation
}

/* Statements */

type ifStmt struct {
	Condition Operation
	IfBody    []Operation
	ElseBody  []Operation
	Operation
}

type whileStmt struct {
	Condition Operation
	Body      []Operation
	Operation
}

type forStmt struct {
	IterName string
	Iter     Operation
	Body     []Operation
	Operation
}

type passStmt struct {
	Operation
}

type returnStmt struct {
	ReturnVal Operation
	Operation
}

type assignStmt struct {
	Target Operation
	Value  Operation
	Operation
}

/* Expressions */

type literalExpr struct {
	Value any
	Operation
}

type identExpr struct {
	Identifier string
	Operation
}

type unaryExpr struct {
	Op    string
	Value Operation
	Operation
}

type binaryExpr struct {
	Op  string
	Lhs Operation
	Rhs Operation
	Operation
}

type ifExpr struct {
	Condition Operation
	IfOp      Operation
	ElseOp    Operation
	Operation
}

type listExpr struct {
	Elements []Operation
	Operation
}

type callExpr struct {
	FuncName  string
	Arguments []Operation
	Operation
}

type indexExpr struct {
	Value Operation
	Index Operation
	Operation
}
