package codegen

import (
	"chogopy/pkg/ast"
	"log"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (cg *CodeGenerator) VisitForStmt(forStmt *ast.ForStmt) {
	forCondBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.cond"))
	forBodyBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.body"))
	forIncBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.inc"))
	forExitBlock := cg.currentFunction.NewBlock(cg.uniqueNames.get("for.exit"))

	iterNameInfo, err := cg.getVar(forStmt.IterName)
	if err != nil {
		log.Fatalln(err.Error())
	}

	iterName := iterNameInfo.value

	forStmt.Iter.Visit(cg)
	iterVal := cg.lastGenerated
	if isIdentOrIndex(forStmt.Iter) {
		iterVal = cg.LoadVal(iterVal)
	}

	iterLen := cg.getLen(iterVal)

	zero := constant.NewInt(types.I32, 0)
	one := constant.NewInt(types.I32, 1)

	// Initialize iteration index
	indexAlloc := cg.currentBlock.NewAlloca(types.I32)
	indexAlloc.LocalName = cg.uniqueNames.get("index_ptr")
	cg.NewStore(zero, indexAlloc)
	cg.currentBlock.NewBr(forCondBlock)

	/* Condition block */
	cg.currentBlock = forCondBlock
	index := cg.currentBlock.NewLoad(types.I32, indexAlloc)
	index.LocalName = cg.uniqueNames.get("index")
	continueLoop := cg.currentBlock.NewICmp(enum.IPredSLT, index, iterLen)
	continueLoop.LocalName = cg.uniqueNames.get("continue")
	cg.currentBlock.NewCondBr(continueLoop, forBodyBlock, forExitBlock)

	/* Body block */
	cg.currentBlock = forBodyBlock
	cg.currentBlock.NewBr(forIncBlock)

	var currentVal value.Value
	if isList(iterVal) {
		currentVal = cg.getListElem(iterVal, index)
	} else {
		currentVal = cg.getStringElem(iterVal, index)
	}

	cg.NewStore(currentVal, iterName)

	for _, bodyOp := range forStmt.Body {
		bodyOp.Visit(cg)
	}

	/* Increment block */
	cg.currentBlock = forIncBlock
	cg.currentBlock.NewBr(forCondBlock)

	incremented := cg.currentBlock.NewAdd(index, one)
	incremented.LocalName = cg.uniqueNames.get("inc")
	cg.NewStore(incremented, indexAlloc)

	/* Exit block */
	cg.currentBlock = forExitBlock
}

func (cg *CodeGenerator) getListContentPtr(list value.Value) value.Value {
	zero := constant.NewInt(types.I32, 0)
	listContentAddr := cg.currentBlock.NewGetElementPtr(
		list.Type().(*types.PointerType).ElemType,
		list,
		zero,
		zero,
	)
	listContentAddr.LocalName = cg.uniqueNames.get("list_content_addr")

	listContentType := cg.getContentType(list)
	listContentPtr := cg.currentBlock.NewLoad(listContentType, listContentAddr)
	listContentPtr.LocalName = cg.uniqueNames.get("list_content_ptr")

	return listContentPtr
}

func (cg *CodeGenerator) getListElemPtr(list value.Value, elemIdx value.Value) value.Value {
	listContentPtr := cg.getListContentPtr(list)
	listElemType := listContentPtr.Type().(*types.PointerType).ElemType
	var elemPtr value.Value
	if isList(listContentPtr) {
		contentIdx := constant.NewInt(types.I32, 0)
		elemPtr = cg.currentBlock.NewGetElementPtr(listElemType, listContentPtr, elemIdx, contentIdx)
	} else {
		elemPtr = cg.currentBlock.NewGetElementPtr(listElemType, listContentPtr, elemIdx)
	}
	return elemPtr
}

func (cg *CodeGenerator) getListElem(list value.Value, elemIdx value.Value) value.Value {
	elemPtr := cg.getListElemPtr(list, elemIdx)
	listElemType := elemPtr.Type().(*types.PointerType).ElemType

	listElem := cg.currentBlock.NewLoad(listElemType, elemPtr)
	listElem.LocalName = cg.uniqueNames.get("list_elem")
	return listElem
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
