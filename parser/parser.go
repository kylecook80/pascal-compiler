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

	ioutil.WriteFile(GenerateTimeString(time.Now())+"_token_file.txt", parser.tokenFile, 0644)
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
			parser.tokenFile = newTokenFile
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
}
func (parser *Parser) program() {
	parser.nextTok()
	parser.expect(PROG)
	parser.expect(ID)
	parser.expect(LEFT_PAREN)
	parser.identifier_list()
	parser.expect(RIGHT_PAREN)
	parser.expect(SEMI)
	parser.program_prime()
}

func (parser *Parser) program_prime() {
	if parser.accept(VAR) {
		parser.declarations()
		parser.program_double_prime()
	} else if parser.accept(PROC) {
		parser.subprogram_declarations()
		parser.compound_statement()
		parser.expect(END)
	} else if parser.accept(BEGIN) {
		parser.compound_statement()
		parser.expect(END)
	} else {
		// ERROR
	}
}

func (parser *Parser) program_double_prime() {
	if parser.accept(PROC) {
		parser.subprogram_declarations()
		parser.compound_statement()
		parser.expect(END)
	} else if parser.accept(BEGIN) {
		parser.compound_statement()
		parser.expect(END)
	} else {
		// ERROR
	}
}

func (parser *Parser) identifier_list() {
	parser.expect(ID)
	parser.identifier_list_prime()
}

func (parser *Parser) identifier_list_prime() {
	if parser.accept(COMMA) {
		parser.expect(COMMA)
		parser.identifier_list_prime()
	} else if parser.accept(RIGHT_PAREN) {
		// NOOP
	} else {
		// ERROR
	}
}

func (parser *Parser) declarations() {
	parser.expect(VAR)
	parser.expect(ID)
	parser.expect(COLON)
	parser.type_prod()
	parser.expect(SEMI)
	parser.declarations_prime()
}

func (parser *Parser) declarations_prime() {
	if parser.accept(VAR) {
		parser.expect(VAR)
		parser.expect(ID)
		parser.expect(COLON)
		parser.type_prod()
		parser.expect(SEMI)
		parser.declarations_prime()
	} else if parser.accept(PROC) || parser.accept(BEGIN) {
		// NOOP
	} else {
		// ERROR
	}
}

func (parser *Parser) type_prod() {
	if parser.accept(INT) || parser.accept(REAL) {
		parser.standard_type()
	} else if parser.accept(ARRAY) {
		parser.expect(ARRAY)
		parser.expect(LEFT_BRACKET)
		parser.expect(NUM)
		parser.expect(RANGE)
		parser.expect(NUM)
		parser.expect(RIGHT_BRACKET)
		parser.expect(OF)
		parser.standard_type()
	} else {
		// ERROR
	}
}

func (parser *Parser) standard_type() {
	if parser.accept(INT_DEC) {
		parser.expect(INT_DEC)
	} else if parser.accept(REAL_DEC) {
		parser.expect(REAL_DEC)
	} else {
		// ERROR
	}
}

func (parser *Parser) subprogram_declarations() {
	parser.subprogram_declaration()
	parser.expect(SEMI)
	parser.subprogram_declarations_prime()
}

func (parser *Parser) subprogram_declarations_prime() {
	if parser.accept(PROC) {
		parser.expect(PROC)
		parser.expect(SEMI)
		parser.subprogram_declarations_prime()
	} else if parser.accept(BEGIN) {
		// NOOP
	} else {
		// ERROR
	}
}

func (parser *Parser) subprogram_declaration() {
	parser.subprogram_head()
	parser.subprogram_declaration_prime()
}

func (parser *Parser) subprogram_declaration_prime() {
	if parser.accept(VAR) {
		parser.declarations()
		parser.subprogram_declaration_double_prime()
	} else if parser.accept(BEGIN) {
		parser.compound_statement()
	} else if parser.accept(PROC) {
		parser.subprogram_declarations()
	} else {
		// ERROR
	}
}

func (parser *Parser) subprogram_declaration_double_prime() {
	if parser.accept(BEGIN) {
		parser.compound_statement()
	} else if parser.accept(PROC) {
		parser.subprogram_declarations()
		parser.compound_statement()
	}
}

func (parser *Parser) subprogram_head() {
	parser.expect(PROC)
	parser.expect(ID)
	parser.subprogram_head_prime()
}

func (parser *Parser) subprogram_head_prime() {
	if parser.accept(LEFT_PAREN) {
		parser.arguments()
	} else if parser.accept(SEMI) {
		parser.expect(SEMI)
	} else {
		// ERROR
	}
}

func (parser *Parser) arguments() {
	parser.expect(LEFT_PAREN)
	parser.parameter_list()
	parser.expect(RIGHT_PAREN)
}

func (parser *Parser) parameter_list() {
	parser.expect(ID)
	parser.expect(COLON)
	parser.type_prod()
	parser.parameter_list_prime()
}

func (parser *Parser) parameter_list_prime() {
	if parser.accept(SEMI) {
		parser.expect(SEMI)
		parser.expect(ID)
		parser.expect(COLON)
		parser.type_prod()
		parser.parameter_list_prime()
	} else if parser.accept()
}

func (parser *Parser) compound_statement() {

}

func (parser *Parser) compound_statement_prime() {

}

func (parser *Parser) optional_statements() {

}

func (parser *Parser) statement_list() {

}

func (parser *Parser) statement_list_prime() {

}

func (parser *Parser) statement() {

}

func (parser *Parser) statement_prime() {

}

func (parser *Parser) variable() {

}

func (parser *Parser) variable_prime() {

}

func (parser *Parser) procedure_statement() {

}

func (parser *Parser) procedure_statement_prime() {

}

func (parser *Parser) expression_list() {

}

func (parser *Parser) expression_list_prime() {

}

func (parser *Parser) expression() {

}

func (parser *Parser) expression_prime() {

}

func (parser *Parser) simple_expression() {

}

func (parser *Parser) simple_expression_prime() {

}

func (parser *Parser) term() {

}

func (parser *Parser) term_prime() {

}

func (parser *Parser) factor() {

}

func (parser *Parser) factor_prime() {

}

func (parser *Parser) sign() {

}
