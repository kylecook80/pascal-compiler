package util

import (
	"os"
	"strconv"
	"strings"
	_ "time"
)

// listingFile is a structure for the creation and saving of
// a source code file during lexical analysis.
type ListingFile struct {
	Buffer
	counter int
}

func NewListingFile() *ListingFile {
	return new(ListingFile)
}

// AddLine adds a line from the source code to the listing file.
// It adds a line number at the beginning.
func (listing *ListingFile) AddLine(line string) error {
	lineNumber := strconv.Itoa(listing.counter + 1)
	_, err := listing.WriteString(lineNumber + ": " + strings.Trim(line, "0x00") + "\n")
	listing.counter += 1
	return err
}

// AddError adds a line to the listing file describing an error.
// It adds "LEXERR" to the front of the error.
func (listing *ListingFile) AddLexError(line string) error {
	_, err := listing.WriteString("Lexical Error: " + line + "\n")
	return err
}

func (listing *ListingFile) AddSyntaxError(line string) error {
	_, err := listing.WriteString("Syntax Error: " + line + "\n")
	return err
}

func (listing *ListingFile) AddTypeError(line string) error {
	_, err := listing.WriteString("Type Error: " + line + "\n")
	return err
}

func (listing *ListingFile) AddScopeError(line string) error {
	_, err := listing.WriteString("Scope Error: " + line + "\n")
	return err
}

func (listing *ListingFile) LineCount() int {
	return listing.counter
}

// Save takes a filename as a string and saves the file
// to the file system. It saves to the same directory as
// the lexer is called from.
func (listing *ListingFile) Save() string {
	file := "listing_file.txt"
	newFile, err := os.Create(file)
	if err != nil {
		return ""
	}
	defer newFile.Close()

	newFile.Write(listing.Bytes())
	return file
}
