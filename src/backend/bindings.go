// Package backend defines methods for converting
// the LLVM IR generated in codegen into an executable
//
// The code in this package is heavily inspired by:
//
// DDP-Projekt/Kompilierer by Hendrik Ziegler
// https://github.com/DDP-Projekt/Kompilierer
package backend

import (
	"io"
	"log"

	"tinygo.org/x/go-llvm"
)

func Init() {
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

func newllvmTarget() llvmTarget {
	target, err := llvm.GetTargetFromTriple(llvm.DefaultTargetTriple())
	if err != nil {
		log.Fatalln("newllvmTarget: failed to get target from default target triple: ", err)
	}

	targetMachine := target.CreateTargetMachine(
		llvm.DefaultTargetTriple(),
		"",
		"",
		llvm.CodeGenOptLevel(llvm.CodeGenLevelDefault),
		llvm.RelocMode(llvm.RelocDefault),
		llvm.CodeModel(llvm.CodeModelDefault),
	)

	targetData := targetMachine.CreateTargetData()

	return llvmTarget{
		targetMachine: targetMachine,
		targetData:    targetData,
	}
}

func (t llvmTarget) Dispose() {
	t.targetMachine.Dispose()
	t.targetData.Dispose()
}

type llvmContext struct {
	context     llvm.Context
	passOptions llvm.PassBuilderOptions
	llvmTarget
}

func NewllvmContext() *llvmContext {
	llvmContext := &llvmContext{}
	llvmContext.llvmTarget = newllvmTarget()
	llvmContext.context = llvm.NewContext()
	llvmContext.passOptions = llvm.NewPassBuilderOptions()

	return llvmContext
}

func (c *llvmContext) Dispose() {
	c.context.Dispose()
	c.passOptions.Dispose()
	c.llvmTarget.Dispose()
}

/* Specific module passes can be added here via commands like c.passOptions.SetLoopInterLeaving(true) .
*
* Module passes can also be run with an optimization level as follows:
*
* module.RunPasses("default<03>", targetMachine, passOptions)*/
func (c *llvmContext) RegisterPasses() {}

func (c *llvmContext) ParseIRFromFile(path string) llvm.Module {
	memBuffer, err := llvm.NewMemoryBufferFromFile(path)
	if err != nil {
		log.Fatalln("parseIrFromFile: failed to create memory buffer from .ll file: ", err)
	}

	module, err := c.context.ParseIR(memBuffer)
	if err != nil {
		log.Fatalln("parseIrFromFile: failed to create module from memory buffer: ", err)
	}

	module.SetTarget(c.targetMachine.Triple())
	module.SetDataLayout(c.targetData.String())

	return module
}

/* Uses optimization level O3 by default for now.
* Maybe add this as a compiler option later. */
func (c *llvmContext) OptimizeModule(module llvm.Module) {
	module.RunPasses("default<O3>", c.targetMachine, c.passOptions)
}

func (c *llvmContext) CompileModule(module llvm.Module, codeGenType llvm.CodeGenFileType, w io.Writer) (int, error) {
	compiledBuffer, err := c.targetMachine.EmitToMemoryBuffer(module, codeGenType)
	if err != nil {
		log.Fatalln("compileModule: failed to compile module: ", err)
	}

	defer compiledBuffer.Dispose()

	return w.Write(compiledBuffer.Bytes())
}
