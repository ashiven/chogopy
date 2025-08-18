package main

import (
	"chogopy/pkg/lexer"
	"chogopy/pkg/parser"
	"fmt"
	"os"
)

func main() {
	filename := "test.choc"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	byteStream, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
	}
	stream := string(byteStream)

	fmt.Println(stream)

	lexer := lexer.NewLexer(stream)
	parser := parser.NewParser(&lexer)

	program := parser.ParseProgram()

	fmt.Println(program)
	fmt.Println(program.Definitions)
	fmt.Printf("%#v\n", program)
}
