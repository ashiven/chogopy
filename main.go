package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"chogopy/pkg/backend"
	"chogopy/pkg/codegen"
	"chogopy/pkg/lexer"
	"chogopy/pkg/parser"
	"chogopy/pkg/scopes"
	"chogopy/pkg/typechecks"

	"github.com/kr/pretty"
	"tinygo.org/x/go-llvm"
)

func main() {
	filePath := ""

	if len(os.Args) > 1 {
		filePath = os.Args[len(os.Args)-1]
	} else {
		log.Fatal("Please provide a file path.")
	}

	llFilePath := replaceFileEnding(filePath, "ll")
	objectFilePath := replaceFileEnding(filePath, "o")

	byteStream, err := os.ReadFile(filePath)
	if err != nil {
		pretty.Println(err.Error())
	}
	stream := string(byteStream)

	myLexer := lexer.NewLexer(stream)
	myParser := parser.NewParser(&myLexer)

	if len(os.Args) > 2 {
		switch os.Args[1] {
		case "-l":
			token := myLexer.Consume(false)
			pretty.Println(token)
			for token.Kind != lexer.EOF {
				token = myLexer.Consume(false)
				pretty.Println(token)
			}
		case "-p":
			program := myParser.ParseProgram()
			pretty.Println(program)
		case "-t":
			program := myParser.ParseProgram()
			staticTyping := typechecks.StaticTyping{}
			staticTyping.Analyze(&program)
		case "-n":
			program := myParser.ParseProgram()
			assignTargets := scopes.AssignTargets{}
			assignTargets.Analyze(&program)
			scopes := scopes.NameScopes{}
			scopes.Analyze(&program)
		case "-c":
			program := myParser.ParseProgram()
			assignTargets := scopes.AssignTargets{}
			assignTargets.Analyze(&program)
			nameScopes := scopes.NameScopes{}
			nameScopes.Analyze(&program)
			staticTyping := typechecks.StaticTyping{}
			staticTyping.Analyze(&program)
			codeGenerator := codegen.CodeGenerator{}
			codeGenerator.Generate(&program)

			// TODO: To keep the test cases working I am only appending .ll to the filePath
			// here but will have to change that in the future and modify the test cases accordingly.
			err := os.WriteFile(
				replaceFileEnding(filePath, "ll"),
				[]byte(codeGenerator.Module.String()),
				0o644,
			)
			if err != nil {
				panic(err)
			}
		}
	} else {
		program := myParser.ParseProgram()
		assignTargets := scopes.AssignTargets{}
		assignTargets.Analyze(&program)
		nameScopes := scopes.NameScopes{}
		nameScopes.Analyze(&program)
		staticTyping := typechecks.StaticTyping{}
		staticTyping.Analyze(&program)
		codeGenerator := codegen.CodeGenerator{}
		codeGenerator.Generate(&program)

		err := os.WriteFile(
			llFilePath,
			[]byte(codeGenerator.Module.String()),
			0o644,
		)
		if err != nil {
			log.Fatalln("Failed to create llvm IR file: ", err)
		}

		backend.Init()
		llvmContext := backend.NewllvmContext()
		defer llvmContext.Dispose()

		module := llvmContext.ParseIRFromFile(llFilePath)
		os.Remove(llFilePath)

		// TODO: This breaks some of the modules which are not entirely correct yet
		// (For instance, modules in which functions allocate strings or lists on their call stack
		// and then return pointers to the now unallocated memory. This still happens when input() or
		// list/string concatenation is used and the resulting value is returned and then used by the caller.)
		llvmContext.OptimizeModule(module)

		objectFile, err := os.Create(objectFilePath)
		if err != nil {
			log.Fatalln("Failed to create object file: ", err)
		}

		_, err = llvmContext.CompileModule(module, llvm.CodeGenFileType(llvm.ObjectFile), objectFile)
		if err != nil {
			log.Fatalln("Failed to compile module: ", err)
		}

		fileName := filepath.Base(filePath)
		outputFile := replaceFileEnding(fileName, "")

		linkerCmd := exec.Command("gcc", "-v", "-Wall", "-Wextra", "-Wwrite-strings", "-g3", "-o"+outputFile, objectFilePath)
		_, err = linkerCmd.CombinedOutput()
		if err != nil {
			log.Fatalln("Failed to link object file: ", err)
		}

		// TODO: The below should be specifiable with an argument.
		// Move the output file into the same directory as the source code
		outputPath := filepath.Join(filepath.Dir(filePath), outputFile)
		os.Rename(outputFile, outputPath)

		os.Remove(objectFilePath)
	}
}

func replaceFileEnding(filePath string, newEnding string) string {
	dotSplit := strings.Split(filePath, ".")

	if newEnding == "" {
		dotSplit = dotSplit[:len(dotSplit)-1]
	} else {
		dotSplit[len(dotSplit)-1] = newEnding
	}

	return strings.Join(dotSplit, ".")
}
