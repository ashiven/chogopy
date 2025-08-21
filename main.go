package main

import (
	"chogopy/pkg/astanalysis"
	"chogopy/pkg/lexer"
	"chogopy/pkg/parser"
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
			assignTargets := astanalysis.AssignTargets{}
			assignTargets.Analyze(&program)
		case "-e":
			program := myParser.ParseProgram()
			environments := astanalysis.EnvironmentBuilder{}
			environments.Analyze(&program)
			pretty.Println(environments.LocalEnvironment)
		}
	}
}
