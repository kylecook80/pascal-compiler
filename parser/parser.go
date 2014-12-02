package parser

import (
	. "compiler/scanner"
	. "compiler/util"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"time"
)

type Parser struct {
	scanner   *Scanner
	tok       Token
	listing   *ListingFile
	tokenFile []byte
	source    *Buffer
}

func NewParser(scanner *Scanner) Parser {
	return Parser{scanner: scanner}
}

func (parser *Parser) Begin(file string) {
	listing := NewListingFile()
	tokenFile := []byte{}
	source := ReadFile(file)

	parser.listing = listing
	parser.source = source
	parser.tokenFile = tokenFile

	parser.program()

	ioutil.WriteFile(GenerateTimeString(time.Now())+"_token_file.txt", tokenFile, 0644)
	parser.listing.Save()
}

func (parser *Parser) nextTok() {
	tok, err := parser.scanner.NextToken()

	if parser.scanner.CurrentLineNumber() >= parser.listing.LineCount() {
		parser.listing.AddLine(parser.source.ReadLine(parser.scanner.CurrentLineNumber()))
	}

	if err == io.EOF {
		parser.tok = Token{}
		return
	} else if err != nil {
		parser.listing.AddError(err.Error())
	} else {
		line := parser.scanner.CurrentLineNumber() + 1
		if tok.Type() != WS {
			newTokenFile := append(parser.tokenFile, []byte(strconv.Itoa(line)+": "+tok.String()+"\n")...)
			fmt.Printf("%s\n", newTokenFile)
		}
	}

	parser.tok = tok
}

func (parser *Parser) accept(t interface{}) bool {
	for parser.tok.Type() == WS {
		parser.nextTok()
	}

	switch sym := t.(type) {
	case TokenType:
		if parser.tok.Type() == sym {
			parser.nextTok()
			return true
		}
	case Attribute:
		if parser.tok.Attr() == sym {
			parser.nextTok()
			return true
		}
	}

	return false
}

func (parser *Parser) expect(t interface{}) {
	if parser.accept(t) {
		return
	} else {
		fmt.Println("Error: unexpected symbol: " + parser.tok.Value())
		fmt.Printf("token type: %s\n", TokenStrings[parser.tok.Type()])
		parser.nextTok()
	}
	return
}

func (parser *Parser) program() {
	parser.nextTok()
	parser.expect(PROG)
	parser.expect(ID)
	parser.expect(LEFT_PAREN)
	return
}
