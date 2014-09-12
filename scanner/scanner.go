package scanner

import (
	"bytes"
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

type Scanner struct {
	line int
	pos  int
	buf  [][]byte
}

type Buffer struct {
	buf []byte
}

type Token struct {
	id   string
	attr string
}

func NewScanner() *Scanner {
	return &Scanner{0, 0, nil}
}

// ReadFile takes a file and reads it into memory.
// It is then returned as a string.
func (scanner *Scanner) ReadFile(file string) {
	openFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer openFile.Close()

	readBuf := make([]byte, 1024)
	fileBuf := make([]byte, 0)
	for {
		n, err := openFile.Read(readBuf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}
		fileBuf = append(fileBuf, readBuf...)
	}
	scanner.buf = bytes.Split(fileBuf, []byte("\n"))
}

func (scanner *Scanner) GetNextToken() Token {
	nextLexeme := scanner.getNextLexeme()
	if nextLexeme != nil {
		fmt.Printf("%s\n", nextLexeme)
	}
	return Token{"catchall", ""}
}

func (scanner *Scanner) getNextLexeme() []byte {
	var lexBuf Buffer

	for {
		nextChar := scanner.peekNextChar()

		// No more characters
		if nextChar == byte(0) {
			scanner.advance()
			return lexBuf.bytes()
		}

		// 0 == null, 9 == \t, 10 == \n, 32 == Space, 59 = ;
		if nextChar == byte(9) || nextChar == byte(10) || nextChar == byte(32) || nextChar == byte(59) {
			scanner.advance()
			break
		}

		scanner.advance()
		lexBuf.add(nextChar)
	}

	return lexBuf.bytes()
}

func (scanner *Scanner) advance() {
	nextChar := scanner.peekNextChar()
	if nextChar == 0 && scanner.line < len(scanner.buf) {
		scanner.line++
		scanner.pos = 0
	} else if scanner.line > len(scanner.buf) {
		return
	} else {
		scanner.pos++
	}
}

func (buffer *Buffer) add(obj ...byte) {
	buffer.buf = append(buffer.buf, obj...)
}

func (buffer *Buffer) bytes() []byte {
	return buffer.buf
}

func (scanner *Scanner) peekNextChar() byte {
	var nextChar byte

	if scanner.line >= len(scanner.buf) || scanner.pos >= len(scanner.buf[scanner.line]) {
		return 0
	}

	nextChar = scanner.buf[scanner.line][scanner.pos]
	return nextChar
}

func isIdentifier(lexeme []byte) bool {
	return true
}
