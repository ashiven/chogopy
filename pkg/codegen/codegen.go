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
	_, nameExists := un[name]
	if !nameExists {
		un[name] = 0
		return name + strconv.Itoa(un[name])
	}

	un[name]++
	return name + strconv.Itoa(un[name])
}

type (
	FuncDefs map[string]*ir.Func
	VarDefs  map[string]*ir.Global
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
}

func (cg *CodeGenerator) VisitGlobalDecl(globalDecl *ast.GlobalDecl) {
}

func (cg *CodeGenerator) VisitNonLocalDecl(nonLocalDecl *ast.NonLocalDecl) {
}

func (cg *CodeGenerator) VisitVarDef(varDef *ast.VarDef) {
	varName := varDef.TypedVar.(*ast.TypedVar).VarName

	varDef.Literal.Visit(cg)
	// TODO: This fails if we have a var definition that is initialized with None
	// --> cg.lastGenerated = nil --> can't convert to constant.Constant at runtime
	literalValue := cg.lastGenerated.(constant.Constant)

	newVar := cg.Module.NewGlobalDef(varName, literalValue)
	cg.varDefs[varName] = newVar
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
}

func (cg *CodeGenerator) VisitForStmt(forStmt *ast.ForStmt) {
	forCondBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.cond"))
	forBodyBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.body"))
	forIncBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.inc"))
	forExitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.exit"))

	// The iterator is either a list literal, a string literal, or an identifier storing one of the former two
	var iterNameType types.Type
	var iterLength int
	switch iter := forStmt.Iter.(type) {
	case *ast.ListExpr:
		elemType := iter.TypeHint.(ast.ListAttribute).ElemType
		iterNameType = attrToType(elemType)
		iterLength = iter.TypeHint.(ast.ListAttribute).Length
	case *ast.LiteralExpr:
		iterNameType = types.I8
		strLiteral := forStmt.Iter.(*ast.LiteralExpr).Value
		iterLength = len(strLiteral.(string))
	case *ast.IdentExpr:
		_, identIsList := iter.TypeHint.(ast.ListAttribute)
		if identIsList {
			elemType := iter.TypeHint.(ast.ListAttribute).ElemType
			iterNameType = attrToType(elemType)
			iterLength = iter.TypeHint.(ast.ListAttribute).Length
		} else {
			iterNameType = types.I8
			// TODO: figure out a way to get the length for a variable to a string
		}
	}

	// Some constants for convenience
	zero := constant.NewInt(types.I32, 0)
	one := constant.NewInt(types.I32, 1)
	iterLen := constant.NewInt(types.I32, int64(iterLength))

	// NOTE: We are using iterName to iterate over a string/list, so we should reset its value to an empty string/0 before assigning to it.
	iterName := cg.varDefs[forStmt.IterName]
	forStmt.Iter.Visit(cg)
	// NOTE: Since the way we are currently handling visits to literals (lastGenerated is merely set to a constant)
	// iterVal will also just be a constant and not an SSA Value. What we can do about this is to either make visits to literals
	// set lastGenerated to an SSA Value via Alloc-StoreConst-Load --as one would expect-- or to let the caller handle allocation
	// and moving the constant into an SSA Value. (not sure how to go about this yet -- if visitLiteral returns an SSA Val, visitVarDef breaks because it expects a constant...)
	iterVal := cg.lastGenerated
	_, iterValIsConst := iterVal.(constant.Constant)
	if iterValIsConst {
		iterValAlloc := cg.currentBlock.NewAlloca(iterVal.Type())
		cg.currentBlock.NewStore(iterVal, iterValAlloc)
		iterVal = iterValAlloc
	}
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
	iterNameCast := cg.currentBlock.NewBitCast(iterName, types.NewPointer(currentVal.Type()))
	cg.currentBlock.NewStore(currentVal, iterNameCast)
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

	// 2) A global function definition. (Do we even support function identifiers anywhere outside of call expressions?)
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
	listAlloc.LocalName = cg.uniqueNames.get("list_ptr")
	cg.currentBlock.NewStore(listConst, listAlloc)

	cg.lastGenerated = listAlloc
}

func (cg *CodeGenerator) VisitCallExpr(callExpr *ast.CallExpr) {
	args := []value.Value{}
	for _, arg := range callExpr.Arguments {
		arg.Visit(cg)

		// We only really want to allocate when an argument is a literal/constant rather than an identifier/SSA value
		// (last visit would have been to visitLiteralExpr()). Why? Because identifier arguments and expression args
		// would be stored inside of an SSA value and we would unnecessarily create another alloc and store to move
		// that SSA value around while a constant does not have its own SSA value and therefore needs an Alloc-Store.
		_, argIsIdentifier := arg.(*ast.IdentExpr)
		if argIsIdentifier {
			args = append(args, cg.lastGenerated)
		} else {
			argAlloc := cg.currentBlock.NewAlloca(cg.lastGenerated.Type())
			cg.currentBlock.NewStore(cg.lastGenerated, argAlloc)
			args = append(args, argAlloc)
		}
	}

	switch callExpr.FuncName {
	case "print":
		for i, arg := range args {
			// Bitcast will convert an argument of a type like [10 x i8]* to an arg of type i8*
			bitCast := cg.currentBlock.NewBitCast(arg, types.I8Ptr)
			args[i] = bitCast
		}
	}

	callee := cg.funcDefs[callExpr.FuncName]
	cg.currentBlock.NewCall(callee, args...)
}

func (cg *CodeGenerator) VisitIndexExpr(indexExpr *ast.IndexExpr) {
}
