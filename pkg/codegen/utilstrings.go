package codegen

import (
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

var MaxBufferSize = uint64(10000)

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

func (cg *CodeGenerator) getStringLen(strVal value.Value) value.Value {
	strLen := cg.currentBlock.NewCall(cg.functions["strlen"], strVal)
	strLen.LocalName = cg.uniqueNames.get("str_len")
	return strLen
}

func (cg *CodeGenerator) getStringElem(strVal value.Value, elemIdx value.Value) value.Value {
	elemAddress := cg.currentBlock.NewGetElementPtr(types.I8, strVal, elemIdx)
	elemAddress.LocalName = cg.uniqueNames.get("str_elem_addr")
	elemVal := cg.LoadVal(elemAddress)
	elemVal = cg.clampString(elemVal)
	return elemVal
}

func (cg *CodeGenerator) stringEqual(lhs value.Value, rhs value.Value) bool {
	if isString(lhs) && isString(rhs) {
		cmpResInt := cg.currentBlock.NewCall(cg.functions["strcmp"], lhs, rhs)
		cmpRes := cg.currentBlock.NewICmp(enum.IPredEQ, cmpResInt, constant.NewInt(types.I32, 0))
		cg.lastGenerated = cmpRes
		return true
	}
	return false
}

func (cg *CodeGenerator) stringNotEqual(lhs value.Value, rhs value.Value) bool {
	if isString(lhs) && isString(rhs) {
		cmpResInt := cg.currentBlock.NewCall(cg.functions["strcmp"], lhs, rhs)
		cmpRes := cg.currentBlock.NewICmp(enum.IPredEQ, cmpResInt, constant.NewInt(types.I32, 1))
		cg.lastGenerated = cmpRes
		return true
	}
	return false
}

// clampStrSize will return a copy of the given string that has size 1
// and only contains the first char of the given string.
func (cg *CodeGenerator) clampString(strVal value.Value) value.Value {
	one := constant.NewInt(types.I32, 1)
	term := constant.NewCharArrayFromString("\x00")

	copyBuffer := cg.currentBlock.NewAlloca(types.NewArray(uint64(2), types.I8))
	copyBuffer.LocalName = cg.uniqueNames.get("clamp_buf_ptr")
	copyRes := cg.currentBlock.NewCall(cg.functions["strcpy"], copyBuffer, strVal)
	copyRes.LocalName = cg.uniqueNames.get("clamp_copy_res")

	elemAddr := cg.currentBlock.NewGetElementPtr(types.I8, copyBuffer, one)
	elemAddr.LocalName = cg.uniqueNames.get("clamp_addr")
	cg.NewStore(term, elemAddr)

	strCast := cg.toString(copyBuffer)
	return strCast
}

func (cg *CodeGenerator) concatStrings(lhs value.Value, rhs value.Value) value.Value {
	// 1) Allocate a destination buffer of size: char[BUFFER_SIZE] (needs extra space for stuff to be appended)
	destBuffer := cg.currentBlock.NewAlloca(types.NewArray(MaxBufferSize, types.I8))
	destBuffer.LocalName = cg.uniqueNames.get("concat_buffer_ptr")
	destBufferCast := cg.toString(destBuffer)

	// 2) Copy the string that should be appended to into that buffer
	copyRes := cg.currentBlock.NewCall(cg.functions["strcpy"], destBufferCast, lhs)
	copyRes.LocalName = cg.uniqueNames.get("concat_copy_res")

	// 3) Append the other string via strcat
	concatRes := cg.currentBlock.NewCall(cg.functions["strcat"], destBufferCast, rhs)
	concatRes.LocalName = cg.uniqueNames.get("concat_append_res")

	return concatRes
}
