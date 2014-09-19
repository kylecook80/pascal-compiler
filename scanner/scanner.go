package scanner

import (
	"bytes"
	"errors"
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
	posF int
	posB int
	buf  Buffer
}

type Token struct {
	id   string
	attr string
}

func NewScanner() *Scanner {
	return &Scanner{0, 0, 0, Buffer{}}
}

func (scanner *Scanner) Buffer() *Buffer {
	return &scanner.buf
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
	fileBuf := new(Buffer)
	for {
		n, err := openFile.Read(readBuf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}
		fileBuf.add(readBuf...)
	}
	scanner.buf = *fileBuf
}

func (scanner *Scanner) GetNextToken() (Token, error) {
	lexBuf := new(Buffer)
	for {
		currentChar, _ := scanner.currentChar()

		// fmt.Printf("currentChar: %d\n", currentChar)
		// fmt.Printf("posF: %d\n", scanner.posF)

		// No more characters
		if nextChar == byte(0) {
			scanner.advance()
			return lexBuf.bytes()
		}

		if currentChar == byte(0) {
			break
		} else {
			// 9 == \t, 10 == \n, 32 == Space, 59 = ;
			if currentChar == byte(9) || currentChar == byte(10) || currentChar == byte(32) || currentChar == byte(59) {
				scanner.advance()
				break
			}

			switch currentChar {
			case byte(40):
				scanner.advance()
				return Token{"lp", string(currentChar)}, nil
			case byte(41):
				scanner.advance()
				return Token{"rp", string(currentChar)}, nil
			case byte(39):
				scanner.advance()
				return Token{"qt", string(currentChar)}, nil
			}

			lexBuf.add(currentChar)
		}

		scanner.advance()
	}

	scanner.advance()
	scanner.commit()

	if bytes.Equal(lexBuf.bytes(), []byte("program")) {
		return Token{"res", ""}, nil
	}

	return Token{}, fmt.Errorf("LEXERR: Unknown symbol %s", lexBuf.bytes())
}

func (scanner *Scanner) peekNextChar() (byte, error) {
	nextChar := scanner.Buffer().At(scanner.posF + 1)
	return nextChar, nil
}

func (scanner *Scanner) currentChar() (byte, error) {
	if scanner.posF > len(scanner.buf.bytes()) {
		return 0, errors.New("End of File Lex Error.")
	}

	return scanner.Buffer().At(scanner.posF), nil
}

func (scanner *Scanner) advance() {
	scanner.posF++
}

func (scanner *Scanner) commit() {
	scanner.posB = scanner.posF
}

type Buffer struct {
	buf []byte
}

func (buffer *Buffer) At(idx int) byte {
	if idx < len(buffer.buf) {
		return buffer.buf[idx]
	}

	return 0
}

func (buffer *Buffer) add(obj ...byte) {
	buffer.buf = append(buffer.buf, obj...)
}

func (buffer *Buffer) bytes() []byte {
	return buffer.buf
}

func isIdentifier(lexeme []byte) bool {
	return true
}
