package codegen

import (
	"chogopy/src/ast"
	"log"

	"github.com/kr/pretty"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

// TypeEnvBuilder performs an AST traversal of the given program
// right before the code generation stage in order to determine
// and register all the different types that appear in the program.
//
// This is useful mainly because of the way in which the list type is implemented.
// List types are dynamically created as they are encountered and can take many different forms.
//
// A program may use lists of type [i32] or it may use lists of type [[str]] or [[[i8]]].
// We need to be able to provide certain methods for lists of all of these types and since I
// haven't been able to find out how to create generic function signatures in llvm, I decided
// that I will just have to create methods for every possible list type appearing in a given program.
//
// For instance, I am now able to create a function for retrieving the element at a list index for
// lists of type [i32] and for lists of type [str] with the respective function signatures:
//
//	([i32], i32) -> i32
//	([str], i32) -> str
type TypeEnvBuilder struct {
	Module *ir.Module

	types Types

	uniqueNames UniqueNames

	ast.BaseVisitor
}

func (tb *TypeEnvBuilder) Build(program *ast.Program) {
	tb.Module = ir.NewModule()
	tb.types = Types{}
	tb.uniqueNames = UniqueNames{}

	objType := tb.Module.NewTypeDef("object", &types.StructType{Opaque: true})
	noneType := tb.Module.NewTypeDef("none", &types.StructType{Opaque: true})
	emptyType := tb.Module.NewTypeDef("empty", &types.StructType{Opaque: true})
	listContent := tb.Module.NewTypeDef("list_content", &types.StructType{Opaque: true})
	listType := tb.Module.NewTypeDef("list",
		types.NewStruct(
			types.NewPointer(listContent),
			types.I32,
			types.I1,
		),
	)
	// fileType := cg.Module.NewTypeDef("FILE", &types.StructType{Opaque: true})-

	tb.types["object"] = objType
	tb.types["none"] = noneType
	tb.types["empty"] = emptyType
	tb.types["list_content"] = listContent
	tb.types["list"] = listType
	// cg.types["FILE"] = fileType

	program.Visit(tb)
}

func (tb *TypeEnvBuilder) VisitVarDef(varDef *ast.VarDef) {
	tb.astTypeToType(varDef.TypedVar.(*ast.TypedVar).VarType)
}

func (tb *TypeEnvBuilder) VisitFuncDef(funcDef *ast.FuncDef) {
	for _, paramNode := range funcDef.Parameters {
		tb.astTypeToType(paramNode.(*ast.TypedVar).VarType)
	}
	tb.astTypeToType(funcDef.ReturnType)
}

// TODO: binaryExpr concatenation on lists may generate another type
// for instance: [i32] + [str] -> [object]
// but I wouldn't know how to even represent such a list of mixed types
// here and so I should either make mixed lists illegal somehow or
// change my entire internal list representation

func (tb *TypeEnvBuilder) VisitListExpr(listExpr *ast.ListExpr) {
	tb.attrToType(listExpr.TypeHint)
	tb.attrToType(listExpr.TypeHint.(ast.ListAttribute).ElemType)
}

func (tb *TypeEnvBuilder) VisitIfExpr(ifExpr *ast.IfExpr) {
	tb.attrToType(ifExpr.TypeHint)
}

func (tb *TypeEnvBuilder) getOrCreate(checkType types.Type) types.Type {
	for _, existingType := range tb.types {
		if structEqual(existingType, checkType) {
			return existingType
		}
	}
	typeName := tb.uniqueNames.get("list")
	tb.Module.NewTypeDef(typeName, checkType)
	tb.types[typeName] = checkType
	return checkType
}

func (tb *TypeEnvBuilder) attrToType(attr ast.TypeAttr) types.Type {
	_, isListAttr := attr.(ast.ListAttribute)
	if isListAttr {
		elemType := tb.attrToType(attr.(ast.ListAttribute).ElemType)
		listType := types.NewStruct(
			types.NewPointer(elemType),
			types.I32,
			types.I1,
		)
		return types.NewPointer(tb.getOrCreate(listType))
	}

	switch attr.(ast.BasicAttribute) {
	case ast.Integer:
		return types.I32
	case ast.Boolean:
		return types.I1
	case ast.String:
		return types.I8Ptr
	case ast.None:
		return types.NewPointer(tb.types["none"])
	case ast.Empty:
		return types.NewPointer(tb.types["empty"])
	case ast.Object:
		return types.NewPointer(tb.types["object"])
	}

	log.Fatalf("Expected type attribute but got: %# v", pretty.Formatter(attr))
	return nil
}

func (tb *TypeEnvBuilder) astTypeToType(astType ast.Node) types.Type {
	_, isListType := astType.(*ast.ListType)
	if isListType {
		elemType := tb.astTypeToType(astType.(*ast.ListType).ElemType)
		listType := types.NewStruct(
			types.NewPointer(elemType),
			types.I32,
			types.I1,
		)
		return types.NewPointer(tb.getOrCreate(listType))

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
		return types.NewPointer(tb.types["none"])
	case "object":
		return types.NewPointer(tb.types["object"])
	}

	log.Fatalf("Expected AST Type but got: %# v", pretty.Formatter(astType))
	return nil
}
