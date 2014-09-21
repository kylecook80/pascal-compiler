package scanner

import (
	_ "bytes"
	_ "fmt"
	"io"
	"strconv"
	"strings"
)

import "compiler/util"

// Define constants and errors
const (
	lineLength = 72
	idLength   = 10
)

type Scanner struct {
	line int
	posF int
	posB int
	buf  util.Buffer
	res  util.Buffer
}

type Token struct {
	id   string
	attr string
}

func NewScanner() *Scanner {
	return new(Scanner)
}

func (scanner *Scanner) Buffer() *util.Buffer {
	return &scanner.buf
}

// ReadFile takes a file and reads it into memory.
// It is then returned as a string.
func (scanner *Scanner) ReadSourceFile(file string) {
	buf := util.ReadFile(file)
	scanner.buf = *buf
}

func (scanner *Scanner) ReadReservedFile(file string) {
	buf := util.ReadFile(file)
	scanner.res = *buf
}

func (scanner *Scanner) CurrentLineNumber() int {
	return scanner.line
}

func (scanner *Scanner) NextToken() (Token, error) {
	lexBuf := new(util.Buffer)

	// IDs / Reserved Words
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.isFirstChar() && !isChar(currentChar) {
			break
		}

		if isChar(currentChar) || isDigit(currentChar) {
			lexBuf.WriteString(currentChar)
			scanner.advance()
		} else {
			scanner.commit()
			return Token{"idres", lexBuf.String()}, nil
		}
	}

	scanner.commit()
	lexBuf.Reset()

	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.isFirstChar() && !isWhitespace(currentChar) {
			break
		}

		if isWhitespace(currentChar) {
			scanner.advance()
		} else {
			scanner.commit()
			return Token{"ws", ""}, nil
		}
	}

	scanner.commit()
	lexBuf.Reset()

	for {

	}

	return Token{lexBuf.String(), ""}, nil
}

func isChar(char string) bool {
	if !isOneChar(char) {
		return false
	}

	if ("a" <= char && char <= "z") || ("A" <= char && char <= "Z") {
		return true
	} else {
		return false
	}
}

func isDigit(char string) bool {
	if !isOneChar(char) {
		return false
	}

	charDigit, err := strconv.Atoi(char)
	if err != nil {
		return false
	}

	if 0 <= charDigit && charDigit <= 9 {
		return true
	} else {
		return false
	}
}

func isWhitespace(char string) bool {
	if !isOneChar(char) {
		return false
	}

	if char == " " {
		return true
	} else {
		return false
	}
}

func isOneChar(char string) bool {
	if len(char) == 1 {
		return true
	} else {
		return false
	}
}

func (scanner *Scanner) currentChar() (string, error) {
	if scanner.posF >= len(scanner.buf.Bytes()) {
		return "", io.EOF
	}

	character, err := scanner.Buffer().ReadAt(scanner.posF)
	if err != nil {
		panic(err)
	}

	return string(character), nil
}

func (scanner *Scanner) advance() {
	scanner.posF++
}

func (scanner *Scanner) retract() {
	scanner.posF--
}

func (scanner *Scanner) commit() {
	scanner.posB = scanner.posF
}

func (scanner *Scanner) reset() {
	scanner.posF = scanner.posB
}

func (scanner *Scanner) isFirstChar() bool {
	if scanner.currentLength() < 1 {
		return true
	} else {
		return false
	}
}

func (scanner *Scanner) currentLength() int {
	return scanner.posF - scanner.posB
}

func (scanner *Scanner) isReservedWord(word string) bool {
	reservedString := scanner.res.String()
	reservedWords := strings.Split(reservedString, "\n")

	for _, resWord := range reservedWords {
		if resWord == word {
			return true
		}
	}

	return false
}

// func (scanner *Scanner) NextToken() (Token, error) {
// 	lexBuf := new(bytes.Buffer)
// 	for {
// 		currentChar, err := scanner.currentChar()

// 		if err != nil {
// 			return Token{}, err
// 		}

// 		// No more characters
// 		if []byte(currentChar)[0] == byte(0) {
// 			return Token{}, io.EOF
// 		} else {
// 			if currentChar == "\n" {
// 				scanner.line++
// 				break
// 			}

// 			if currentChar == "\t" || currentChar == " " {
// 				break
// 			}

// 			lexeme := string(lexBuf.Bytes())
// 			for _, word := range scanner.ReservedWords() {
// 				if lexeme == word {
// 					return Token{"res", word}, nil
// 				}
// 			}

// 			switch currentChar {
// 			case "(":
// 				scanner.advance()
// 				return Token{"lp", string(currentChar)}, nil
// 			case ")":
// 				scanner.advance()
// 				return Token{"rp", string(currentChar)}, nil
// 			case ":":
// 				scanner.advance()
// 				return Token{"col", string(currentChar)}, nil
// 			case ";":
// 				scanner.advance()
// 				return Token{"scol", string(currentChar)}, nil
// 			case ",":
// 				scanner.advance()
// 				return Token{"com", string(currentChar)}, nil
// 			}

// 			lexBuf.WriteString(currentChar)
// 		}
// 		scanner.advance()
// 	}

// 	scanner.advance()
// 	scanner.commit()

// 	return Token{}, fmt.Errorf("Unknown symbol")
// }
