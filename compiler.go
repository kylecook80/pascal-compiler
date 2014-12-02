package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

import scan "compiler/scanner"
import parse "compiler/parser"
import "compiler/util"

func main() {
	// Get the arguments passed to the compiler
	args := os.Args
	// symbolTable := make(map[string]interface{})

	if len(args) > 1 {
		file := args[1]

		listing := util.NewListingFile()
		source := util.ReadFile(file)

		scanner := scan.NewScanner()
		scanner.ReadReservedFile("scanner/reserved_words.list")
		scanner.ReadSourceFile(file)

		tokenFile := []byte{}

		for {
			if scanner.CurrentLineNumber() >= listing.LineCount() {
				listing.AddLine(source.ReadLine(scanner.CurrentLineNumber()))
			}

			token, err := scanner.NextToken()
			// fmt.Println(token)
			// if token.Type() == scan.ID && token.Value()

			if err == io.EOF {
				break
			} else if err != nil {
				listing.AddError(err.Error())
			} else {
				line := scanner.CurrentLineNumber() + 1
				if token.Type() != scan.WS {
					tokenFile = append(tokenFile, []byte(strconv.Itoa(line)+": "+token.String()+"\n")...)
				}
			}
		}

		ioutil.WriteFile(util.GenerateTimeString(time.Now())+"_token_file.txt", tokenFile, 0644)
		listing.Save()
	} else {
		fmt.Println("Please specify a file name.")
	}
}
