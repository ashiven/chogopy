package codegen

import (
	"chogopy/pkg/ast"
	"log"

	"github.com/kr/pretty"
	"github.com/llir/llvm/ir/constant"
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

func hasType(val value.Value, type_ types.Type) bool {
	return val.Type().Equal(type_)
}

func isPtrTo(val value.Value, type_ types.Type) bool {
	_, valIsPtr := val.Type().(*types.PointerType)
	if !valIsPtr {
		return false
	}
	return val.Type().(*types.PointerType).ElemType.Equal(type_)
}

// NewLiteral takes any literal of types int, bool, string, or nil
// and creates a new allocation and store for that value.
// If the literal is an int, bool, or nil it also loads the newly allocated
// store into an SSA Value and returns that. If it is a string,
// the pointer to the allocated store is returned instead.
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
		literalConst = constant.NewNull(types.I8Ptr)
	}

	literalAlloc := cg.currentBlock.NewAlloca(literalConst.Type())
	literalAlloc.LocalName = cg.uniqueNames.get("literal_ptr")
	cg.currentBlock.NewStore(literalConst, literalAlloc)

	if _, ok := literal.(string); ok {
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
