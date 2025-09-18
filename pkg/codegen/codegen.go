// Package codegen implements methods for converting
// an AST into a flattened series of LLVM IR instructions.
package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"chogopy/pkg/ast"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type UniqueNames map[string]int

func (un UniqueNames) get(name string) string {
	if _, ok := un[name]; ok {
		un[name]++
	} else {
		un[name] = 0
	}
	return name + strconv.Itoa(un[name])
}

type VarInfo struct {
	name     string
	elemType types.Type
	value    value.Value
	init     constant.Constant
}

type (
	Strings   map[string]*ir.Global
	Functions map[string]*ir.Func
	VarCtx    map[*ir.Func]Variables
	Variables map[string]VarInfo
	Types     map[string]types.Type
)

type CodeGenerator struct {
	Module      *ir.Module
	uniqueNames UniqueNames

	types     Types
	strings   Strings
	functions Functions

	varContext VarCtx
	heapAllocs []value.Value

	mainFunction *ir.Func
	mainBlock    *ir.Block

	currentFunction *ir.Func
	currentBlock    *ir.Block

	lastGenerated value.Value
	ast.BaseVisitor
}

func (cg *CodeGenerator) Generate(program *ast.Program) {
	typeEnvBuilder := TypeEnvBuilder{}
	typeEnvBuilder.Build(program)

	cg.Module = typeEnvBuilder.Module
	cg.uniqueNames = typeEnvBuilder.uniqueNames
	cg.types = typeEnvBuilder.types

	cg.strings = Strings{}
	cg.functions = Functions{}
	cg.registerFuncs()

	cg.varContext = VarCtx{}
	cg.heapAllocs = []value.Value{}

	cg.mainFunction = cg.Module.NewFunc("main", types.I32)
	cg.mainBlock = cg.mainFunction.NewBlock(cg.uniqueNames.get("entry"))

	cg.currentFunction = cg.mainFunction
	cg.currentBlock = cg.mainBlock

	for _, definition := range program.Definitions {
		if _, ok := definition.(*ast.FuncDef); ok {
			definition.Visit(cg)
			cg.currentFunction = cg.mainFunction
			cg.currentBlock = cg.mainBlock
		} else {
			definition.Visit(cg)
		}
	}

	for _, statement := range program.Statements {
		statement.Visit(cg)
	}

	// NOTE: Freeing the heap like this is kind of redudant because
	// any memory allocated by the program will automatically be deallocated
	// once the program terminates (at the end of the main function)

	// Add a free() for each call to malloc() at the end of the main function
	// cg.freeHeap()

	cg.currentBlock.NewRet(constant.NewInt(types.I32, 0))
}

func (cg *CodeGenerator) Traverse() bool {
	return false
}

func (cg *CodeGenerator) varCtx(global bool) Variables {
	if global {
		if _, ok := cg.varContext[cg.mainFunction]; !ok {
			cg.varContext[cg.mainFunction] = Variables{}
		}
		return cg.varContext[cg.mainFunction]
	}

	if _, ok := cg.varContext[cg.currentFunction]; !ok {
		cg.varContext[cg.currentFunction] = Variables{}
	}
	return cg.varContext[cg.currentFunction]
}

func (cg *CodeGenerator) getVar(name string) (VarInfo, error) {
	switch cg.currentFunction {
	case cg.mainFunction:
		globalVars := cg.varCtx(true)
		if _, ok := globalVars[name]; ok {
			return globalVars[name], nil
		}

	default:
		// If we are not in the main function but rather inside
		// of a local scope, we want to first check whether the variable is defined locally
		// and only if it isn't we want to check the global context for the variable.
		localVars := cg.varCtx(false)
		if _, ok := localVars[name]; ok {
			return localVars[name], nil
		}

		globalVars := cg.varCtx(true)
		if _, ok := globalVars[name]; ok {
			return globalVars[name], nil
		}
	}
	return VarInfo{}, fmt.Errorf("failed to find variable: %s", name)
}

func (cg *CodeGenerator) setVar(varInfo VarInfo) {
	if cg.currentFunction == cg.mainFunction {
		globalVars := cg.varCtx(true)
		globalVars[varInfo.name] = varInfo
	} else {
		localVars := cg.varCtx(false)
		localVars[varInfo.name] = varInfo
	}
}

func (cg *CodeGenerator) freeHeap() {
	for _, ptr := range cg.heapAllocs {
		// First, we need to check whether the pointer SSA value is even defined in the current scope.
		// This is relevant when a pointer is heap-allocated and then returned from a function,
		// which results in the pointer SSA value being shadowed by the function return SSA value.
		inScope := false
		for _, inst := range cg.currentBlock.Insts {
			if strings.Contains(inst.LLString(), ptr.Ident()) {
				inScope = true
				break
			}
		}

		if inScope {
			cg.currentBlock.NewCall(cg.functions["free"], ptr)
		}
	}
}
