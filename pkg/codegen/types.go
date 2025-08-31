package codegen

import (
	"chogopy/pkg/ast"
	"log"

	"github.com/kr/pretty"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) registerTypes() {
	objType := cg.Module.NewTypeDef("object", &types.StructType{Opaque: true})
	noneType := cg.Module.NewTypeDef("none", &types.StructType{Opaque: true})
	emptyType := cg.Module.NewTypeDef("empty", &types.StructType{Opaque: true})

	listContent := cg.Module.NewTypeDef("list_content", &types.StructType{Opaque: true})
	listType := cg.Module.NewTypeDef("list",
		types.NewStruct(
			types.NewPointer(listContent),
			types.I32,
		),
	)
	// fileType := cg.Module.NewTypeDef("FILE", &types.StructType{Opaque: true})

	cg.types["object"] = objType
	cg.types["none"] = noneType
	cg.types["empty"] = emptyType
	cg.types["list"] = listType
	// cg.types["FILE"] = fileType
}

/* Type check utils */

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

// isString returns true if the value is a
// - char array: [n x i8]
// - string: i8*
// - contains a char array: [n x i8]*
// - contains a string: i8**
func isString(val value.Value) bool {
	return isCharArr(val) ||
		containsCharArr(val) ||
		hasType(val, types.I8Ptr) ||
		hasType(val, types.NewPointer(types.I8Ptr))
}

func isCharArr(val value.Value) bool {
	if _, ok := val.Type().(*types.ArrayType); ok {
		if val.Type().(*types.ArrayType).ElemType.Equal(types.I8) {
			return true
		}
	}
	return false
}

func containsCharArr(val value.Value) bool {
	if _, ok := val.Type().(*types.PointerType); ok {
		if _, ok := val.Type().(*types.PointerType).ElemType.(*types.ArrayType); ok {
			if val.Type().(*types.PointerType).ElemType.(*types.ArrayType).ElemType.Equal(types.I8) {
				return true
			}
		}
	}
	return false
}

func (cg *CodeGenerator) toString(val value.Value) value.Value {
	strCast := cg.currentBlock.NewBitCast(val, types.I8Ptr)
	strCast.LocalName = cg.uniqueNames.get("str_cast")
	return strCast
}

/* Type conversion utils */

func (cg *CodeGenerator) attrToType(attr ast.TypeAttr) types.Type {
	_, isListAttr := attr.(ast.ListAttribute)
	if isListAttr {
		elemType := cg.attrToType(attr.(ast.ListAttribute).ElemType)
		listType := cg.types["list"]
		listType.(*types.StructType).Fields[0] = types.NewPointer(elemType)
		return types.NewPointer(listType)
	}

	switch attr.(ast.BasicAttribute) {
	case ast.Integer:
		return types.I32
	case ast.Boolean:
		return types.I1
	case ast.String:
		return types.I8Ptr
	case ast.None:
		return types.NewPointer(cg.types["none"])
	case ast.Empty:
		return types.NewPointer(cg.types["empty"])
	case ast.Object:
		return types.NewPointer(cg.types["object"])
	}

	log.Fatalf("Expected type attribute but got: %# v", pretty.Formatter(attr))
	return nil
}

func (cg *CodeGenerator) astTypeToType(astType ast.Node) types.Type {
	_, isListType := astType.(*ast.ListType)
	if isListType {
		elemType := cg.astTypeToType(astType.(*ast.ListType).ElemType)
		listType := cg.types["list"]
		listType.(*types.StructType).Fields[0] = types.NewPointer(elemType)
		return types.NewPointer(listType)

		// [int]  	-->  list{content: i32*, size: i32}*
		// [str]  	-->  list{content: i8**, size: i32}*
		// [[int]]  -->  list{content: list{content: i32*, size: i32}*, size: i32}*
	}

	switch astType.(*ast.NamedType).TypeName {
	case "int":
		return types.I32
	case "str":
		return types.I8Ptr
	case "bool":
		return types.I1
	case "<None>":
		return types.NewPointer(cg.types["none"])
	case "object":
		return types.NewPointer(cg.types["object"])
	}

	log.Fatalf("Expected AST Type but got: %# v", pretty.Formatter(astType))
	return nil
}
