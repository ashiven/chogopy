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
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func attrToType(attr ast.TypeAttr) types.Type {
	_, isListAttr := attr.(ast.ListAttribute)
	if isListAttr {
		elemType := attrToType(attr.(ast.ListAttribute).ElemType)
		return types.NewPointer(elemType)
	}

	switch attr.(ast.BasicAttribute) {
	case ast.Integer:
		return types.I32
	case ast.Boolean:
		return types.I1
	case ast.String:
		return types.I8Ptr
	case ast.None:
		return types.Void
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

type UniqueNames map[string]int

func (un UniqueNames) get(name string) string {
	if _, ok := un[name]; ok {
		un[name]++
	} else {
		un[name] = 0
	}
	return name + strconv.Itoa(un[name])
}

type VarDef struct {
	name     string
	elemType types.Type
	value    value.Value
}

type (
	FuncDefs map[string]*ir.Func
	VarDefs  map[string]VarDef
)

type CodeGenerator struct {
	Module *ir.Module

	currentFunction *ir.Func
	currentBlock    *ir.Block

	uniqueNames UniqueNames

	varDefs  VarDefs
	funcDefs FuncDefs

	lastGenerated value.Value
	ast.BaseVisitor
}

func (cg *CodeGenerator) Generate(program *ast.Program) {
	cg.Module = ir.NewModule()
	cg.uniqueNames = UniqueNames{}
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
	mainBlock := mainFunction.NewBlock(cg.uniqueNames.get("entry"))
	cg.currentFunction = mainFunction
	cg.currentBlock = mainBlock

	for _, statement := range program.Statements {
		statement.Visit(cg)
	}

	cg.currentBlock.NewRet(constant.NewInt(types.I32, 0))
}

func (cg *CodeGenerator) Traverse() bool {
	return false
}

// OptionalLoad is a convenience method that can be called on a value
// if one is unsure whether it is a pointer to something whose value we would like to use
// or if it already is an SSA Value containing that something.
// If the given value is a pointer, it will load the value at that pointer.
func (cg *CodeGenerator) OptionalLoad(val value.Value) value.Value {
	if _, ok := val.Type().(*types.PointerType); ok {
		valueLoad := cg.currentBlock.NewLoad(val.Type().(*types.PointerType).ElemType, val)
		valueLoad.LocalName = cg.uniqueNames.get("opt_val")
		return valueLoad
	}
	return val
}

func (cg *CodeGenerator) VisitNamedType(namedType *ast.NamedType) {
	/* no op */
}

func (cg *CodeGenerator) VisitListType(listType *ast.ListType) {
	/* no op */
}

func (cg *CodeGenerator) VisitProgram(program *ast.Program) {
	/* no op */
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
	newBlock := newFunction.NewBlock(cg.uniqueNames.get("entry"))

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
	/* no op */
}

func (cg *CodeGenerator) VisitGlobalDecl(globalDecl *ast.GlobalDecl) {
	/* no op */
}

func (cg *CodeGenerator) VisitNonLocalDecl(nonLocalDecl *ast.NonLocalDecl) {
	/* no op */
}

func (cg *CodeGenerator) VisitVarDef(varDef *ast.VarDef) {
	varName := varDef.TypedVar.(*ast.TypedVar).VarName
	varType := astTypeToType(varDef.TypedVar.(*ast.TypedVar).VarType)
	literalVal := varDef.Literal.(*ast.LiteralExpr).Value

	var literalConst constant.Constant
	switch varDef.Literal.(*ast.LiteralExpr).TypeHint {
	case ast.Integer:
		literalConst = constant.NewInt(types.I32, int64(literalVal.(int)))
	case ast.Boolean:
		literalConst = constant.NewBool(literalVal.(bool))
	case ast.String:
		literalConst = constant.NewCharArrayFromString(literalVal.(string))
	case ast.None:
		switch varType := varType.(type) {
		case *types.PointerType:
			literalConst = constant.NewNull(varType)
		case *types.IntType:
			literalConst = constant.NewInt(varType, int64(0))
		}
	}

	newVar := cg.Module.NewGlobalDef(varName, literalConst)
	cg.varDefs[varName] = VarDef{name: varName, elemType: newVar.Typ.ElemType, value: newVar}
}

/* Statements */

func (cg *CodeGenerator) VisitIfStmt(ifStmt *ast.IfStmt) {
	ifStmt.Condition.Visit(cg)
	condition := cg.lastGenerated

	ifBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("if"))
	elseBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("else"))
	exitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("exit"))

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
	whileCondBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("while.cond"))
	whileBodyBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("while.body"))
	whileExitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("while.exit"))

	whileStmt.Condition.Visit(cg)
	cond := cg.lastGenerated
	cg.currentBlock.NewBr(whileCondBlock)

	/* Condition block */
	cg.currentBlock = whileCondBlock
	continueLoop := cg.OptionalLoad(cond)
	cg.currentBlock.NewCondBr(continueLoop, whileBodyBlock, whileExitBlock)

	/* Body block */
	cg.currentBlock = whileBodyBlock
	for _, bodyOp := range whileStmt.Body {
		bodyOp.Visit(cg)
	}
	cg.currentBlock.NewBr(whileCondBlock)

	/* Exit block */
	cg.currentBlock = whileExitBlock
}

