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

// convertPrintArgs converts a list of argument values serving as input
// to a call of the print function so that they are printed correctly.
// For example, given an arg of type bool (I1), it converts it into the string "True"
// if its value is 1, or "False" if its value is 0.
func (cg *CodeGenerator) convertPrintArgs(args []value.Value) []value.Value {
	printArgs := []value.Value{}
	for _, arg := range args {
		if hasType(arg, types.I32) || isPtrTo(arg, types.I32) {

			/* Integer print */
			digitStr := cg.NewLiteral("%d")
			argVal := cg.LoadVal(arg)
			printArgs = append(printArgs, digitStr)
			printArgs = append(printArgs, argVal)
		} else if hasType(arg, types.I1) || isPtrTo(arg, types.I1) {

			/* Boolean print
			> Uses a ternary to determine whether to print "True" or "False" based on the given I1 argument */
			ifBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("boolprint.then"))
			elseBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("boolprint.else"))
			exitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("boolprint.exit"))

			resAlloc := cg.currentBlock.NewAlloca(types.I8Ptr)
			resAlloc.LocalName = cg.uniqueNames.get("boolprint_res_ptr")

			cond := arg
			cond = cg.LoadVal(cond)
			cg.currentBlock.NewCondBr(cond, ifBlock, elseBlock)

			cg.currentBlock = ifBlock
			ifBlockRes := cg.NewLiteral("True\n")
			cg.currentBlock.NewStore(ifBlockRes, resAlloc)
			cg.currentBlock.NewBr(exitBlock)

			cg.currentBlock = elseBlock
			elseBlockRes := cg.NewLiteral("False\n")
			cg.currentBlock.NewStore(elseBlockRes, resAlloc)
			cg.currentBlock.NewBr(exitBlock)

			cg.currentBlock = exitBlock

			argVal := cg.LoadVal(resAlloc)
			printArgs = append(printArgs, argVal)

		} else {

			/* String print */
			argVal := cg.LoadVal(arg)
			printArgs = append(printArgs, argVal)
		}
	}
	return printArgs
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
		literalConst = constant.NewNull(types.I8Ptr)
	}

	literalAlloc := cg.currentBlock.NewAlloca(literalConst.Type())
	literalAlloc.LocalName = cg.uniqueNames.get("literal_ptr")
	cg.currentBlock.NewStore(literalConst, literalAlloc)

	if _, ok := literal.(string); ok {
		literalAllocCast := cg.currentBlock.NewBitCast(literalAlloc, types.I8Ptr)
		literalAllocCast.LocalName = cg.uniqueNames.get("literal_val")
		return literalAllocCast
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
