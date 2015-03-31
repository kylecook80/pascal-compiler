package scanner

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

import . "compiler/util"

// Define constants and errors
const (
	lineLength = 72
	idLength   = 10
)

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
	buf        Buffer
	res        Buffer
	symTable   *SymbolTable
}

func NewScanner() *Scanner {
	scanner := Scanner{symTable: NewSymbolTable()}
	return &scanner
}

func (scanner *Scanner) Buffer() *Buffer {
	return &scanner.buf
}

func (scanner *Scanner) SymbolTable() *SymbolTable {
	return scanner.symTable
}

// ReadFile takes a file and reads it into memory.
// It is then returned as a string.
func (scanner *Scanner) ReadSourceFile(file string) {
	buf := ReadFile(file)
	scanner.buf = *buf
}

func (scanner *Scanner) ReadReservedFile(file string) {
	buf := ReadFile(file)
	scanner.res = *buf
}

func (scanner *Scanner) CurrentLineNumber() int {
	return scanner.line
}

func (scanner *Scanner) NextToken() (Token, error) {
	lexBuf := new(Buffer)

	// Newlines and Tabs
	for {
		currentChar, err := scanner.currentChar()
		if err != nil {
			if err == io.EOF {
				return NewToken(EOF, NULL, ""), nil
			}
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
				resToken := scanner.checkReservedWord(lexBuf.String())
				return NewToken(RES, resToken, lexBuf.String()), nil
			} else {
				str := lexBuf.String()

				sym := NewSymbol(str)
				add := scanner.symTable.AddSymbol(sym)
				scanner.symTable.Print()
				if add == false {
					fmt.Errorf("Error adding to symbol table")
				}

				return NewToken(ID, NULL, str), nil
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
			return NewToken(WS, NULL, lexBuf.String()), nil
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// Long Real
	for {
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
			if strings.Contains(lexBuf.String(), ".") && strings.Contains(lexBuf.String(), "E") {
				scanner.commit()
				return NewToken(NUM, LONG_REAL, lexBuf.String()), nil
			} else {
				break
			}
		}
	}

	scanner.reset()
	lexBuf.Reset()

	// Real
	for {
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
			if strings.Contains(lexBuf.String(), ".") {
				scanner.commit()
				return NewToken(NUM, REAL, lexBuf.String()), nil
			} else {
				break
			}
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
			return NewToken(NUM, INT, lexBuf.String()), nil
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
			return NewToken(ASSIGNOP, NULL, lexBuf.String()), nil
		}

		break
	}

	scanner.reset()
	lexBuf.Reset()

	// relop
	for {
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
				return NewToken(RELOP, NOT_EQ, lexBuf.String()), nil
			case "=":
				lexBuf.WriteString(currentChar + "=")
				scanner.advance()
				scanner.commit()
				return NewToken(RELOP, LESS_EQ, lexBuf.String()), nil
			default:
				lexBuf.WriteString(currentChar)
				scanner.advance()
				scanner.commit()
				return NewToken(RELOP, LESS, lexBuf.String()), nil
			}
		}

		if scanner.isFirstChar() && currentChar == ">" {
			if scanner.peek() == "=" {
				scanner.advance()
				scanner.commit()
				return NewToken(RELOP, GREATER_EQ, lexBuf.String()), nil
			} else {
				scanner.advance()
				scanner.commit()
				return NewToken(RELOP, GREATER, lexBuf.String()), nil
			}
		}

		if scanner.isFirstChar() && currentChar == "=" {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return NewToken(RELOP, EQ, lexBuf.String()), nil
		}

		break
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
			return NewToken(ADDOP, ADD, lexBuf.String()), nil
		}

		if currentChar == "-" {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return NewToken(ADDOP, SUB, lexBuf.String()), nil
		}

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
			return NewToken(MULOP, MUL, lexBuf.String()), nil
		}

		if currentChar == "/" {
			lexBuf.WriteString(currentChar)
			scanner.advance()
			scanner.commit()
			return NewToken(MULOP, DIV, lexBuf.String()), nil
		}

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
			return NewToken(RES, LEFT_PAREN, lexBuf.String()), nil
		} else if currentChar == ")" {
			scanner.advance()
			scanner.commit()
			return NewToken(RES, RIGHT_PAREN, lexBuf.String()), nil
		} else if currentChar == "[" {
			scanner.advance()
			scanner.commit()
			return NewToken(RES, LEFT_BRACKET, lexBuf.String()), nil
		} else if currentChar == "]" {
			scanner.advance()
			scanner.commit()
			return NewToken(RES, RIGHT_BRACKET, lexBuf.String()), nil
		} else if currentChar == "," {
			scanner.advance()
			scanner.commit()
			return NewToken(RES, COMMA, lexBuf.String()), nil
		} else if currentChar == ";" {
			scanner.advance()
			scanner.commit()
			return NewToken(RES, SEMI, lexBuf.String()), nil
		} else if currentChar == ":" {
			scanner.advance()
			scanner.commit()
			return NewToken(RES, COLON, lexBuf.String()), nil
		} else if currentChar == "." {
			scanner.advance()
			scanner.commit()
			if scanner.peek() == "." {
				scanner.advance()
				scanner.commit()
				return NewToken(RANGE, NULL, lexBuf.String()), nil
			} else {
				return NewToken(RES, END, lexBuf.String()), nil
			}
		} else {
			break
		}
	}

	scanner.reset()
	lexBuf.Reset()

	currentChar, _ := scanner.currentChar()

	scanner.advance()
	scanner.commit()

	return NewToken(LEXERR, UNREC, lexBuf.String()), fmt.Errorf("Invalid character: %s", currentChar)
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

func (scanner *Scanner) checkReservedWord(word string) AttributeType {
	if word == "program" {
		return PROG
	}

	if word == "var" {
		return VAR
	}

	if word == "of" {
		return OF
	}

	if word == "integer" {
		return INT_DEC
	}

	if word == "real" {
		return REAL_DEC
	}

	if word == "array" {
		return ARRAY
	}

	if word == "procedure" {
		return PROC
	}

	if word == "begin" {
		return BEGIN
	}

	if word == "end" {
		return END_DEC
	}

	if word == "if" {
		return IF
	}

	if word == "then" {
		return THEN
	}

	if word == "else" {
		return ELSE
	}

	if word == "while" {
		return WHILE
	}

	if word == "do" {
		return DO
	}

	if word == "and" {
		return AND
	}

	if word == "or" {
		return OR
	}

	if word == "not" {
		return NOT
	}

	if word == "mod" {
		return MOD
	}

	if word == "call" {
		return CALL
	}

	return NULL
}
