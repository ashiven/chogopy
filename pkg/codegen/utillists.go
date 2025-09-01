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

// TODO: implement
func (cg *CodeGenerator) loadListElems(list value.Value) []value.Value {
	listLen := cg.getListLen(list)
	_ = listLen

	return nil
}

func (cg *CodeGenerator) concatLists(lhs value.Value, rhs value.Value, listType types.Type) value.Value {
	lhsElems := cg.loadListElems(lhs)
	rhsElems := cg.loadListElems(rhs)

	combinedElems := append(lhsElems, rhsElems...)
	combinedList := cg.newList(combinedElems, listType)

	return combinedList
}

func (cg *CodeGenerator) newList(listElems []value.Value, listType types.Type) value.Value {
	/* list alloc */
	listPtr := cg.currentBlock.NewAlloca(listType)
	listPtr.LocalName = cg.uniqueNames.get("list_ptr")

	/* list.size and list.init */
	listSize := constant.NewInt(types.I32, int64(len(listElems)))
	listInit := constant.NewBool(true)

	/* list.content alloc */
	listElemType := getListElemTypeFromListType(listType)
	listContentPtr := cg.currentBlock.NewAlloca(listElemType)
	listContentPtr.LocalName = cg.uniqueNames.get("list_content_ptr")

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
