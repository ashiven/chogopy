package codegen

import (
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

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

func (cg *CodeGenerator) getStringElem(strVal value.Value, elemIdx value.Value) value.Value {
	elemAddress := cg.currentBlock.NewGetElementPtr(types.I8, strVal, elemIdx)
	elemAddress.LocalName = cg.uniqueNames.get("str_elem_addr")
	elemVal := cg.LoadVal(elemAddress)
	elemVal = cg.clampStrSize(elemVal)
	return elemVal
}

// clampStrSize will return a copy of the given string that has size 1
// and only contains the first char of the given string.
func (cg *CodeGenerator) clampStrSize(strVal value.Value) value.Value {
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

func (cg *CodeGenerator) stringEQ(lhs value.Value, rhs value.Value) bool {
	if isString(lhs) && isString(rhs) {
		cmpResInt := cg.currentBlock.NewCall(cg.functions["strcmp"], lhs, rhs)
		cmpRes := cg.currentBlock.NewICmp(enum.IPredEQ, cmpResInt, constant.NewInt(types.I32, 0))
		cg.lastGenerated = cmpRes
		return true
	}
	return false
}

func (cg *CodeGenerator) stringNE(lhs value.Value, rhs value.Value) bool {
	if isString(lhs) && isString(rhs) {
		cmpResInt := cg.currentBlock.NewCall(cg.functions["strcmp"], lhs, rhs)
		cmpRes := cg.currentBlock.NewICmp(enum.IPredEQ, cmpResInt, constant.NewInt(types.I32, 1))
		cg.lastGenerated = cmpRes
		return true
	}
	return false
}
