package main

import (
	"chogopy/pkg/lexer"
	"fmt"
	"os"
)

func main() {
	stream, err := os.ReadFile("test.txt")
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(stream))

	scanner := lexer.NewScanner(string(stream))
	fmt.Println(scanner.Consume())
}
