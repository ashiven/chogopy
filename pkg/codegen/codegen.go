// Package codegen implements methods for converting
// an AST into a flattened series of LLVM IR instructions.
package codegen

import (
	"chogopy/pkg/ast"
	"strconv"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func attrToType(attr ast.TypeAttr) types.Type {
	_, isListAttr := attr.(ast.ListAttribute)
	if isListAttr {
		elemType := attrToType(attr.(ast.ListAttribute).ElemType)
		listLength := attr.(ast.ListAttribute).Length
		return types.NewArray(uint64(listLength), elemType)
	}

	switch attr.(ast.BasicAttribute) {
	case ast.Integer:
		return types.I32
	case ast.Boolean:
		return types.I1
	case ast.String:
		return types.I8Ptr
	case ast.None:
		return types.NewPointer(types.I1)
	case ast.Empty:
		// TODO:
	case ast.Object:
		// TODO:
	}

	// TODO: error
	return nil
}

func astTypeToType(astType any) types.Type {
	_, isListType := astType.(*ast.ListType)
	if isListType {
		elemType := astTypeToType(astType.(*ast.ListType).ElemType)
		return types.NewPointer(elemType)
	}

	switch astType.(*ast.NamedType).TypeName {
	case "int":
		return types.I32
	case "str":
		return types.I8Ptr
	case "bool":
		return types.I1
	case "<None>":
		return types.Void
	case "object":
		// TODO: support object type somehow?
		return types.Void
	}

	// TODO: error
	return nil
}

type BlockNames map[string]int

func (bn BlockNames) get(name string) string {
	_, nameExists := bn[name]
	if !nameExists {
		bn[name] = 0
		return name + strconv.Itoa(bn[name])
	}

	bn[name]++
	return name + strconv.Itoa(bn[name])
}

type CodeGenerator struct {
	Module *ir.Module

	currentFunction *ir.Func
	currentBlock    *ir.Block

	blockNames BlockNames

	lastGenerated value.Value
	ast.BaseVisitor
}

// TODO: Analyze sounds wrong.
// Maybe remove this method from the visitor interface
// and give it whatever name fits the case.

func (cg *CodeGenerator) Analyze(program *ast.Program) {
	cg.Module = ir.NewModule()

	cg.blockNames = BlockNames{}

	// TODO: add functions for builtin calls to print, input, and len
	print_ := cg.Module.NewFunc(
		"puts",
		types.I32,
		ir.NewParam("", types.NewPointer(types.I8)),
	)
	_ = print_

	for _, definition := range program.Definitions {
		definition.Visit(cg)
	}

	mainFunction := cg.Module.NewFunc("main", types.I32)
	mainBlock := mainFunction.NewBlock(cg.blockNames.get("entry"))

	cg.currentFunction = mainFunction
	cg.currentBlock = mainBlock

	// TODO: It would probably be safe to just put all of the statements into a main function
	// since the definitions will just sit at the llvm module level.
	for _, statement := range program.Statements {
		statement.Visit(cg)
	}

	mainBlock.NewRet(constant.NewInt(types.I32, 0))
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
	returnType := astTypeToType(funcDef.ReturnType)
	newFunction := cg.Module.NewFunc(funcDef.FuncName, returnType)
	newBlock := newFunction.NewBlock(cg.blockNames.get("entry"))

	cg.currentFunction = newFunction
	cg.currentBlock = newBlock
	for _, bodyNode := range funcDef.FuncBody {
		bodyNode.Visit(cg)
	}

	if returnType == types.Void {
		// TODO: look into null pointer type
		returnVal := constant.NewNull(types.NewPointer(types.I1))
		cg.currentBlock.NewRet(returnVal)
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
	ifStmt.Condition.Visit(cg)
	condition := cg.lastGenerated

	ifBlock := cg.currentFunction.NewBlock(cg.blockNames.get("if"))
	elseBlock := cg.currentFunction.NewBlock(cg.blockNames.get("else"))
	exitBlock := cg.currentFunction.NewBlock(cg.blockNames.get("exit"))

	cg.currentBlock.NewCondBr(condition, ifBlock, elseBlock)

	cg.currentBlock = ifBlock
	for _, ifBodyNode := range ifStmt.IfBody {
		ifBodyNode.Visit(cg)
	}
	cg.currentBlock.NewBr(exitBlock)

	cg.currentBlock = elseBlock
	for _, elseBodyNode := range ifStmt.ElseBody {
		elseBodyNode.Visit(cg)
	}
	cg.currentBlock.NewBr(exitBlock)

	cg.currentBlock = exitBlock
}

func (cg *CodeGenerator) VisitWhileStmt(whileStmt *ast.WhileStmt) {
}

func (cg *CodeGenerator) VisitForStmt(forStmt *ast.ForStmt) {
}

func (cg *CodeGenerator) VisitPassStmt(passStmt *ast.PassStmt) {
}

func (cg *CodeGenerator) VisitReturnStmt(returnStmt *ast.ReturnStmt) {
	var returnVal value.Value
	if returnStmt.ReturnVal != nil {
		returnStmt.ReturnVal.Visit(cg)
		returnVal = cg.lastGenerated
	} else {
		// TODO: look into null pointer type
		returnVal = constant.NewNull(types.NewPointer(types.I1))
	}

	cg.currentBlock.NewRet(returnVal)
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
	listType := attrToType(listExpr.TypeHint).(*types.ArrayType)

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
