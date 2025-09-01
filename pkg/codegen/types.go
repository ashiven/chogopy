package codegen

import (
	"chogopy/pkg/ast"
	"log"
	"strings"

	"github.com/kr/pretty"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

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

func isList(val value.Value) bool {
	if _, ok := val.Type().(*types.PointerType); ok {
		typeName := val.Type().(*types.PointerType).ElemType.String()
		if strings.Contains(typeName, "list") {
			return true
		}
	}
	return false
}

func (cg *CodeGenerator) getContentType(list value.Value) types.Type {
	if isList(list) {
		return list.Type().(*types.PointerType).ElemType.(*types.StructType).Fields[0]
	}

	log.Fatalln("getContentType: expected value of type list*")
	return nil
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

func structEqual(s1 types.Type, s2 types.Type) bool {
	s1Struct, s1IsStruct := s1.(*types.StructType)
	s2Struct, s2IsStruct := s2.(*types.StructType)
	if !s1IsStruct || !s2IsStruct {
		return false
	}
	if len(s1Struct.Fields) != len(s2Struct.Fields) {
		return false
	}
	for i := range s1Struct.Fields {
		if !s1Struct.Fields[i].Equal(s2Struct.Fields[i]) {
			return false
		}
	}
	return true
}

/* Type conversion utils */

func (cg *CodeGenerator) getOrCreate(checkType types.Type) types.Type {
	for _, existingType := range cg.types {
		if structEqual(existingType, checkType) {
			return existingType
		}
	}
	typeName := cg.uniqueNames.get("list")
	cg.Module.NewTypeDef(typeName, checkType)
	cg.types[typeName] = checkType
	return checkType
}

func (cg *CodeGenerator) attrToType(attr ast.TypeAttr) types.Type {
	_, isListAttr := attr.(ast.ListAttribute)
	if isListAttr {
		elemType := cg.attrToType(attr.(ast.ListAttribute).ElemType)
		listType := types.NewStruct(
			types.NewPointer(elemType),
			types.I32,
			types.I1,
		)
		return types.NewPointer(cg.getOrCreate(listType))
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
		listType := types.NewStruct(
			types.NewPointer(elemType),
			types.I32,
			types.I1,
		)
		return types.NewPointer(cg.getOrCreate(listType))

		// [int]  	-->  list{content: i32*, size: i32, init: i1}*
		// [str]  	-->  list{content: i8**, size: i32, init: i1}*
		// [[int]]  -->  list{content: list{content: i32*, size: i32, init: i1}*, size: i32}*
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
