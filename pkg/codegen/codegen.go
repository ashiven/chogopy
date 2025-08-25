// Package codegen implements methods for converting
// an AST into a flattened series of LLVM IR instructions.
package codegen

import (
	"chogopy/pkg/ast"
	"strings"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type CodeGenerator struct {
	Module *ir.Module

	currentBlock *ir.Block

	lastGenerated value.Value
	ast.BaseVisitor
}

func (cg *CodeGenerator) Analyze(program *ast.Program) {
	cg.Module = ir.NewModule()

	for _, definition := range program.Definitions {
		definition.Visit(cg)
	}
	for _, statement := range program.Statements {
		statement.Visit(cg)
	}
}

func (cg *CodeGenerator) Traverse() bool {
	return false
}

func (cg *CodeGenerator) VisitNamedType(namedType *ast.NamedType) {
}

func (cg *CodeGenerator) VisitListType(listType *ast.ListType) {
}

func (cg *CodeGenerator) VisitProgram(program *ast.Program) {
}

/* Definitions */

func (cg *CodeGenerator) VisitFuncDef(funcDef *ast.FuncDef) {
	funcName := funcDef.FuncName
	returnType := funcDef.ReturnType

	var opReturnType types.Type

	_, returnIsList := returnType.(*ast.ListType)
	if returnIsList {
		opReturnType = types.I32Ptr
	} else {
		switch returnType.(*ast.NamedType).TypeName {
		case "int":
			opReturnType = types.I32
		case "bool":
			opReturnType = types.I1
		case "str":
			opReturnType = types.I8Ptr
		case "<None>":
			opReturnType = types.Void
		}
	}

	funcDefOp := cg.Module.NewFunc(funcName, opReturnType)

	cg.currentBlock = funcDefOp.NewBlock("entry")
	for _, bodyNode := range funcDef.FuncBody {
		bodyNode.Visit(cg)
	}
}

func (cg *CodeGenerator) VisitTypedVar(typedVar *ast.TypedVar) {
}

func (cg *CodeGenerator) VisitGlobalDecl(globalDecl *ast.GlobalDecl) {
}

func (cg *CodeGenerator) VisitNonLocalDecl(nonLocalDecl *ast.NonLocalDecl) {
}

func (cg *CodeGenerator) VisitVarDef(varDef *ast.VarDef) {
	varName := varDef.TypedVar.(*ast.TypedVar).VarName

	varDef.Literal.Visit(cg)
	literalValue := cg.lastGenerated.(constant.Constant)

	cg.Module.NewGlobalDef(varName, literalValue)
}

/* Statements */

func (cg *CodeGenerator) VisitIfStmt(ifStmt *ast.IfStmt) {
}

func (cg *CodeGenerator) VisitWhileStmt(whileStmt *ast.WhileStmt) {
}

func (cg *CodeGenerator) VisitForStmt(forStmt *ast.ForStmt) {
}

func (cg *CodeGenerator) VisitPassStmt(passStmt *ast.PassStmt) {
}

func (cg *CodeGenerator) VisitReturnStmt(returnStmt *ast.ReturnStmt) {
	returnStmt.ReturnVal.Visit(cg)

	cg.currentBlock.NewRet(cg.lastGenerated)
}

func (cg *CodeGenerator) VisitAssignStmt(assignStmt *ast.AssignStmt) {
}

/* Expressions */

func (cg *CodeGenerator) VisitLiteralExpr(literalExpr *ast.LiteralExpr) {
	switch literalVal := literalExpr.Value.(type) {
	case int:
		cg.lastGenerated = constant.NewInt(types.I32, int64(literalVal))
	case bool:
		cg.lastGenerated = constant.NewBool(literalVal)
	case string:
		cg.lastGenerated = constant.NewCharArrayFromString(literalVal)
	default:
		// TODO: look into the pointer type
		cg.lastGenerated = constant.NewNull(types.NewPointer(types.I1))
		return
	}
}

func (cg *CodeGenerator) VisitIdentExpr(identExpr *ast.IdentExpr) {
}

func (cg *CodeGenerator) VisitUnaryExpr(unaryExpr *ast.UnaryExpr) {
}

func (cg *CodeGenerator) VisitBinaryExpr(binaryExpr *ast.BinaryExpr) {
	binaryExpr.Lhs.Visit(cg)
	lhsValue := cg.lastGenerated

	binaryExpr.Rhs.Visit(cg)
	rhsValue := cg.lastGenerated

	switch binaryExpr.Op {
	case "+":
		newAdd := cg.currentBlock.NewAdd(lhsValue, rhsValue)
		cg.lastGenerated = newAdd
	}
}

func (cg *CodeGenerator) VisitIfExpr(ifExpr *ast.IfExpr) {
}

func (cg *CodeGenerator) VisitListExpr(listExpr *ast.ListExpr) {
	listLength := uint64(len(listExpr.Elements))

	// TODO: Since the typehint is currently a string we have to do it like this.
	// This should be changed in the future when the type hints have been changed.
	var listType *types.ArrayType
	switch {
	case strings.Contains(listExpr.TypeHint, "Integer"):
		listType = types.NewArray(listLength, types.I32)
	case strings.Contains(listExpr.TypeHint, "Boolean"):
		listType = types.NewArray(listLength, types.I1)
	case strings.Contains(listExpr.TypeHint, "String"):
		listType = types.NewArray(listLength, types.I8Ptr)
	case strings.Contains(listExpr.TypeHint, "None"):
		listType = types.NewArray(listLength, types.NewPointer(types.I1))
	}

	listElems := []constant.Constant{}
	for _, elem := range listExpr.Elements {
		elem.Visit(cg)
		listElems = append(listElems, cg.lastGenerated.(constant.Constant))
	}

	listConst := constant.NewArray(listType, listElems...)
	listAlloc := cg.currentBlock.NewAlloca(listType)
	cg.currentBlock.NewStore(listConst, listAlloc)

	cg.lastGenerated = listAlloc
}

func (cg *CodeGenerator) VisitCallExpr(callExpr *ast.CallExpr) {
}

func (cg *CodeGenerator) VisitIndexExpr(indexExpr *ast.IndexExpr) {
}
