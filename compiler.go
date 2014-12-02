package main

import (
	"fmt"
	"os"
)

import scan "compiler/scanner"
import parse "compiler/parser"

func main() {
	// Get the arguments passed to the compiler
	args := os.Args
	file := args[1]

	if len(args) > 1 {
		scanner := scan.NewScanner()
		scanner.ReadReservedFile("scanner/reserved_words.list")
		scanner.ReadSourceFile(file)

		parser := parse.NewParser(scanner)
		parser.Begin(file)
	} else {
		fmt.Println("Please specify a file name.")
	}
}
