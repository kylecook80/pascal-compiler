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

// var MalformedNumber = errors.New("Number is malformed")
var LengthError = errors.New("Identifier or Number is too long")

type ScannerError struct {
	msg string
	inv string
}

type Scanner struct {
	line       int
	lineLength int
	posF       int
	posB       int
	buf        util.Buffer
	res        util.Buffer
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
		// fmt.Println("Newlines")
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
		// fmt.Println("IDs")
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.currentLength() > idLength {
			return Token{}, LengthError
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
				str := lexBuf.String()
				return Token{ID, NULL, str}, nil
			}
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// Whitespace
	for {
		// fmt.Println("WS")
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
	for {
		// fmt.Println("Long Real")
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.currentLength() > idLength {
			return Token{}, LengthError
		}

		if scanner.isFirstChar() && !isDigit(currentChar) {
			break
		}

		if scanner.isFirstChar() && currentChar == "0" {
			if scanner.peek() == "." {
				lexBuf.WriteString(currentChar)
				scanner.advance()
				continue
			} else {
				break
			}
		}

		if !scanner.isFirstChar() && currentChar == "." {
			if isDigit(scanner.peek()) {
				lexBuf.WriteString(currentChar)
				scanner.advance()
				continue
			} else {
				break
			}
		}

		if !scanner.isFirstChar() && currentChar == "E" {
			if isDigit(scanner.peek()) {
				lexBuf.WriteString(currentChar)
				scanner.advance()
				continue
			} else {
				break
			}
		}

		if isDigit(currentChar) {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			continue
		} else {
			scanner.commit()
			return Token{LONG_REAL, NULL, lexBuf.String()}, nil
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// Real
	for {
		// fmt.Println("Real")
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.currentLength() > idLength {
			return Token{}, LengthError
		}

		if scanner.isFirstChar() && !isDigit(currentChar) {
			break
		}

		if scanner.isFirstChar() && currentChar == "0" {
			if scanner.peek() == "." {
				lexBuf.WriteString(currentChar)
				scanner.advance()
				continue
			} else {
				break
			}
		}

		if !scanner.isFirstChar() && currentChar == "." {
			if isDigit(scanner.peek()) {
				lexBuf.WriteString(currentChar)
				scanner.advance()
				continue
			} else {
				break
			}
		}

		if isDigit(currentChar) {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			continue
		} else {
			scanner.commit()
			return Token{REAL, NULL, lexBuf.String()}, nil
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// Integers
	for {
		// fmt.Println("Integers")
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.currentLength() > idLength {
			return Token{}, LengthError
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
		// fmt.Println("assignop")
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
	for {
		// fmt.Println("relop")
		currentChar, err := scanner.currentChar()
		if err != nil {
			return Token{}, err
		}

		if scanner.isFirstChar() && currentChar == "<" {
			switch scanner.peek() {
			case ">":
				lexBuf.WriteString(currentChar + ">")
				scanner.advance()
				scanner.commit()
				return Token{RELOP, NOT_EQ, lexBuf.String()}, nil
			case "=":
				lexBuf.WriteString(currentChar + "=")
				scanner.advance()
				scanner.commit()
				return Token{RELOP, LESS_EQ, lexBuf.String()}, nil
			default:
				lexBuf.WriteString(currentChar)
				scanner.advance()
				scanner.commit()
				return Token{RELOP, LESS, lexBuf.String()}, nil
			}
		}

		if scanner.isFirstChar() && currentChar == ">" {
			if scanner.peek() == "=" {
				scanner.advance()
				scanner.commit()
				return Token{RELOP, GREATER_EQ, lexBuf.String()}, nil
			} else {
				scanner.advance()
				scanner.commit()
				return Token{RELOP, GREATER, lexBuf.String()}, nil
			}
		}

		if scanner.isFirstChar() && currentChar == "=" {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return Token{RELOP, EQ, lexBuf.String()}, nil
		}

		break
	}

	scanner.reset()
	lexBuf.Reset()

	// addop
	for {
		// fmt.Println("addop")
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

		break
	}

	scanner.reset()
	lexBuf.Reset()

	// mulop
	for {
		// fmt.Println("mulop")
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

		break
	}

	scanner.reset()
	lexBuf.Reset()

	// Other
	for {
		// fmt.Println("other")
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

func (scanner *Scanner) peek() string {
	scanner.advance()
	str, _ := scanner.currentChar()
	scanner.retract()
	return str
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
