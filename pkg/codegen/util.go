package codegen

import (
	"chogopy/pkg/ast"
	"log"

	"github.com/kr/pretty"
	"github.com/llir/llvm/ir/constant"
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

func (cg *CodeGenerator) getLength(val value.Value) int {
	if _, ok := cg.lengths[val]; ok {
		return cg.lengths[val]
	}
	return cg.variables[val.Ident()[1:]].length
}

func (cg *CodeGenerator) concatLists(lhs value.Value, rhs value.Value, elemType types.Type) value.Value {
	concatListPtr := cg.currentBlock.NewAlloca(elemType)
	concatListPtr.LocalName = cg.uniqueNames.get("concat_list_ptr")
	concatListLength := 0

	// TODO: we need a method to get the lengths of lists/strings at runtime
	for i := range cg.getLength(lhs) {
		index := constant.NewInt(types.I64, int64(i))
		elemPtr := cg.currentBlock.NewGetElementPtr(lhs.Type().(*types.PointerType).ElemType, lhs, index)
		elem := cg.currentBlock.NewLoad(lhs.Type().(*types.PointerType).ElemType, elemPtr)
		cg.NewStore(elem, concatListPtr)
		concatListLength++
	}

	for i := range cg.getLength(rhs) {
		index := constant.NewInt(types.I64, int64(i))
		elemPtr := cg.currentBlock.NewGetElementPtr(rhs.Type().(*types.PointerType).ElemType, rhs, index)
		elem := cg.currentBlock.NewLoad(rhs.Type().(*types.PointerType).ElemType, elemPtr)
		cg.NewStore(elem, concatListPtr)
		concatListLength++
	}

	cg.lengths[concatListPtr] = concatListLength
	return concatListPtr
}

func (cg *CodeGenerator) needsTypeCast(val value.Value) bool {
	for _, type_ := range cg.types {
		if hasType(val, type_) || isPtrTo(val, type_) {
			return true
		}
	}
	return false
}

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
		return cg.types["none"]
	case ast.Empty:
		return cg.types["empty"]
	case ast.Object:
		return cg.types["object"]
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
		return cg.types["none"]
	case "object":
		return cg.types["object"]
	}

	log.Fatalf("Expected AST Type but got: %# v", pretty.Formatter(astType))
	return nil
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
			digitStr := cg.NewLiteral("%d\n")
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
			cg.NewStore(ifBlockRes, resAlloc)
			cg.currentBlock.NewBr(exitBlock)

			cg.currentBlock = elseBlock
			elseBlockRes := cg.NewLiteral("False\n")
			cg.NewStore(elseBlockRes, resAlloc)
			cg.currentBlock.NewBr(exitBlock)

			cg.currentBlock = exitBlock

			argVal := cg.LoadVal(resAlloc)
			printArgs = append(printArgs, argVal)

		} else {

			/* String print */

			// TODO: append newline to string before printing

			if containsCharArr(arg) {
				arg = cg.currentBlock.NewBitCast(arg, types.I8Ptr)
			} else {
				arg = cg.LoadVal(arg)
			}

			printArgs = append(printArgs, arg)
		}
	}
	return printArgs
}

// NewStore is a wrapper around the regular block.NewStore() that first
// checks whether the src or the target are of the types which require a typecast: object, none, empty.
// And if that is the case, it performs a typecast before adding a new store instruction.
func (cg *CodeGenerator) NewStore(src value.Value, target value.Value) {
	if cg.needsTypeCast(target) {
		target = cg.currentBlock.NewBitCast(target, types.NewPointer(src.Type()))
	}
	if cg.needsTypeCast(src) {
		src = cg.currentBlock.NewBitCast(src, target.Type().(*types.PointerType).ElemType)
	}

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
