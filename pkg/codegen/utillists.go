package codegen

import (
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

func getListElemType(list value.Value) types.Type {
	return list.Type().(*types.PointerType).ElemType.(*types.StructType).Fields[0].(*types.PointerType).ElemType
}

func getListElemTypeFromListType(listType types.Type) types.Type {
	return listType.(*types.StructType).Fields[0].(*types.PointerType).ElemType
}

func (cg *CodeGenerator) getListLen(list value.Value) value.Value {
	lenFuncName := list.Type().(*types.PointerType).ElemType.Name() + "_len"
	listLen := cg.currentBlock.NewCall(cg.functions[lenFuncName], list)
	listLen.LocalName = cg.uniqueNames.get("list_len")
	return listLen
}

func (cg *CodeGenerator) getListElemPtr(list value.Value, index value.Value) value.Value {
	getElemPtrFunc := list.Type().(*types.PointerType).ElemType.Name() + "_elemptr"
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

func (cg *CodeGenerator) setListLen(list value.Value, len value.Value) {
	zero := constant.NewInt(types.I32, 0)
	one := constant.NewInt(types.I32, 1)
	listLenAddr := cg.currentBlock.NewGetElementPtr(list.Type().(*types.PointerType).ElemType, list, zero, one)
	listLenAddr.LocalName = cg.uniqueNames.get("list_len_addr")
	cg.NewStore(len, listLenAddr)
}

func (cg *CodeGenerator) setListInit(list value.Value, init value.Value) {
	zero := constant.NewInt(types.I32, 0)
	two := constant.NewInt(types.I32, 2)
	listInitAddr := cg.currentBlock.NewGetElementPtr(list.Type().(*types.PointerType).ElemType, list, zero, two)
	listInitAddr.LocalName = cg.uniqueNames.get("list_init_addr")
	cg.NewStore(init, listInitAddr)
}

func (cg *CodeGenerator) setListContent(list value.Value, content value.Value) {
	zero := constant.NewInt(types.I32, 0)
	listContentAddr := cg.currentBlock.NewGetElementPtr(list.Type().(*types.PointerType).ElemType, list, zero, zero)
	listContentAddr.LocalName = cg.uniqueNames.get("list_content_addr")
	cg.NewStore(content, listContentAddr)
}

func (cg *CodeGenerator) concatLists(lhs value.Value, rhs value.Value, listType types.Type) value.Value {
	zero := constant.NewInt(types.I32, 0)
	four := constant.NewInt(types.I32, 4)

	lhsContentPtr := cg.getListElemPtr(lhs, zero)
	lhsLenFunc := lhs.Type().(*types.PointerType).ElemType.Name() + "_len"
	lhsLen := cg.currentBlock.NewCall(cg.functions[lhsLenFunc], lhs)
	// TODO: We are multiplying the list lengths by four because memcpy expects a length in bytes (i8) rather than words (i32).
	// However, if we are concatenating nested lists, we may need to adjust these lengths differently.
	lhsLenByte := cg.currentBlock.NewMul(lhsLen, four)

	rhsContentPtr := cg.getListElemPtr(rhs, zero)
	rhsLenFunc := rhs.Type().(*types.PointerType).ElemType.Name() + "_len"
	rhsLen := cg.currentBlock.NewCall(cg.functions[rhsLenFunc], rhs)
	rhsLenByte := cg.currentBlock.NewMul(rhsLen, four)

	concatPtr := cg.currentBlock.NewAlloca(listType)
	concatLen := cg.currentBlock.NewAdd(lhsLen, rhsLen)
	concatInit := constant.NewBool(true)

	concatListElemType := getListElemTypeFromListType(listType)
	concatContentPtr := cg.NewAllocN(concatListElemType, concatLen)
	cg.currentBlock.NewCall(cg.functions["memcpy"], concatContentPtr, lhsContentPtr, lhsLenByte)

	if isList(concatContentPtr) {
		/* Content is another list */
		contentIdx := constant.NewInt(types.I32, 0)
		concatContentPtrShifted := cg.currentBlock.NewGetElementPtr(concatContentPtr.ElemType, concatContentPtr, lhsLen, contentIdx)
		cg.currentBlock.NewCall(cg.functions["memcpy"], concatContentPtrShifted, rhsContentPtr, rhsLenByte)

	} else {
		/* Regular list content */
		concatContentPtrShifted := cg.currentBlock.NewGetElementPtr(concatContentPtr.ElemType, concatContentPtr, lhsLen)
		cg.currentBlock.NewCall(cg.functions["memcpy"], concatContentPtrShifted, rhsContentPtr, rhsLenByte)
	}

	cg.setListLen(concatPtr, concatLen)
	cg.setListInit(concatPtr, concatInit)
	cg.setListContent(concatPtr, concatContentPtr)
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
	listContentPtr := cg.NewAllocN(listElemType, listLen)
	listContentPtr.LocalName = cg.uniqueNames.get("list_content_ptr")

	/* list.content store */
	for elemIdx, elem := range listElems {
		elemIdx := constant.NewInt(types.I32, int64(elemIdx))

		if isList(listContentPtr) {
			/* Content is another list */
			contentIdx := constant.NewInt(types.I32, 0)
			elemAddr := cg.currentBlock.NewGetElementPtr(listElemType, listContentPtr, elemIdx, contentIdx)
			elemAddr.LocalName = cg.uniqueNames.get("list_content_elem_addr")
			cg.NewStore(elem, elemAddr)

		} else {
			/* Regular list content */
			elemAddr := cg.currentBlock.NewGetElementPtr(listElemType, listContentPtr, elemIdx)
			elemAddr.LocalName = cg.uniqueNames.get("list_content_elem_addr")
			cg.NewStore(elem, elemAddr)
		}
	}

	/* list store */
	cg.setListLen(listPtr, listLen)
	cg.setListInit(listPtr, listInit)
	cg.setListContent(listPtr, listContentPtr)
	return listPtr
}
