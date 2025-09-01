package codegen

import (
	"fmt"
	"strings"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func isList(val value.Value) bool {
	if _, ok := val.Type().(*types.PointerType); ok {
		if isListType(val.Type().(*types.PointerType).ElemType) {
			return true
		}
	}
	return false
}

func isListType(type_ types.Type) bool {
	return strings.Contains(type_.Name(), "list") && type_.Name() != "list" && type_.Name() != "list_content"
}

func getListElemType(list value.Value) types.Type {
	return list.Type().(*types.PointerType).ElemType.(*types.StructType).Fields[0].(*types.PointerType).ElemType
}

func getListElemTypeFromListType(listType types.Type) types.Type {
	return listType.(*types.StructType).Fields[0].(*types.PointerType).ElemType
}

func (cg *CodeGenerator) getListLen(list value.Value) value.Value {
	listLen := cg.currentBlock.NewCall(cg.functions["listlen"], list)
	listLen.LocalName = cg.uniqueNames.get("list_len")
	return listLen
}

func (cg *CodeGenerator) getListElemPtr(list value.Value, index value.Value) value.Value {
	listTypeName := list.Type().(*types.PointerType).ElemType.Name()
	getElemPtrFunc := fmt.Sprintf("%s_elemptr", listTypeName)
	listElemPtr := cg.currentBlock.NewCall(cg.functions[getElemPtrFunc], list, index)
	listElemPtr.LocalName = cg.uniqueNames.get("list_elem_ptr")
	return listElemPtr
}

func (cg *CodeGenerator) getListElem(list value.Value, index value.Value) value.Value {
	listElemPtr := cg.getListElemPtr(list, index)
	listElemType := getListElemType(list)
	listElem := cg.currentBlock.NewLoad(listElemType, listElemPtr)
	listElem.LocalName = cg.uniqueNames.get("list_elem")
	return listElem
}

func (cg *CodeGenerator) setLen(list value.Value, len value.Value) {
	zero := constant.NewInt(types.I32, 0)
	one := constant.NewInt(types.I32, 1)
	listLenAddr := cg.currentBlock.NewGetElementPtr(list.Type().(*types.PointerType).ElemType, list, zero, one)
	listLenAddr.LocalName = cg.uniqueNames.get("list_len_addr")
	cg.NewStore(len, listLenAddr)
}

func (cg *CodeGenerator) setInit(list value.Value, init value.Value) {
	zero := constant.NewInt(types.I32, 0)
	two := constant.NewInt(types.I32, 2)
	listInitAddr := cg.currentBlock.NewGetElementPtr(list.Type().(*types.PointerType).ElemType, list, zero, two)
	listInitAddr.LocalName = cg.uniqueNames.get("list_init_addr")
	cg.NewStore(init, listInitAddr)
}

func (cg *CodeGenerator) setContent(list value.Value, content value.Value) {
	zero := constant.NewInt(types.I32, 0)
	listContentAddr := cg.currentBlock.NewGetElementPtr(list.Type().(*types.PointerType).ElemType, list, zero, zero)
	listContentAddr.LocalName = cg.uniqueNames.get("list_content_addr")
	cg.NewStore(content, listContentAddr)
}

func (cg *CodeGenerator) concatLists(lhs value.Value, rhs value.Value, listType types.Type) value.Value {
	zero := constant.NewInt(types.I32, 0)
	lhsContentPtr := cg.getListElemPtr(lhs, zero)
	lhsLen := cg.currentBlock.NewCall(cg.functions["listlen"], lhs)

	rhsContentPtr := cg.getListElemPtr(rhs, zero)
	rhsLen := cg.currentBlock.NewCall(cg.functions["listlen"], rhs)

	concatLen := cg.currentBlock.NewAdd(lhsLen, rhsLen)
	concatInit := constant.NewBool(true)

	// TODO: have to allocate enough for the actual size of the concatenated list (currently only allocates for one element)
	concatPtr := cg.currentBlock.NewAlloca(listType)
	concatListElemType := getListElemTypeFromListType(listType)

	concatContentPtr := cg.currentBlock.NewAlloca(concatListElemType)
	cg.currentBlock.NewCall(cg.functions["memcpy"], concatContentPtr, lhsContentPtr, lhsLen)

	// TODO: adjust for nested lists
	concatContentPtrShifted := cg.currentBlock.NewGetElementPtr(concatContentPtr.ElemType, concatContentPtr, lhsLen)
	cg.currentBlock.NewCall(cg.functions["memcpy"], concatContentPtrShifted, rhsContentPtr, rhsLen)

	cg.setLen(concatPtr, concatLen)
	cg.setInit(concatPtr, concatInit)
	cg.setContent(concatPtr, concatContentPtr)

	return concatPtr
}

func (cg *CodeGenerator) newList(listElems []value.Value, listType types.Type) value.Value {
	/* list alloc */
	listPtr := cg.currentBlock.NewAlloca(listType)
	listPtr.LocalName = cg.uniqueNames.get("list_ptr")

	/* list.size and list.init */
	listLen := constant.NewInt(types.I32, int64(len(listElems)))
	listInit := constant.NewBool(true)

	/* list.content alloc */
	listElemType := getListElemTypeFromListType(listType)
	listContentPtr := cg.currentBlock.NewAlloca(types.NewArray(uint64(len(listElems)), listElemType))
	listContentPtr.LocalName = cg.uniqueNames.get("list_content_ptr")
	listContentPtrCast := cg.currentBlock.NewBitCast(listContentPtr, types.NewPointer(listElemType))
	listContentPtrCast.LocalName = cg.uniqueNames.get("list_content_ptr_cast")

	/* list.content store */
	for elemIdx, elem := range listElems {
		elemIdx := constant.NewInt(types.I32, int64(elemIdx))
		var elemAddr *ir.InstGetElementPtr
		if isList(listContentPtr) {
			contentIdx := constant.NewInt(types.I32, 0)
			elemAddr = cg.currentBlock.NewGetElementPtr(listElemType, listContentPtr, elemIdx, contentIdx)
		} else {
			elemAddr = cg.currentBlock.NewGetElementPtr(listElemType, listContentPtr, elemIdx)
		}
		elemAddr.LocalName = cg.uniqueNames.get("list_content_elem_addr")
		cg.NewStore(elem, elemAddr)
	}

	/* list store */
	cg.setLen(listPtr, listLen)
	cg.setInit(listPtr, listInit)
	cg.setContent(listPtr, listContentPtrCast)
	return listPtr
}
