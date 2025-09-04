package main

import (
	"chogopy/pkg/backend"
	"chogopy/pkg/codegen"
	"chogopy/pkg/lexer"
	"chogopy/pkg/parser"
	"chogopy/pkg/scopes"
	"chogopy/pkg/typechecks"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/kr/pretty"
	"tinygo.org/x/go-llvm"
)

func main() {
	filename := ""

	if len(os.Args) > 2 {
		filename = os.Args[2]
	} else {
		log.Fatal("Please provide a filename and a mode")
	}

	llFileName := replaceFileEnding(filename, "ll")
	objectFileName := replaceFileEnding(filename, "o")

	byteStream, err := os.ReadFile(filename)
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
			pretty.Println(program)
		case "-n":
			program := myParser.ParseProgram()
			assignTargets := scopes.AssignTargets{}
			assignTargets.Analyze(&program)
			scopes := scopes.NameScopes{}
			scopes.Analyze(&program)
			pretty.Println(scopes.NameContext)
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
			pretty.Println(codeGenerator.Module.String())

			// TODO: To keep the test cases working I am only appending .ll to the filename
			// here but will have to change that in the future and modify the test cases accordingly.
			err := os.WriteFile(
				filename+".ll",
				[]byte(codeGenerator.Module.String()),
				0644,
			)
			if err != nil {
				panic(err)
			}
		case "-o":
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
				llFileName,
				[]byte(codeGenerator.Module.String()),
				0644,
			)
			if err != nil {
				log.Fatalln("Failed to create llvm IR file: ", err)
			}

			backend.Init()
			llvmContext := backend.NewllvmContext()
			defer llvmContext.Dispose()

			module := llvmContext.ParseIRFromFile(llFileName)
			os.Remove(llFileName)

			// TODO: Somehow optimizing the module breaks things so I will leave it commented out for now.
			// I believe it might be related to functions that incorrectly handle strings via local allocation.
			// llvmContext.OptimizeModule(module)

			objectFile, err := os.Create(objectFileName)
			if err != nil {
				log.Fatalln("Failed to create object file: ", err)
			}

			_, err = llvmContext.CompileModule(module, llvm.CodeGenFileType(llvm.ObjectFile), objectFile)
			if err != nil {
				log.Fatalln("Failed to write object file: ", err)
			}

			linkerCmd := exec.Command("gcc", "-o "+replaceFileEnding(filename, ""), objectFileName)
			_, err = linkerCmd.CombinedOutput()
			if err != nil {
				log.Fatalln("Failed to link object file: ", err)
			}
		}
	}
}

func replaceFileEnding(filename string, newEnding string) string {
	dotSplit := strings.Split(filename, ".")

	if newEnding == "" {
		dotSplit = dotSplit[:len(dotSplit)-1]
	} else {
		dotSplit[len(dotSplit)-1] = newEnding
	}

	return strings.Join(dotSplit, ".")
}
