package scanner

import (
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
				return Token{RES, NULL, lexBuf.String()}, nil
			} else {
				return Token{ID, NULL, lexBuf.String()}, nil
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
			return Token{WS, NULL, lexBuf.String()}, nil
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// Long Real
	var period bool = true
	var exponent bool = false
	var mustBePeriod bool = false
	var foundExponent bool = false
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.isFirstChar() && !isDigit(currentChar) {
			break
		}

		if scanner.isFirstChar() && currentChar == "0" {
			lexBuf.WriteString(currentChar)
			mustBePeriod = true
			scanner.advance()

			continue
		}

		if currentChar != "." && mustBePeriod == true {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{}, MalformedNumber
		}

		if currentChar == "." && period == false {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{}, MalformedNumber
		}

		if currentChar == "E" && exponent == false {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{}, MalformedNumber
		}

		if isDigit(currentChar) {
			lexBuf.WriteString(currentChar)
			scanner.advance()
		} else if currentChar == "." {
			lexBuf.WriteString(currentChar)
			period = false
			exponent = true
			mustBePeriod = false
			scanner.advance()
		} else if currentChar == "E" {
			lexBuf.WriteString(currentChar)
			exponent = false
			foundExponent = true
			scanner.advance()
		} else if foundExponent {
			scanner.commit()
			return Token{LONG_REAL, NULL, lexBuf.String()}, nil
		} else {
			break
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// Real
	period = true
	mustBePeriod = false
	var foundPeriod bool = false
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.isFirstChar() && !isDigit(currentChar) {
			break
		}

		if scanner.isFirstChar() && currentChar == "0" {
			lexBuf.WriteString(currentChar)
			mustBePeriod = true
			scanner.advance()
			continue
		}

		if currentChar != "." && mustBePeriod == true {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{}, MalformedNumber
		}

		if currentChar == "." && period == false {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{}, MalformedNumber
		}

		if isDigit(currentChar) {
			lexBuf.WriteString(currentChar)
			scanner.advance()
		} else if currentChar == "." {
			lexBuf.WriteString(currentChar)
			period = false
			mustBePeriod = false
			foundPeriod = true
			scanner.advance()
		} else if foundPeriod {
			scanner.commit()
			return Token{REAL, NULL, lexBuf.String()}, nil
		} else {
			break
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
			lexBuf.WriteString(currentChar)
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
			return Token{INT, NULL, lexBuf.String()}, nil
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// assignop
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.isFirstChar() && currentChar == ":" {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			continue
		}

		if lexBuf.String() == ":" && currentChar == "=" {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{ASSIGNOP, NULL, lexBuf.String()}, nil
		}

		break
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
			lexBuf.WriteString(currentChar)
			scanner.advance()
			lt = true
			continue
		}

		if currentChar == ">" && scanner.isFirstChar() {
			lexBuf.WriteString(currentChar)
			scanner.advance()

			if lt == true {
				scanner.commit()
				return Token{RELOP, NOT_EQ, lexBuf.String()}, nil
			} else {
				gt = true
			}

			continue
		}

		if currentChar == ":" && scanner.isFirstChar() {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			colon = true
			continue
		} else if currentChar == "=" {
			lexBuf.WriteString(currentChar)
			scanner.advance()

			if lt == true {
				scanner.commit()
				return Token{RELOP, LESS_EQ, lexBuf.String()}, nil
			}

			if gt == true {
				scanner.commit()
				return Token{RELOP, GREATER_EQ, lexBuf.String()}, nil
			}

			if colon == true {
				scanner.commit()
				return Token{RELOP, EQ, lexBuf.String()}, nil
			}
			break
		}

		if !scanner.isFirstChar() {
			lexBuf.WriteString(currentChar)
			if currentChar == "<" {
				return Token{RELOP, LESS, lexBuf.String()}, nil
			} else {
				return Token{RELOP, GREATER, lexBuf.String()}, nil
			}
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
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{ADDOP, ADD, lexBuf.String()}, nil
		}

		if currentChar == "-" {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{ADDOP, SUB, lexBuf.String()}, nil
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
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{MULOP, MUL, lexBuf.String()}, nil
		}

		if currentChar == "/" {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{MULOP, DIV, lexBuf.String()}, nil
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

		lexBuf.WriteString(currentChar)

		if currentChar == "(" {
			scanner.advance()
			scanner.commit()
			return Token{LEFT_PAREN, NULL, lexBuf.String()}, nil
		} else if currentChar == ")" {
			scanner.advance()
			scanner.commit()
			return Token{RIGHT_PAREN, NULL, lexBuf.String()}, nil
		} else if currentChar == "[" {
			scanner.advance()
			scanner.commit()
			return Token{LEFT_BRACKET, NULL, lexBuf.String()}, nil
		} else if currentChar == "]" {
			scanner.advance()
			scanner.commit()
			return Token{RIGHT_BRACKET, NULL, lexBuf.String()}, nil
		} else if currentChar == "," {
			scanner.advance()
			scanner.commit()
			return Token{COMMA, NULL, lexBuf.String()}, nil
		} else if currentChar == ";" {
			scanner.advance()
			scanner.commit()
			return Token{SEMI, NULL, lexBuf.String()}, nil
		} else if currentChar == ":" {
			scanner.advance()
			scanner.commit()
			return Token{COLON, NULL, lexBuf.String()}, nil
		} else if currentChar == "." {
			scanner.advance()
			scanner.commit()
			return Token{END, NULL, lexBuf.String()}, nil
		} else {
			break
		}
	}

	scanner.reset()
	lexBuf.Reset()

	currentChar, _ := scanner.currentChar()

	scanner.advance()
	scanner.commit()

	return Token{LEXERR, UNREC, lexBuf.String()}, fmt.Errorf("Invalid character: %s", currentChar)
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
