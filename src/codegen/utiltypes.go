package codegen

import (
	"chogopy/src/ast"
	"log"

	"github.com/kr/pretty"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

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

func (cg *CodeGenerator) sizeof(type_ types.Type, multiplier value.Value) value.Value {
	typeSize := cg.currentBlock.NewGetElementPtr(type_, constant.NewNull(types.NewPointer(type_)), constant.NewInt(types.I32, 1))
	typeSize.LocalName = cg.uniqueNames.get("type_size_ptr")
	typeSizeInt := cg.currentBlock.NewPtrToInt(typeSize, types.I32)
	typeSizeInt.LocalName = cg.uniqueNames.get("type_size_int")
	typeSizeMult := cg.currentBlock.NewMul(typeSizeInt, multiplier)
	typeSizeMult.LocalName = cg.uniqueNames.get("type_size_mul")
	return typeSizeMult
}

func (cg *CodeGenerator) getListType(checkType types.Type) types.Type {
	for _, existingType := range cg.types {
		if structEqual(existingType, checkType) {
			return existingType
		}
	}

	log.Fatalf("getListType: Expected list type but found: %# v\n", pretty.Formatter(checkType))
	return nil
}

func (cg *CodeGenerator) attrToType(attr ast.TypeAttr) types.Type {
	if _, ok := attr.(ast.ListAttribute); ok {
		elemType := cg.attrToType(attr.(ast.ListAttribute).ElemType)
		listType := types.NewStruct(types.NewPointer(elemType), types.I32, types.I1)
		return types.NewPointer(cg.getListType(listType))
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
	if _, ok := astType.(*ast.ListType); ok {
		elemType := cg.astTypeToType(astType.(*ast.ListType).ElemType)
		listType := types.NewStruct(types.NewPointer(elemType), types.I32, types.I1)
		return types.NewPointer(cg.getListType(listType))

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

	log.Fatalf("Expected AST type but got: %# v", pretty.Formatter(astType))
	return nil
}

func isIdentOrIndex(astNode ast.Node) bool {
	switch astNode.(type) {
	case *ast.IdentExpr:
		return true
	case *ast.IndexExpr:
		return true
	}
	return false
}
