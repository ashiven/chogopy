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
	length   int
}

type (
	Functions map[string]*ir.Func
	Variables map[string]VarInfo
	Types     map[string]types.Type
	Lengths   map[value.Value]int // keeps track of the length of string and list values
)

type CodeGenerator struct {
	Module *ir.Module

	currentFunction *ir.Func
	currentBlock    *ir.Block

	uniqueNames UniqueNames

	variables Variables
	functions Functions
	types     Types

	lengths Lengths

	lastGenerated value.Value
	ast.BaseVisitor
}

func (cg *CodeGenerator) Generate(program *ast.Program) {
	cg.Module = ir.NewModule()
	cg.uniqueNames = UniqueNames{}
	cg.variables = Variables{}
	cg.functions = Functions{}
	cg.types = Types{}
	cg.lengths = Lengths{}

	print_ := cg.Module.NewFunc(
		"printf",
		types.I32,
		ir.NewParam("", types.NewPointer(types.I8)),
	)
	input := cg.Module.NewFunc(
		"scanf",
		types.I32,
	)
	len_ := cg.Module.NewFunc(
		"strlen",
		types.I32,
		ir.NewParam("", types.NewPointer(types.I8)),
	)

	cg.functions["print"] = print_
	cg.functions["input"] = input
	cg.functions["len"] = len_

	// We use arbitrary unused pointer types and then bitcasting them to match the actually expected pointer type
	objType := cg.Module.NewTypeDef("object", types.I16Ptr)
	noneType := cg.Module.NewTypeDef("none", types.I64Ptr)
	emptyType := cg.Module.NewTypeDef("empty", types.I128Ptr)

	cg.types["object"] = objType
	cg.types["none"] = noneType
	cg.types["empty"] = emptyType

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
