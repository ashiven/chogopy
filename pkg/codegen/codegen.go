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

var MaxBufferSize = uint64(10000)

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
	Lengths   map[value.Value]int // keeps track of the length of string and list literals
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

	objType := cg.Module.NewTypeDef("object", &types.StructType{Opaque: true})
	noneType := cg.Module.NewTypeDef("none", &types.StructType{Opaque: true})
	emptyType := cg.Module.NewTypeDef("empty", &types.StructType{Opaque: true})
	// fileType := cg.Module.NewTypeDef("FILE", &types.StructType{Opaque: true})

	cg.types["object"] = objType
	cg.types["none"] = noneType
	cg.types["empty"] = emptyType
	// cg.types["FILE"] = fileType

	print_ := cg.Module.NewFunc(
		"printf",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)
	input := cg.Module.NewFunc(
		"input",
		types.I8Ptr,
	)
	len_ := cg.Module.NewFunc(
		"len",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)
	strcat := cg.Module.NewFunc(
		"strcat",
		types.I8Ptr,
		ir.NewParam("", types.I8Ptr),
		ir.NewParam("", types.I8Ptr),
	)
	scanf := cg.Module.NewFunc(
		"scanf",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)
	strcpy := cg.Module.NewFunc(
		"strcpy",
		types.I8Ptr,
		ir.NewParam("", types.I8Ptr),
		ir.NewParam("", types.I8Ptr),
	)
	strcmp := cg.Module.NewFunc(
		"strcmp",
		types.I32,
		ir.NewParam("", types.I8Ptr),
		ir.NewParam("", types.I8Ptr),
	)
	strlen := cg.Module.NewFunc(
		"strlen",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)
	//fgets := cg.Module.NewFunc(
	//	"fgets",
	//	types.I8Ptr,
	//	ir.NewParam("", types.I8Ptr),
	//	ir.NewParam("", types.I32),
	//	ir.NewParam("", types.I8Ptr),
	//)
	//fdopen := cg.Module.NewFunc(
	//	"fdopen",
	//	cg.types["FILE"],
	//	ir.NewParam("", types.I32),
	//	ir.NewParam("", types.I8Ptr),
	//)

	cg.functions["print"] = print_
	cg.functions["input"] = input
	cg.functions["len"] = len_
	cg.functions["strcat"] = strcat
	cg.functions["scanf"] = scanf
	cg.functions["strcpy"] = strcpy
	cg.functions["strcmp"] = strcmp
	cg.functions["strlen"] = strlen
	// cg.functions["fgets"] = fgets
	// cg.functions["fdopen"] = fdopen

	cg.functions["boolprint"] = cg.defineBoolPrint()
	// cg.functions["floordiv"] = cg.defineFloorDiv()

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

// NewStore is a wrapper around the regular ir.Block.NewStore() that first checks whether the src requires a typecast.
func (cg *CodeGenerator) NewStore(src value.Value, target value.Value) {
	if !isPtrTo(target, src.Type()) {
		target = cg.currentBlock.NewBitCast(target, types.NewPointer(src.Type()))
		target.(*ir.InstBitCast).LocalName = cg.uniqueNames.get("store_cast")
	}

	srcLen := 1
	if _, ok := cg.lengths[src]; ok {
		srcLen = cg.lengths[src]
	}

	targetName := target.Ident()[1:] // get rid of the @ or % in front of llvm ident names
	if _, ok := cg.variables[targetName]; ok {
		varInfo := cg.variables[targetName]
		varInfo.length = srcLen
		cg.variables[targetName] = varInfo
	}

	cg.currentBlock.NewStore(src, target)
}

// NewLiteral takes any literal of type int, bool, string, or nil and creates a new allocation and store for that value.
// It returns an SSA value containing the value of the given literal with the following types:
// int: i32   str: [n x i8]*   bool: i1   nil: %none*
func (cg *CodeGenerator) NewLiteral(literal any) value.Value {
	var literalConst constant.Constant

	switch literal := literal.(type) {
	case int:
		literalConst = constant.NewInt(types.I32, int64(literal))
	case bool:
		literalConst = constant.NewBool(literal)
	case string:
		literalConst = constant.NewCharArrayFromString(literal + "\x00")
	case nil:
		literalConst = constant.NewNull(types.NewPointer(cg.types["none"]))
	}

	literalPtr := cg.currentBlock.NewAlloca(literalConst.Type())
	literalPtr.LocalName = cg.uniqueNames.get("literal_ptr")
	cg.NewStore(literalConst, literalPtr)

	if _, ok := literal.(string); ok {
		cg.lengths[literalPtr] = len(literal.(string))
		return literalPtr

	} else {
		literalLoad := cg.currentBlock.NewLoad(literalConst.Type(), literalPtr)
		literalLoad.LocalName = cg.uniqueNames.get("literal_val")
		return literalLoad
	}
}

// LoadVal can be used to load the value out of an identifier or an index expression.
// If the given value points to a char array [n x i8]* it will simply be cast into a string i8* and returned.
// If the given value is already a string or is not of a pointer type (variables are always of a pointer type) it will simply be returned.
func (cg *CodeGenerator) LoadVal(val value.Value) value.Value {
	_, valIsPtr := val.Type().(*types.PointerType)

	switch {
	case containsCharArr(val):
		strCast := cg.currentBlock.NewBitCast(val, types.I8Ptr)
		strCast.LocalName = cg.uniqueNames.get("load_str_cast")
		return strCast

	case hasType(val, types.I8Ptr):
		return val

	// NOTE: We have to tread carefully here because list literals will also have a
	// pointer type but are still literals and not variables to be loaded.
	// Therefore, the caller should always check first to ensure that he is calling this
	// method for a value that really represents a variable to load from. (i.e. using isIdentOrIndex() )
	case valIsPtr:
		valueLoad := cg.currentBlock.NewLoad(val.Type().(*types.PointerType).ElemType, val)
		valueLoad.LocalName = cg.uniqueNames.get("val")
		return valueLoad

	default:
		return val
	}
}
