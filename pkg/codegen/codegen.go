// Package codegen implements methods for converting
// an AST into a flattened series of LLVM IR instructions.
package codegen

import (
	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

type SSAVal any

type CodeGenerator struct {
	Module        *ir.Module
	lastGenerated SSAVal
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
		// TODO:
		return
	}
}

func (cg *CodeGenerator) VisitIdentExpr(identExpr *ast.IdentExpr) {
}

func (cg *CodeGenerator) VisitUnaryExpr(unaryExpr *ast.UnaryExpr) {
}

func (cg *CodeGenerator) VisitBinaryExpr(binaryExpr *ast.BinaryExpr) {
}

func (cg *CodeGenerator) VisitIfExpr(ifExpr *ast.IfExpr) {
}

func (cg *CodeGenerator) VisitListExpr(listExpr *ast.ListExpr) {
}

func (cg *CodeGenerator) VisitCallExpr(callExpr *ast.CallExpr) {
}

func (cg *CodeGenerator) VisitIndexExpr(indexExpr *ast.IndexExpr) {
}