func (cg *CodeGenerator) VisitForStmt(forStmt *ast.ForStmt) {
	forCondBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.cond"))
	forBodyBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.body"))
	forIncBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.inc"))
	forExitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.exit"))

	// NOTE: We are using iterName to iterate over a string/list, so we should reset its value to an empty string/0 before assigning to it.
	iterName := cg.varDefs[forStmt.IterName].value
	iterNameType := cg.varDefs[forStmt.IterName].elemType

	forStmt.Iter.Visit(cg)
	iterVal := cg.lastGenerated

	var iterLength int
	switch iter := forStmt.Iter.(type) {
	case *ast.ListExpr:
		iterLength = iter.TypeHint.(ast.ListAttribute).Length
	case *ast.LiteralExpr:
		strLiteral := forStmt.Iter.(*ast.LiteralExpr).Value
		iterLength = len(strLiteral.(string))
	case *ast.IdentExpr:
		_, identIsList := iter.TypeHint.(ast.ListAttribute)
		if identIsList {
			iterLength = iter.TypeHint.(ast.ListAttribute).Length
		}
		// TODO: figure out a way to get the length for a string variable.
	}

	// Some constants for convenience
	zero := constant.NewInt(types.I32, 0)
	one := constant.NewInt(types.I32, 1)
	iterLen := constant.NewInt(types.I32, int64(iterLength))

	// Initialize iteration index
	indexAlloc := cg.currentBlock.NewAlloca(types.I32)
	indexAlloc.LocalName = cg.uniqueNames.get("index_ptr")
	cg.currentBlock.NewStore(zero, indexAlloc)
	cg.currentBlock.NewBr(forCondBlock)

	/* Condition block */
	cg.currentBlock = forCondBlock
	index := cg.currentBlock.NewLoad(types.I32, indexAlloc)
	index.LocalName = cg.uniqueNames.get("index")
	continueLoop := cg.currentBlock.NewICmp(enum.IPredSLT, index, iterLen)
	continueLoop.LocalName = cg.uniqueNames.get("continue")
	cg.currentBlock.NewCondBr(continueLoop, forBodyBlock, forExitBlock)

	/* Body block */
	cg.currentBlock = forBodyBlock
	currentAddress := cg.currentBlock.NewGetElementPtr(iterNameType, iterVal, index)
	currentAddress.LocalName = cg.uniqueNames.get("curr_addr")
	currentVal := cg.currentBlock.NewLoad(iterNameType, currentAddress)
	currentVal.LocalName = cg.uniqueNames.get("curr_val")
	cg.currentBlock.NewStore(currentVal, iterName)
	for _, bodyOp := range forStmt.Body {
		bodyOp.Visit(cg)
	}
	cg.currentBlock.NewBr(forIncBlock)

	/* Increment block */
	cg.currentBlock = forIncBlock
	incremented := cg.currentBlock.NewAdd(index, one)
	incremented.LocalName = cg.uniqueNames.get("inc")
	cg.currentBlock.NewStore(incremented, indexAlloc)
	cg.currentBlock.NewBr(forCondBlock)

	/* Exit block */
	cg.currentBlock = forExitBlock
}

func (cg *CodeGenerator) VisitPassStmt(passStmt *ast.PassStmt) {
	/* no op */
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
	assignStmt.Target.Visit(cg)
	target := cg.lastGenerated

	assignStmt.Value.Visit(cg)
	value := cg.lastGenerated

	cg.currentBlock.NewStore(value, target)
}

/* Expressions */

