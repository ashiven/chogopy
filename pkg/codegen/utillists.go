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

func (cg *CodeGenerator) concatLists(lhs value.Value, rhs value.Value, elemType types.Type) value.Value {
	concatListPtr := cg.currentBlock.NewAlloca(elemType)
	concatListPtr.LocalName = cg.uniqueNames.get("concat_list_ptr")
	concatListLength := 0

	// TODO: we need a method to get the lengths of lists at runtime
	for i := range 0 {
		index := constant.NewInt(types.I64, int64(i))
		elemPtr := cg.currentBlock.NewGetElementPtr(lhs.Type().(*types.PointerType).ElemType, lhs, index)
		elem := cg.currentBlock.NewLoad(lhs.Type().(*types.PointerType).ElemType, elemPtr)
		cg.NewStore(elem, concatListPtr)
		concatListLength++
	}

	for i := range 0 {
		index := constant.NewInt(types.I64, int64(i))
		elemPtr := cg.currentBlock.NewGetElementPtr(rhs.Type().(*types.PointerType).ElemType, rhs, index)
		elem := cg.currentBlock.NewLoad(rhs.Type().(*types.PointerType).ElemType, elemPtr)
		cg.NewStore(elem, concatListPtr)
		concatListLength++
	}

	return concatListPtr
}

func (cg *CodeGenerator) newList(listElems []value.Value, listType types.Type) value.Value {
	/* list alloc */
	listPtr := cg.currentBlock.NewAlloca(listType)
	listPtr.LocalName = cg.uniqueNames.get("list_ptr")

	/* list.content alloc */
	listElemType := getListElemTypeFromListType(listType)
	listContentPtr := cg.currentBlock.NewAlloca(listElemType)
	listContentPtr.LocalName = cg.uniqueNames.get("list_content_ptr")

	/* list.content store */
	for elemIdx, elem := range listElems {
		// In case the content of the list is a pointer to another list (lists have a struct type)
		// the first GEP index will select the list struct and the second index will select the field to store into (list.content)
		// Think about it like this:
		//
		// - You have a pointer to a struct list: list*
		// - This points to a contiguous location in memory at which one or more list structs reside
		//
		// Memory(starting at list*):
		//
		// 							0:	list{content: i32*, size: i32}
		// elemIdx ->		1:	list{content: i32*, size: i32}
		// 							2:	list{content: i32*, size: i32}
		// 							3:	list{content: i32*, size: i32}
		//													^
		//													|
		//										contentIdx
		//
		// - Now GEP will first need to know which of these lists to address (elemIdx)
		// - Then GEP will want to know which field of the list to address (contentIdx)

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

	/* list.size */
	listSize := constant.NewInt(types.I32, int64(len(listElems)))

	/* list.init */
	listInit := constant.NewBool(true)

	/* list store */
	zero := constant.NewInt(types.I32, 0)
	one := constant.NewInt(types.I32, 1)
	two := constant.NewInt(types.I32, 2)

	listContentAddr := cg.currentBlock.NewGetElementPtr(listType, listPtr, zero, zero)
	listContentAddr.LocalName = cg.uniqueNames.get("list_content_addr")
	cg.NewStore(listContentPtr, listContentAddr)

	listSizeAddr := cg.currentBlock.NewGetElementPtr(listType, listPtr, zero, one)
	listSizeAddr.LocalName = cg.uniqueNames.get("list_size_addr")
	cg.NewStore(listSize, listSizeAddr)

	listInitAddr := cg.currentBlock.NewGetElementPtr(listType, listPtr, zero, two)
	listInitAddr.LocalName = cg.uniqueNames.get("list_init_addr")
	cg.NewStore(listInit, listInitAddr)

	return listPtr
}
