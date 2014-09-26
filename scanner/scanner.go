package scanner

import (
	_ "bytes"
	"errors"
	"fmt"
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

var IntegerZero = errors.New("Integers cannot start with 0")
var MalformedNumber = errors.New("Number is malformed")

type ScannerError struct {
	msg string
	inv string
}

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

	// Newlines and Tabs
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if currentChar == "\n" {
			scanner.advance()
			scanner.line++
			continue
		}

		if currentChar == "\t" {
			scanner.advance()
			continue
		}

		scanner.commit()
		break
	}

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
			if scanner.isReservedWord(lexBuf.String()) {
				return Token{"res", lexBuf.String()}, nil
			} else {
				return Token{"id", lexBuf.String()}, nil
			}
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// Whitespace
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

	scanner.reset()
	lexBuf.Reset()

	// Long Real
	var period bool = true
	var exponent bool = false
	var mustBePeriod bool = false
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.isFirstChar() && !isDigit(currentChar) {
			break
		}

		if scanner.isFirstChar() && currentChar == "0" {
			mustBePeriod = true
			scanner.advance()
			continue
		}

		if currentChar != "." && mustBePeriod == true {
			scanner.advance()
			scanner.commit()
			return Token{}, MalformedNumber
		}

		if currentChar == "." && period == false {
			scanner.advance()
			scanner.commit()
			return Token{}, MalformedNumber
		}

		if currentChar == "E" && exponent == false {
			scanner.advance()
			scanner.commit()
			return Token{}, MalformedNumber
		}

		if isDigit(currentChar) {
			lexBuf.WriteString(currentChar)
			scanner.advance()
		} else if currentChar == "." {
			period = false
			exponent = true
			mustBePeriod = false
			lexBuf.WriteString(currentChar)
			scanner.advance()
		} else if currentChar == "E" {
			exponent = false
			scanner.advance()
		} else {
			scanner.commit()
			return Token{"lgrl", lexBuf.String()}, nil
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// Real
	period = true
	mustBePeriod = false
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.isFirstChar() && !isDigit(currentChar) {
			break
		}

		if scanner.isFirstChar() && currentChar == "0" {
			mustBePeriod = true
			scanner.advance()
			continue
		}

		if currentChar != "." && mustBePeriod == true {
			scanner.advance()
			scanner.commit()
			return Token{}, MalformedNumber
		}

		if currentChar == "." && period == false {
			scanner.advance()
			scanner.commit()
			return Token{}, MalformedNumber
		}

		if isDigit(currentChar) {
			lexBuf.WriteString(currentChar)
			scanner.advance()
		} else if currentChar == "." {
			period = false
			mustBePeriod = false
			lexBuf.WriteString(currentChar)
			scanner.advance()
		} else {
			scanner.commit()
			return Token{"real", lexBuf.String()}, nil
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// Integers
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.isFirstChar() && currentChar == "0" {
			scanner.advance()
			scanner.commit()
			return Token{}, IntegerZero
		}

		if scanner.isFirstChar() && !isDigit(currentChar) {
			break
		}

		if isDigit(currentChar) {
			lexBuf.WriteString(currentChar)
			scanner.advance()
		} else {
			scanner.commit()
			return Token{"int", lexBuf.String()}, nil
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// relop
	var lt bool = false
	var gt bool = false
	var colon bool = false
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if currentChar == "<" && scanner.isFirstChar() {
			scanner.advance()
			lt = true
			lexBuf.WriteString(currentChar)
			continue
		}

		if currentChar == ">" && scanner.isFirstChar() {
			scanner.advance()

			if lt == true {
				scanner.commit()
				return Token{"relop", "<>"}, nil
			} else {
				gt = true
			}

			lexBuf.WriteString(currentChar)
			continue
		}

		if currentChar == ":" && scanner.isFirstChar() {
			scanner.advance()
			colon = true
			lexBuf.WriteString(currentChar)
			continue
		} else if currentChar == "=" {
			scanner.advance()

			if lt == true {
				scanner.commit()
				return Token{"relop", "<="}, nil
			}

			if gt == true {
				scanner.commit()
				return Token{"relop", ">="}, nil
			}

			if colon == true {
				scanner.commit()
				return Token{"relop", ":="}, nil
			}
			break
		}

		if !scanner.isFirstChar() {
			return Token{"relop", currentChar}, nil
		} else {
			break
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// addop
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if currentChar == "+" {
			scanner.advance()
			scanner.commit()
			return Token{"addop", currentChar}, nil
		}

		if currentChar == "-" {
			scanner.advance()
			scanner.commit()
			return Token{"addop", currentChar}, nil
		}

		scanner.advance()
		break
	}

	scanner.reset()
	lexBuf.Reset()

	// mulop
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if currentChar == "*" {
			scanner.advance()
			scanner.commit()
			return Token{"mulop", currentChar}, nil
		}

		if currentChar == "/" {
			scanner.advance()
			scanner.commit()
			return Token{"mulop", currentChar}, nil
		}

		scanner.advance()
		break
	}

	scanner.reset()
	lexBuf.Reset()

	// Other
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if currentChar == "(" {
			scanner.advance()
			scanner.commit()
			return Token{"lp", "("}, nil
		} else if currentChar == ")" {
			scanner.advance()
			scanner.commit()
			return Token{"rp", ")"}, nil
		} else if currentChar == "[" {
			scanner.advance()
			scanner.commit()
			return Token{"lb", "["}, nil
		} else if currentChar == "]" {
			scanner.advance()
			scanner.commit()
			return Token{"rb", "]"}, nil
		} else if currentChar == "," {
			scanner.advance()
			scanner.commit()
			return Token{"com", ","}, nil
		} else if currentChar == ";" {
			scanner.advance()
			scanner.commit()
			return Token{"scol", ";"}, nil
		} else if currentChar == ":" {
			scanner.advance()
			scanner.commit()
			return Token{"col", ":"}, nil
		} else if currentChar == "=" {
			scanner.advance()
			scanner.commit()
			return Token{"eq", "="}, nil
		} else if currentChar == "." {
			scanner.advance()
			scanner.commit()
			return Token{"end", "."}, nil
		} else {
			break
		}
	}

	scanner.reset()
	lexBuf.Reset()

	currentChar, _ := scanner.currentChar()

	scanner.advance()
	scanner.commit()

	return Token{"err", currentChar}, fmt.Errorf("Invalid character: %s", currentChar)
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
