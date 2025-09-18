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

// TODO: move this into its own function to avoid code repetition
func (cg *CodeGenerator) concatLists(lhs value.Value, rhs value.Value, listType types.Type) value.Value {
	zero := constant.NewInt(types.I32, 0)
	four := constant.NewInt(types.I32, 4)

	// Compute lhs list content pointer and length (word-to-byte-adjusted)
	lhsContentPtr := cg.getListElemPtr(lhs, zero)
	lhsLenFunc := lhs.Type().(*types.PointerType).ElemType.Name() + "_len"
	lhsLen := cg.currentBlock.NewCall(cg.functions[lhsLenFunc], lhs)
	lhsLen.LocalName = cg.uniqueNames.get("lhs_len_word")
	lhsLenByte := cg.currentBlock.NewMul(lhsLen, four)
	lhsLenByte.LocalName = cg.uniqueNames.get("lhs_len_byte")

	// Compute rhs list content pointer and length (word-to-byte-adjusted)
	rhsContentPtr := cg.getListElemPtr(rhs, zero)
	rhsLenFunc := rhs.Type().(*types.PointerType).ElemType.Name() + "_len"
	rhsLen := cg.currentBlock.NewCall(cg.functions[rhsLenFunc], rhs)
	rhsLen.LocalName = cg.uniqueNames.get("rhs_len_word")
	rhsLenByte := cg.currentBlock.NewMul(rhsLen, four)
	rhsLenByte.LocalName = cg.uniqueNames.get("rhs_len_byte")

	// Trick to get the size of a list struct for malloc
	listTypeSize := cg.currentBlock.NewGetElementPtr(listType, constant.NewNull(types.NewPointer(listType)), constant.NewInt(types.I32, 1))
	listTypeSize.LocalName = cg.uniqueNames.get("list_size_ptr")
	listSizeInt := cg.currentBlock.NewPtrToInt(listTypeSize, types.I32)
	listSizeInt.LocalName = cg.uniqueNames.get("list_size_int")

	// Heap-allocation for the list struct
	concatPtr := cg.currentBlock.NewCall(cg.functions["malloc"], listSizeInt)
	concatPtr.LocalName = cg.uniqueNames.get("concat_ptr")
	concatPtrCast := cg.currentBlock.NewBitCast(concatPtr, types.NewPointer(listType))
	concatPtrCast.LocalName = cg.uniqueNames.get("concat_ptr_cast")
	cg.heapAllocs = append(cg.heapAllocs, concatPtrCast)

	// Initial values for list init and length
	concatLen := cg.currentBlock.NewAdd(lhsLen, rhsLen)
	concatLen.LocalName = cg.uniqueNames.get("concat_len_word")
	concatLenByte := cg.currentBlock.NewAdd(lhsLenByte, rhsLenByte)
	concatLenByte.LocalName = cg.uniqueNames.get("concat_len_byte")
	concatInit := constant.NewBool(true)

	// Heap-allocation for the list content
	concatListElemType := getListElemTypeFromListType(listType)
	concatContentPtr := cg.currentBlock.NewCall(cg.functions["malloc"], concatLenByte)
	concatContentPtr.LocalName = cg.uniqueNames.get("concat_content_ptr")
	concatContentPtrCast := cg.currentBlock.NewBitCast(concatContentPtr, types.NewPointer(concatListElemType))
	concatContentPtrCast.LocalName = cg.uniqueNames.get("concat_content_ptr")
	cg.heapAllocs = append(cg.heapAllocs, concatContentPtrCast)

	// Copy lhs into concat content and then rhs into shifted concat content ptr
	lhsCpyRes := cg.currentBlock.NewCall(cg.functions["memcpy"], concatContentPtrCast, lhsContentPtr, lhsLenByte)
	lhsCpyRes.LocalName = cg.uniqueNames.get("lhs_cpy_res")

	if isList(concatContentPtrCast) {
		/* Content is another list */
		contentIdx := constant.NewInt(types.I32, 0)
		concatContentPtrShifted := cg.currentBlock.NewGetElementPtr(concatListElemType, concatContentPtrCast, lhsLen, contentIdx)
		concatContentPtrShifted.LocalName = cg.uniqueNames.get("concat_content_ptr_shifted")
		rhsCpyRes := cg.currentBlock.NewCall(cg.functions["memcpy"], concatContentPtrShifted, rhsContentPtr, rhsLenByte)
		rhsCpyRes.LocalName = cg.uniqueNames.get("rhs_cpy_res")

	} else {
		/* Regular list content */
		concatContentPtrShifted := cg.currentBlock.NewGetElementPtr(concatListElemType, concatContentPtrCast, lhsLen)
		concatContentPtrShifted.LocalName = cg.uniqueNames.get("concat_content_ptr_shifted")
		rhsCpyRes := cg.currentBlock.NewCall(cg.functions["memcpy"], concatContentPtrShifted, rhsContentPtr, rhsLenByte)
		rhsCpyRes.LocalName = cg.uniqueNames.get("rhs_cpy_res")
	}

	cg.setListLen(concatPtrCast, concatLen)
	cg.setListInit(concatPtrCast, concatInit)
	cg.setListContent(concatPtrCast, concatContentPtrCast)
	return concatPtrCast
}

