package parser

/* Types */

type Operation any

type namedType struct {
	typeName string
	Operation
}

type listType struct {
	elemType Operation
	Operation
}

/* Definitions */

type astProgram struct {
	definitions []Operation
	statements  []Operation
	Operation
}

type astFuncDef struct {
	funcName   string
	parameters []Operation
	funcBody   []Operation
	returnType Operation
	Operation
}

type typedVar struct {
	varName string
	varType Operation
	Operation
}

type globalDecl struct {
	declName string
	Operation
}

type nonLocalDecl struct {
	declName string
	Operation
}

type varDef struct {
	typedVar *typedVar
	literal  Operation
	Operation
}

/* Statements */

type ifStmt struct {
	condition Operation
	ifBody    []Operation
	elseBody  []Operation
	Operation
}

type whileStmt struct {
	condition Operation
	body      []Operation
	Operation
}

type forStmt struct {
	iterName string
	iter     Operation
	body     []Operation
	Operation
}

type passStmt struct {
	Operation
}

type returnStmt struct {
	returnVal Operation
	Operation
}

type assignStmt struct {
	target Operation
	value  Operation
	Operation
}

/* Expressions */

// Uses *int so the literal can also be nil
type literalType interface {
	string | int | bool | *int
}

type literalExpr[T literalType] struct {
	// TODO: if this leads to complications just replace with any
	value T
	Operation
}

type identExpr struct {
	identifier string
	Operation
}

type unaryExpr struct {
	operation string
	value     Operation
	Operation
}

type binaryExpr struct {
	operation string
	lhs       Operation
	rhs       Operation
	Operation
}

type ifExpr struct {
	condition Operation
	ifOp      Operation
	elseOp    Operation
	Operation
}

type listExpr struct {
	elements []Operation
	Operation
}

type callExpr struct {
	funcName  string
	arguments []Operation
	Operation
}

type indexExpr struct {
	value Operation
	index Operation
	Operation
}
