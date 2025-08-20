package main

import (
	"chogopy/pkg/astanalysis"
	"chogopy/pkg/lexer"
	"chogopy/pkg/parser"
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
		fmt.Println(err.Error())
	}
	stream := string(byteStream)
	fmt.Println(stream)

	myLexer := lexer.NewLexer(stream)
	myParser := parser.NewParser(&myLexer)

	if len(os.Args) > 2 {
		switch os.Args[1] {
		case "l":
			token := myLexer.Consume(false)
			for token.Kind != lexer.EOF {
				fmt.Printf("%# v\n", pretty.Formatter(token))
				token = myLexer.Consume(false)
			}
			fmt.Printf("%# v\n", pretty.Formatter(token))
		case "-p":
			program := myParser.ParseProgram()
			fmt.Printf("%# v\n", pretty.Formatter(program))
		case "-v":
			program := myParser.ParseProgram()
			assignTargetVisitor := astanalysis.AssignTargetVisitor{}
			assignTargetVisitor.Analyze(&program)
		}
	}
}
