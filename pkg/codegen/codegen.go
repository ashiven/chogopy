// Package codegen implements methods for converting
// an AST into a flattened series of LLVM IR instructions.
package codegen

import (
	"chogopy/pkg/ast"
	"log"
	"strconv"

	"github.com/kr/pretty"
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

	log.Fatalf("Expected type attribute but got: %# v", pretty.Formatter(attr))
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

	log.Fatalf("Expected AST Type but got: %# v", pretty.Formatter(astType))
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

type (
	FuncDefs map[string]*ir.Func
	VarDefs  map[string]*ir.Global
)

type CodeGenerator struct {
	Module *ir.Module

	currentFunction *ir.Func
	currentBlock    *ir.Block

	blockNames BlockNames

	varDefs  VarDefs
	funcDefs FuncDefs

	lastGenerated value.Value
	ast.BaseVisitor
}

func (cg *CodeGenerator) Generate(program *ast.Program) {
	cg.Module = ir.NewModule()
	cg.blockNames = BlockNames{}
	cg.varDefs = VarDefs{}
	cg.funcDefs = FuncDefs{}

	/* Builtin functions: print, input, len */

	// TODO: add functions for builtin calls to print, input, and len
	// Note that since we are using puts for print, it currently only supports string literals.
	print_ := cg.Module.NewFunc(
		"puts",
		types.I32,
		ir.NewParam("", types.NewPointer(types.I8)),
	)
	cg.funcDefs["print"] = print_

	/* Definitions followed by statements in main func */

	for _, definition := range program.Definitions {
		definition.Visit(cg)
	}

	mainFunction := cg.Module.NewFunc("main", types.I32)
	mainBlock := mainFunction.NewBlock(cg.blockNames.get("entry"))
	cg.currentFunction = mainFunction
	cg.currentBlock = mainBlock

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
	params := []*ir.Param{}
	for _, paramNode := range funcDef.Parameters {
		paramName := paramNode.(*ast.TypedVar).VarName
		paramType := astTypeToType(paramNode.(*ast.TypedVar).VarType)
		param := ir.NewParam(paramName, paramType)
		params = append(params, param)
	}

	returnType := astTypeToType(funcDef.ReturnType)

	newFunction := cg.Module.NewFunc(funcDef.FuncName, returnType, params...)
	newBlock := newFunction.NewBlock(cg.blockNames.get("entry"))

	cg.funcDefs[funcDef.FuncName] = newFunction
	cg.currentFunction = newFunction
	cg.currentBlock = newBlock

	for _, bodyNode := range funcDef.FuncBody {
		bodyNode.Visit(cg)
	}

	if returnType == types.Void {
		cg.currentBlock.NewRet(nil)
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

	newVar := cg.Module.NewGlobalDef(varName, literalValue)
	cg.varDefs[varName] = newVar
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
		returnVal = nil
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
		cg.lastGenerated = constant.NewCharArrayFromString(literalVal + "\x00")
	default:
		cg.lastGenerated = nil
	}
}

func (cg *CodeGenerator) VisitIdentExpr(identExpr *ast.IdentExpr) {
	identName := identExpr.Identifier

	// The identifier can refer to one of three things:
	// (Also, locals should overwrite globals imo but I am not sure.)

	// 1) A global variable definition.
	for _, varDef := range cg.varDefs {
		if identName == varDef.GlobalName {
			cg.lastGenerated = varDef
		}
	}

	// 2) A global function definition.
	for _, funcDef := range cg.funcDefs {
		if identName == funcDef.GlobalName {
			cg.lastGenerated = funcDef
		}
	}

	// 3) The name of a parameter of the current function.
	for _, param := range cg.currentFunction.Params {
		if identName == param.LocalName {
			cg.lastGenerated = param
		}
	}
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
	args := []value.Value{}
	for _, arg := range callExpr.Arguments {
		arg.Visit(cg)

		// TODO: this isn't right. maybe generate allocs inside of visit literal without
		// breaking other stuff or do entirely without allocs.
		argAlloc := cg.currentBlock.NewAlloca(cg.lastGenerated.Type())
		cg.currentBlock.NewStore(cg.lastGenerated, argAlloc)
		args = append(args, argAlloc)
	}

	switch callExpr.FuncName {
	case "print":
		for _, arg := range args {
			// Bitcast will convert an argument of a type like [10 x i8] to an arg of type i8*
			bitCast := cg.currentBlock.NewBitCast(arg, types.I8Ptr)
			args[0] = bitCast
		}
	}

	callee := cg.funcDefs[callExpr.FuncName]

	cg.currentBlock.NewCall(callee, args...)
}

func (cg *CodeGenerator) VisitIndexExpr(indexExpr *ast.IndexExpr) {
}
