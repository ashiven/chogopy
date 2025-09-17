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
		cmpRes := cg.currentBlock.NewICmp(enum.IPredNE, cmpResInt, constant.NewInt(types.I32, 0))
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

	strLen := cg.currentBlock.NewCall(cg.functions["strlen"], strVal)
	strLen.LocalName = cg.uniqueNames.get("str_len")
	copyBuffer := cg.currentBlock.NewCall(cg.functions["malloc"], strLen)
	copyBuffer.LocalName = cg.uniqueNames.get("copy_buffer")
	cg.heapAllocs = append(cg.heapAllocs, copyBuffer)
	copyBuffer.LocalName = cg.uniqueNames.get("clamp_buf_ptr")
	copyRes := cg.currentBlock.NewCall(cg.functions["strcpy"], copyBuffer, strVal)
	copyRes.LocalName = cg.uniqueNames.get("clamp_copy_res")

	elemAddr := cg.currentBlock.NewGetElementPtr(types.I8, copyBuffer, one)
	elemAddr.LocalName = cg.uniqueNames.get("clamp_addr")
	cg.NewStore(term, elemAddr)

	return copyBuffer
}

// TODO: maybe switch to malloc for string literals so I don't have
// to differentiate between statically allocated strings and heap allocated strings for function returns
func (cg *CodeGenerator) concatStrings(lhs value.Value, rhs value.Value) value.Value {
	// 1) Allocate a destination buffer of size: char[lhsLen + rhsLen + 1] (one more for the zero byte)
	lhsLen := cg.currentBlock.NewCall(cg.functions["strlen"], lhs)
	lhsLen.LocalName = cg.uniqueNames.get("lhs_len")
	rhsLen := cg.currentBlock.NewCall(cg.functions["strlen"], rhs)
	rhsLen.LocalName = cg.uniqueNames.get("rhs_len")
	concatLen := cg.currentBlock.NewAdd(lhsLen, rhsLen)
	concatLen = cg.currentBlock.NewAdd(concatLen, constant.NewInt(types.I32, 1))
	concatLen.LocalName = cg.uniqueNames.get("concat_len")
	concatStr := cg.currentBlock.NewCall(cg.functions["malloc"], concatLen)
	concatStr.LocalName = cg.uniqueNames.get("concat_str")
	cg.heapAllocs = append(cg.heapAllocs, concatStr)

	// 2) Copy the string that should be appended to into that buffer
	copyRes := cg.currentBlock.NewCall(cg.functions["strcpy"], concatStr, lhs)
	copyRes.LocalName = cg.uniqueNames.get("concat_copy_res")

	// 3) Append the other string via strcat
	concatRes := cg.currentBlock.NewCall(cg.functions["strcat"], concatStr, rhs)
	concatRes.LocalName = cg.uniqueNames.get("concat_append_res")

	return concatRes
}

// func (cg *CodeGenerator) concatStrings(lhs string, rhs string) value.Value {
// 	concatStr := cg.NewLiteral(lhs + rhs)
// 	return concatStr
// }
