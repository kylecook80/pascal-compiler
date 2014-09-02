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
	args := os.Args
	var file string
	if len(args) > 1 {
		file = args[1]

		data, err := lexer.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		dataSlice := strings.Split(data, "\n")

		var list lexer.ListingFile

		for i, line := range dataSlice {
			if err := list.AddLine(line); err != nil {
				fmt.Println("Error outputting to listing file.")
				return
			}
			if i == 4 {
				if err := list.AddError("Something is wrong"); err != nil {
					fmt.Println("Error outputting to listing file.")
					return
				}
			}
		}

		fileName := lexer.GenerateTimeString(time.Now()) + "_listing_file.txt"
		if err := list.Save(fileName); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("Listing file: " + fileName)
		}
	} else {
		fmt.Println("Please specify a file name.")
	}
}
