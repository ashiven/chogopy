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

	// TODO: This is complete nonsense by the way.
	// If you have a loop that increases the length of a variable in
	// each iteration, this already becomes incorrect.
	// Length tracking needs to be performed at runtime using for instance strlen
	// or other similar methods to get the length of a string or a list
	// So I need to start using ArrayTypes for lists and CharArray types for strings
	// because otherwise tracking the length at runtime will be impossible.
	// It is okay to use this to track the length of string/list literals
	// but thinking you can track the length of variables statically is some nonsense
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
		ir.NewParam("", types.I8),
	)
	input := cg.Module.NewFunc(
		"input",
		types.I8Ptr,
	)
	len_ := cg.Module.NewFunc(
		"strlen",
		types.I32,
		ir.NewParam("", types.I8Ptr),
	)
	strcat := cg.Module.NewFunc(
		"strcat",
		types.I8Ptr,
		ir.NewParam("", types.I8),
		ir.NewParam("", types.I8),
	)
	scanf := cg.Module.NewFunc(
		"scanf",
		types.I32,
		ir.NewParam("", types.I8),
	)

	cg.functions["print"] = print_
	cg.functions["input"] = input
	cg.functions["len"] = len_
	cg.functions["strcat"] = strcat
	cg.functions["scanf"] = scanf

	// We use arbitrary unused pointer types for our custom types and then bitcast them to match the actually expected pointer type
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

// NewStore is a wrapper around the regular block.NewStore() that first
// checks whether the src or the target are of the types which require a typecast: object, none, empty.
// And if that is the case, it performs a typecast before adding a new store instruction.
func (cg *CodeGenerator) NewStore(src value.Value, target value.Value) {
	if !isPtrTo(target, src.Type()) {
		target = cg.currentBlock.NewBitCast(target, types.NewPointer(src.Type()))
		target.(*ir.InstBitCast).LocalName = cg.uniqueNames.get("assign_cast")
	}
	//if cg.needsTypeCast(src) {
	//	src = cg.currentBlock.NewBitCast(src, target.Type().(*types.PointerType).ElemType)
	//}

	// If src is a list or a string literal, we check for its length inside of cg.lengths
	// and then update the variable info of target via cg.variables[target.name].length = newLength
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

// NewLiteral takes any literal of type int, bool, string, or nil
// and creates a new allocation and store for that value.
// It returns an SSA value containing the value of the given literal.
// The types for this value are int: I32  str: I8*  bool: I1  nil: I8*
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
		literalConst = constant.NewNull(cg.types["none"].(*types.PointerType))
	}

	literalAlloc := cg.currentBlock.NewAlloca(literalConst.Type())
	literalAlloc.LocalName = cg.uniqueNames.get("literal_ptr")
	cg.NewStore(literalConst, literalAlloc)

	if _, ok := literal.(string); ok {
		// literalConst will be something like [4 x i8] leading to an allocation type of [4 x i8]*.
		cg.lengths[literalAlloc] = len(literal.(string))
		return literalAlloc

	} else {
		literalLoad := cg.currentBlock.NewLoad(literalConst.Type(), literalAlloc)
		literalLoad.LocalName = cg.uniqueNames.get("literal_val")
		return literalLoad
	}
}

// LoadVal is a convenience method that can be called on a value
// if one is unsure whether it is a pointer to something whose value one would like to use
// or if it already is an SSA Value containing that something.
// If the given value is a pointer, it will load the value at that pointer, otherwise it will
// simply return the given value again.
func (cg *CodeGenerator) LoadVal(val value.Value) value.Value {
	if _, ok := val.Type().(*types.PointerType); ok {
		valueLoad := cg.currentBlock.NewLoad(val.Type().(*types.PointerType).ElemType, val)
		valueLoad.LocalName = cg.uniqueNames.get("val")
		return valueLoad
	}
	return val
}