func (cg *CodeGenerator) VisitLiteralExpr(literalExpr *ast.LiteralExpr) {
	var literalConst constant.Constant

	switch literalVal := literalExpr.Value.(type) {
	case int:
		literalConst = constant.NewInt(types.I32, int64(literalVal))
	case bool:
		literalConst = constant.NewBool(literalVal)
	case string:
		literalConst = constant.NewCharArrayFromString(literalVal + "\x00")
	default:
		literalConst = constant.NewNull(types.I8Ptr)
	}

	literalAlloc := cg.currentBlock.NewAlloca(literalConst.Type())
	literalAlloc.LocalName = cg.uniqueNames.get("literal_ptr")
	cg.currentBlock.NewStore(literalConst, literalAlloc)

	if _, ok := literalExpr.Value.(string); ok {
		cg.lastGenerated = literalAlloc
	} else {
		literalLoad := cg.currentBlock.NewLoad(literalConst.Type(), literalAlloc)
		literalLoad.LocalName = cg.uniqueNames.get("literal_val")
		cg.lastGenerated = literalLoad
	}
}

func (cg *CodeGenerator) VisitIdentExpr(identExpr *ast.IdentExpr) {
	identName := identExpr.Identifier

	// Case 1: The identifier refers to a global var def.
	if varDef, ok := cg.varDefs[identName]; ok {
		cg.lastGenerated = varDef.value
	}

	// Case 2: The identifier refers to the name of a parameter of the current function. (overwrites global def)
	for _, param := range cg.currentFunction.Params {
		if identName == param.LocalName {
			cg.lastGenerated = param
		}
	}
}

func (cg *CodeGenerator) VisitUnaryExpr(unaryExpr *ast.UnaryExpr) {
	// TODO: implement
}

func (cg *CodeGenerator) VisitBinaryExpr(binaryExpr *ast.BinaryExpr) {
	binaryExpr.Lhs.Visit(cg)
	lhsValue := cg.lastGenerated

	binaryExpr.Rhs.Visit(cg)
	rhsValue := cg.lastGenerated

	// TODO: implement
	switch binaryExpr.Op {
	case "+":
		newAdd := cg.currentBlock.NewAdd(lhsValue, rhsValue)
		cg.lastGenerated = newAdd
	}
}

func (cg *CodeGenerator) VisitIfExpr(ifExpr *ast.IfExpr) {
	// TODO: implement
}

func (cg *CodeGenerator) VisitListExpr(listExpr *ast.ListExpr) {
	listElemType := attrToType(listExpr.TypeHint.(ast.ListAttribute).ElemType)
	listAlloc := cg.currentBlock.NewAlloca(listElemType)
	listAlloc.LocalName = cg.uniqueNames.get("list_ptr")

	for elemIdx, elem := range listExpr.Elements {
		elem.Visit(cg)
		elemVal := cg.lastGenerated

		elemIdxConst := constant.NewInt(types.I32, int64(elemIdx))
		elemPtr := cg.currentBlock.NewGetElementPtr(listElemType, listAlloc, elemIdxConst)
		elemPtr.LocalName = cg.uniqueNames.get("list_elem_ptr")

		cg.currentBlock.NewStore(elemVal, elemPtr)
	}

	cg.lastGenerated = listAlloc
}

func (cg *CodeGenerator) VisitCallExpr(callExpr *ast.CallExpr) {
	args := []value.Value{}
	for _, arg := range callExpr.Arguments {
		arg.Visit(cg)
		args = append(args, cg.lastGenerated)
	}

	switch callExpr.FuncName {
	case "print":
		for i, arg := range args {
			bitCast := cg.currentBlock.NewBitCast(arg, types.I8Ptr)
			bitCast.LocalName = cg.uniqueNames.get("print_arg_cast")
			args[i] = bitCast
		}
	}

	callee := cg.funcDefs[callExpr.FuncName]
	newCall := cg.currentBlock.NewCall(callee, args...)
	newCall.LocalName = cg.uniqueNames.get("call")
}

func (cg *CodeGenerator) VisitIndexExpr(indexExpr *ast.IndexExpr) {
	indexExpr.Value.Visit(cg)
	// We have to do this to get the value out of the global variable pointer
	// returned after visiting identExpr but it will break stuff if this is done again inside a nested indexExpr
	val := cg.OptionalLoad(cg.lastGenerated)

	indexExpr.Index.Visit(cg)
	index := cg.lastGenerated

	currentAddr := cg.currentBlock.NewGetElementPtr(val.Type().(*types.PointerType).ElemType, val, index)
	currentAddr.LocalName = cg.uniqueNames.get("index_addr")

	cg.lastGenerated = cg.OptionalLoad(currentAddr)

	// TODO: this works for simple index exprs but will go horribly
	// wrong for anything more complicated like nested index exprs
	// or indexing into a string literal etc.
}
