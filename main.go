package main

import (
	"chogopy/pkg/codegen"
	"chogopy/pkg/lexer"
	"chogopy/pkg/namescopes"
	"chogopy/pkg/parser"
	"chogopy/pkg/typechecking"
	"fmt"
	"log"
	"os"

	"github.com/kr/pretty"
)

func main() {
	filename := ""
	if len(os.Args) > 2 {
		filename = os.Args[2]
	} else {
		log.Fatal("Please provide a filename and a mode")
	}

	byteStream, err := os.ReadFile(filename)
	if err != nil {
		pretty.Println(err.Error())
	}
	stream := string(byteStream)
	pretty.Println(stream)

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
		case "-v":
			program := myParser.ParseProgram()
			assignTargets := namescopes.AssignTargets{}
			assignTargets.Analyze(&program)
		case "-e":
			program := myParser.ParseProgram()
			environmentBuilder := typechecking.EnvironmentBuilder{}
			environmentBuilder.Build(&program)
			pretty.Println(environmentBuilder.LocalEnv)
		case "-t":
			program := myParser.ParseProgram()
			staticTyping := typechecking.StaticTyping{}
			staticTyping.Analyze(&program)
			pretty.Println(program)
		case "-n":
			program := myParser.ParseProgram()
			nameScopes := namescopes.NameScopes{}
			nameScopes.Analyze(&program)
			pretty.Println(nameScopes.NameContext)
		case "-c":
			program := myParser.ParseProgram()
			staticTyping := typechecking.StaticTyping{}
			staticTyping.Analyze(&program)
			codeGenerator := codegen.CodeGenerator{}
			codeGenerator.Generate(&program)
			pretty.Println(codeGenerator.Module.String())

			err := os.WriteFile(
				fmt.Sprintf("%s.ll", filename),
				[]byte(codeGenerator.Module.String()),
				0644,
			)
			if err != nil {
				panic(err)
			}
		}
	}
}
