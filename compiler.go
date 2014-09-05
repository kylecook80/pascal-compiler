package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

import "compiler/lexer"

func main() {
	// Get the arguments passed to the compiler
	args := os.Args

	if len(args) > 1 {
		file := args[1]
		lexer := lexer.NewLexer()
		lexer.Begin(file)
	} else {
		fmt.Println("Please specify a file name.")
	}
}
