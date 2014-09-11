package lexer

import (
	// "bytes"
	"fmt"
	"io"
	"os"
)

// Define constants and errors
const (
	lineLength = 72
	idLength   = 10
)

var lineLengthError = fmt.Sprintf("The length of the line exceeds %d characters.", lineLength)
var identifierLengthError = fmt.Sprintf("The length of the identifier exceeds %d characters.", idLength)

type Lexer struct {
	line int
	pos  int
	buf  [][]byte
}

type Token struct {
	id   string
	attr string
}

func NewLexer() Lexer {
	return Lexer{0, 0, nil}
}

// ReadFile takes a file and reads it into memory.
// It is then returned as a string.
func (lexer Lexer) ReadFile(file string) {
	openFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer openFile.Close()

	var fileBuf []byte
	for {
		n, err := openFile.Read(fileBuf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}
		// lexer.buf = bytes.Split(fileBuf, []byte("\n"))
	}

	fmt.Println(fileBuf)
}

func (lexer Lexer) GetNextToken() Token {
	nextLexeme := lexer.getNextLexeme()
	fmt.Println(nextLexeme)

	// if isIdentifier(nextLexeme) {
	// 	return Token{"id", "symbol"}
	// }

	return Token{"catchall", ""}
}

func (lexer Lexer) getNextLexeme() []byte {
	var nextChar byte
	var tokenBuf []byte
	fmt.Println(lexer.buf)
	for string(nextChar) != " " {
		nextChar = lexer.buf[lexer.line][lexer.pos]
		tokenBuf = append(tokenBuf, nextChar)
	}
	return tokenBuf
}

func isIdentifier(lexeme []byte) bool {
	return true
}

// func (lexer Lexer) Begin(file string) error {
// 	data, err := ReadFile(file)
// 	if err != nil {
// 		return err
// 	}

// 	dataSlice := strings.Split(data, "\n")
// 	var list listingFile

// 	// Generate listing file.
// 	fileName := GenerateTimeString(time.Now()) + "_listing_file.txt"
// 	if err := list.Save(fileName); err != nil {
// 		return err
// 	} else {
// 		fmt.Println("Listing file: " + fileName)
// 	}

// 	lexer.tokens = tokens
// 	return nil // No errors
// }

// func (lexer Lexer) Lex(line string) ([]Token, error) {
// 	lineBytes = []bytes(line)
// 	hCounter = 0
// 	tCounter = 0

// 	return make([]Token, 1), nil
// }
