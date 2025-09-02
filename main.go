package main

import (
	"chogopy/pkg/codegen"
	"chogopy/pkg/lexer"
	"chogopy/pkg/parser"
	"chogopy/pkg/scopes"
	"chogopy/pkg/typechecks"
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
