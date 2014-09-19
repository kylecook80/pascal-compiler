package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

import "compiler/scanner"

// listingFile is a structure for the creation and saving of
// a source code file during lexical analysis.
type listingFile struct {
	buf     bytes.Buffer
	counter int
}

// AddLine adds a line from the source code to the listing file.
// It adds a line number at the beginning.
func (listing *listingFile) addLine(line string) error {
	lineNumber := strconv.Itoa(listing.counter + 1)
	_, err := listing.buf.WriteString(lineNumber + ": " + line + "\n")
	listing.counter += 1
	return err
}

// AddError adds a line to the listing file describing an error.
// It adds "LEXERR" to the front of the error.
func (listing *listingFile) addError(line string) error {
	_, err := listing.buf.WriteString("LEXERR: " + line + "\n")
	return err
}

// Save takes a filename as a string and saves the file
// to the file system. It saves to the same directory as
// the lexer is called from.
func (listing *listingFile) Save() error {
	file := generateTimeString(time.Now())
	newFile, err := os.Create(file)
	if err != nil {
		return err
	}
	defer newFile.Close()

	newFile.Write(listing.buf.Bytes())
	return nil
}

// GenerateTimeString takes a time and formats it as an underscored
// string, suitable for a filename.
func generateTimeString(t time.Time) string {
	formattedTime := t.Format("2006-01-2-15-04-05")
	underscoreTime := strings.Replace(formattedTime, "-", "_", -1)
	return underscoreTime
}

func main() {
	// Get the arguments passed to the compiler
	args := os.Args

	if len(args) > 1 {
		file := args[1]
		scanner := scanner.NewScanner()
		scanner.ReadFile(file)
		for i := 0; i < 20; i++ {
			// scanner.GetNextToken()
			fmt.Println(scanner.GetNextToken())
		}
	} else {
		fmt.Println("Please specify a file name.")
	}
}
