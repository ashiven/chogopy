package codegen

import (
	"fmt"
	"strings"

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

func (cg *CodeGenerator) getListLen(list value.Value) value.Value {
	listLen := cg.currentBlock.NewCall(cg.functions["listlen"], list)
	listLen.LocalName = cg.uniqueNames.get("list_len")
	return listLen
}

func (cg *CodeGenerator) getListElemPtr(list value.Value, elemIdx value.Value) value.Value {
	listTypeName := list.Type().(*types.PointerType).ElemType.Name()
	getElemPtrFunc := fmt.Sprintf("%s_elemptr", listTypeName)
	elemPtr := cg.currentBlock.NewCall(cg.functions[getElemPtrFunc], list, elemIdx)
	elemPtr.LocalName = cg.uniqueNames.get("list_elem_ptr")
	return elemPtr
}

func (cg *CodeGenerator) getListElem(list value.Value, elemIdx value.Value) value.Value {
	elemPtr := cg.getListElemPtr(list, elemIdx)

	listElemType := elemPtr.Type().(*types.PointerType).ElemType
	listElem := cg.currentBlock.NewLoad(listElemType, elemPtr)
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
