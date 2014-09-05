package lexer

import (
	"bytes"
	"error"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type Lexer struct{}
type Token struct{}

func NewLexer() *Lexer {
	return new(Lexer)
}

func (lexer *Lexer) Begin(file string) error {
	data, err := ReadFile(file)
	if err != nil {
		return err
	}

	dataSlice := strings.Split(data, "\n")
	var list listingFile

	// Main for loop that iterates over lines.
	for _, line := range dataSlice {
		tokens, err := lexer.Lex(line)
		if err != nil {
			return errors.New("Did not lex correctly.")
		}
		fmt.Println(tokens)
	}

	// Generate listing file.
	fileName := GenerateTimeString(time.Now()) + "_listing_file.txt"
	if err := list.Save(fileName); err != nil {
		return errors.New("Did not lex correctly.")
	} else {
		fmt.Println("Listing file: " + fileName)
	}

	return nil // No errors
}

func (lexer *Lexer) Lex(line string) ([]Token, error) {
	return make([]Token, 1), nil
}

// listingFile is a structure for the creation and saving of
// a source code file during lexical analysis.
type listingFile struct {
	buf     bytes.Buffer
	counter int
}

// AddLine adds a line from the source code to the listing file.
// It adds a line number at the beginning.
func (listing *listingFile) AddLine(line string) error {
	lineNumber := strconv.Itoa(listing.counter + 1)
	_, err := listing.buf.WriteString(lineNumber + ": " + line + "\n")
	listing.counter += 1
	return err
}

// AddError adds a line to the listing file describing an error.
// It adds "LEXERR" to the front of the error.
func (listing *listingFile) AddError(line string) error {
	_, err := listing.buf.WriteString("LEXERR: " + line + "\n")
	return err
}

// Save takes a filename as a string and saves the file
// to the file system. It saves to the same directory as
// the lexer is called from.
func (listing *listingFile) Save(file string) error {
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
func GenerateTimeString(t time.Time) string {
	formattedTime := t.Format("2006-01-2-15-04-05")
	underscoreTime := strings.Replace(formattedTime, "-", "_", -1)
	return underscoreTime
}

// ReadFile takes a file and reads it into memory.
// It is then returned as a string.
func ReadFile(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	return string(data), err
}
