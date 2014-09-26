package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

import scan "compiler/scanner"
import "compiler/util"

func main() {
	// Get the arguments passed to the compiler
	args := os.Args

	if len(args) > 1 {
		file := args[1]
		// data, _ := ioutil.ReadFile(file)

		// fmt.Printf("%q", data)

		listing := util.NewListingFile()
		source := util.ReadFile(file)

		scanner := scan.NewScanner()
		scanner.ReadReservedFile("scanner/reserved_words.list")
		scanner.ReadSourceFile(file)

		for {
			if scanner.CurrentLineNumber() >= listing.LineCount() {
				listing.AddLine(source.ReadLine(scanner.CurrentLineNumber()))
			}

			token, err := scanner.NextToken()

			if err == io.EOF {
				break
			} else if err != nil {
				listing.AddError(err.Error())
				fmt.Println(err)
			} else {
				line := scanner.CurrentLineNumber() + 1
				fmt.Println(strconv.Itoa(line) + ": " + token.String())
			}
		}

		listing.Save()
	} else {
		fmt.Println("Please specify a file name.")
	}
}
