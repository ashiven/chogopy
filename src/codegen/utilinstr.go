package codegen

import (
	"log"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// NewStore is a wrapper around the regular ir.Block.NewStore() that first checks whether the src requires a typecast.
func (cg *CodeGenerator) NewStore(src value.Value, target value.Value) {
	if !isPtrTo(target, src.Type()) {
		target = cg.currentBlock.NewBitCast(target, types.NewPointer(src.Type()))
		target.(*ir.InstBitCast).LocalName = cg.uniqueNames.get("store_cast")
	}

	cg.currentBlock.NewStore(src, target)
}

// NewLiteral takes any literal of type int, bool, string, or nil and creates a new allocation and store for that value.
// It returns an SSA value containing the value of the given literal with the following types:
// int: i32   str: i8*   bool: i1   nil: %none*
func (cg *CodeGenerator) NewLiteral(literal any) value.Value {
	switch literal := literal.(type) {
	case int:
		intConst := constant.NewInt(types.I32, int64(literal))
		intLiteral := cg.currentBlock.NewCall(cg.functions["newint"], intConst)
		intLiteral.LocalName = cg.uniqueNames.get("int_literal")
		return intLiteral

	case bool:
		boolConst := constant.NewBool(literal)
		boolLiteral := cg.currentBlock.NewCall(cg.functions["newbool"], boolConst)
		boolLiteral.LocalName = cg.uniqueNames.get("bool_literal")
		return boolLiteral

	case string:
		charArrConst := constant.NewCharArrayFromString(literal + "\x00")
		charArrGlobal := cg.Module.NewGlobalDef(cg.uniqueNames.get("str"), charArrConst)

		// Store static string into stack-allocated string ptr
		zero := constant.NewInt(types.I32, 0)
		strConst := constant.NewGetElementPtr(charArrGlobal.Typ.ElemType, charArrGlobal, zero, zero)
		strPtrStack := cg.currentBlock.NewAlloca(types.I8Ptr)
		strPtrStack.LocalName = cg.uniqueNames.get("str_ptr_stack")
		cg.NewStore(strConst, strPtrStack)
		strStack := cg.currentBlock.NewLoad(types.I8Ptr, strPtrStack)
		strStack.LocalName = cg.uniqueNames.get("str_stack")

		// Copy string into heap-allocated string ptr
		strHeap := cg.NewMalloc(types.I8, constant.NewInt(types.I32, int64(len(literal)+1)))
		strCopy := cg.currentBlock.NewCall(cg.functions["sprintf"], strHeap, cg.strings["str_format"], strStack)
		strCopy.LocalName = cg.uniqueNames.get("strcpy_res")

		return strHeap

	case nil:
		noneConst := constant.NewNull(types.NewPointer(cg.types["none"]))
		nonePtr := cg.currentBlock.NewAlloca(noneConst.Type())
		nonePtr.LocalName = cg.uniqueNames.get("none_ptr")
		cg.NewStore(noneConst, nonePtr)
		noneLoad := cg.currentBlock.NewLoad(noneConst.Type(), nonePtr)
		noneLoad.LocalName = cg.uniqueNames.get("none_val")
		return noneLoad
	}

	log.Fatalln("NewLiteral: expected literal of type int, bool, str, or nil")
	return nil
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

	case valIsPtr:
		valueLoad := cg.currentBlock.NewLoad(val.Type().(*types.PointerType).ElemType, val)
		valueLoad.LocalName = cg.uniqueNames.get("val")
		return valueLoad

	default:
		return val
	}
}

func (cg *CodeGenerator) NewAllocN(elemType types.Type, NElems value.Value) *ir.InstAlloca {
	instAlloc := &ir.InstAlloca{ElemType: elemType, NElems: NElems}
	instAlloc.Type()
	cg.currentBlock.Insts = append(cg.currentBlock.Insts, instAlloc)
	return instAlloc
}

func (cg *CodeGenerator) NewMalloc(elemType types.Type, NElems value.Value) value.Value {
	elemPtr := cg.currentBlock.NewCall(cg.functions["malloc"], cg.sizeof(elemType, NElems))
	elemPtr.LocalName = cg.uniqueNames.get("heap_ptr")
	elemPtrCast := cg.currentBlock.NewBitCast(elemPtr, types.NewPointer(elemType))
	elemPtrCast.LocalName = cg.uniqueNames.get("heap_ptr_cast")
	cg.heapAllocs = append(cg.heapAllocs, elemPtrCast)

	return elemPtrCast
}