// newList dynamically allocates and returns a pointer to a list literal
// on the call stack of the currently executing function.
// This is problematic if the list literal were to be used outside of the
// current function, for instance, if the function returned a pointer to this list.
// This pointer would then point to unallocated memory (the call stack is freed after function return)
func (cg *CodeGenerator) newList(listElems []value.Value, listType types.Type) value.Value {
	/* list.size and list.init */
	listLen := constant.NewInt(types.I32, int64(len(listElems)))
	listLenByte := constant.NewInt(types.I32, int64(len(listElems)*4))
	listInit := constant.NewBool(true)

	/* list.content alloc */
	listElemType := getListElemTypeFromListType(listType)
	listContentPtr := cg.currentBlock.NewCall(cg.functions["malloc"], listLenByte)
	listContentPtr.LocalName = cg.uniqueNames.get("list_content_ptr")
	listContentPtrCast := cg.currentBlock.NewBitCast(listContentPtr, types.NewPointer(listElemType))
	listContentPtrCast.LocalName = cg.uniqueNames.get("list_content_ptr_cast")
	cg.heapAllocs = append(cg.heapAllocs, listContentPtrCast)

	/* list.content store */
	for elemIdx, elem := range listElems {
		elemIdx := constant.NewInt(types.I32, int64(elemIdx))

		if isList(listContentPtrCast) {
			/* Content is another list */
			contentIdx := constant.NewInt(types.I32, 0)
			elemAddr := cg.currentBlock.NewGetElementPtr(listElemType, listContentPtrCast, elemIdx, contentIdx)
			elemAddr.LocalName = cg.uniqueNames.get("list_content_elem_addr")
			cg.NewStore(elem, elemAddr)

		} else {
			/* Regular list content */
			elemAddr := cg.currentBlock.NewGetElementPtr(listElemType, listContentPtrCast, elemIdx)
			elemAddr.LocalName = cg.uniqueNames.get("list_content_elem_addr")
			cg.NewStore(elem, elemAddr)
		}
	}

	/* trick to get the size of a list struct for malloc */
	listTypeSize := cg.currentBlock.NewGetElementPtr(listType, constant.NewNull(types.NewPointer(listType)), constant.NewInt(types.I32, 1))
	listTypeSize.LocalName = cg.uniqueNames.get("list_size_ptr")
	listSizeInt := cg.currentBlock.NewPtrToInt(listTypeSize, types.I32)
	listSizeInt.LocalName = cg.uniqueNames.get("list_size_int")

	/* list alloc */
	listPtr := cg.currentBlock.NewCall(cg.functions["malloc"], listSizeInt)
	listPtr.LocalName = cg.uniqueNames.get("list_ptr")
	listPtrCast := cg.currentBlock.NewBitCast(listPtr, types.NewPointer(listType))
	listPtrCast.LocalName = cg.uniqueNames.get("list_ptr_cast")
	cg.heapAllocs = append(cg.heapAllocs, listPtrCast)

	/* list store */
	cg.setListLen(listPtrCast, listLen)
	cg.setListInit(listPtrCast, listInit)
	cg.setListContent(listPtrCast, listContentPtrCast)
	return listPtrCast
}
