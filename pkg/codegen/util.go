package codegen

import (
	"chogopy/pkg/ast"
	"log"

	"github.com/kr/pretty"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func isIdentOrIndex(astNode ast.Node) bool {
	switch astNode.(type) {
	case *ast.IdentExpr:
		return true
	case *ast.IndexExpr:
		return true
	}
	return false
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

//func (cg *CodeGenerator) needsTypeCast(val value.Value) bool {
//	for _, type_ := range cg.types {
//		if hasType(val, type_) || isPtrTo(val, type_) {
//			return true
//		}
//	}
//	return false
//}

func (cg *CodeGenerator) attrToType(attr ast.TypeAttr) types.Type {
	_, isListAttr := attr.(ast.ListAttribute)
	if isListAttr {
		elemType := cg.attrToType(attr.(ast.ListAttribute).ElemType)
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
		return types.NewPointer(cg.types["none"])
	case "object":
		return types.NewPointer(cg.types["object"])
	}

	log.Fatalf("Expected AST Type but got: %# v", pretty.Formatter(astType))
	return nil
}
