// Package llvmbindings defines methods for converting
// the LLVM IR generated in codegen into an executable
package llvmbindings

import (
	"log"

	"tinygo.org/x/go-llvm"
)

func init() {
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
}

type llvmTarget struct {
	targetMachine llvm.TargetMachine
	targetData    llvm.TargetData
}

func newllvmTarget() *llvmTarget {
	target, err := llvm.GetTargetFromTriple(llvm.DefaultTargetTriple())
	if err != nil {
		log.Fatalln("newllvmTarget: failed to get target from default target triple")
	}

	targetMachine := target.CreateTargetMachine(
		llvm.DefaultTargetTriple(),
		"generic",
		"",
		llvm.CodeGenOptLevel(llvm.CodeGenLevelDefault),
		llvm.RelocMode(llvm.RelocDynamicNoPic),
		llvm.CodeModel(llvm.CodeModelDefault),
	)

	targetData := targetMachine.CreateTargetData()

	return &llvmTarget{
		targetMachine: targetMachine,
		targetData:    targetData,
	}
}

func (t *llvmTarget) Dispose() {
	t.targetMachine.Dispose()
	t.targetData.Dispose()
}
