package parser

import (
	. "compiler/scanner"
	. "compiler/util"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
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
	parser.tok = tok

	if parser.scanner.CurrentLineNumber() >= parser.listing.LineCount() {
		parser.listing.AddLine(parser.source.ReadLine(parser.scanner.CurrentLineNumber()))
	}

	if err != nil {
		parser.listing.AddLexError(err.Error())
	} else {
		line := parser.scanner.CurrentLineNumber() + 1
		if tok.Type() != WS {
			newTokenFile := append(parser.tokenFile, []byte(strconv.Itoa(line)+": "+tok.String()+"\n")...)
			parser.tokenFile = newTokenFile
		}

		if tok.Type() == EOF {
			return
		}
	}
}

func (parser *Parser) accept(t interface{}) bool {
	for parser.tok.Type() == WS {
		parser.nextTok()
	}

	for parser.tok.Type() == LEXERR {
		parser.nextTok()
	}

	switch sym := t.(type) {
	case TokenType:
		if parser.tok.Type() == sym&parser.tok.Type() {
			// fmt.Printf("Expected: %s, got %s\n", sym, parser.tok)
			return true
		}
	case AttributeType:
		if parser.tok.Attr() == sym&parser.tok.Attr() {
			// fmt.Printf("Expected: %s, got %s\n", sym, parser.tok)
			return true
		}
	}

	return false
}

func (parser *Parser) expect(t interface{}) {
	if parser.accept(t) {
		parser.nextTok()
	} else {
		parser.printError(t)
		parser.sync(t)
		parser.nextTok()
	}
}

func (parser *Parser) printError(t ...interface{}) {
	str := make([]string, len(t))

	for i, v := range t {
		switch sym := v.(type) {
		case TokenType:
			str[i] = sym.String()
		case AttributeType:
			str[i] = sym.String()
		case string:
			str[i] = sym
		}
	}

	if parser.tok.Type() == EOF {
		return
	}

	msg := fmt.Sprintf("expected \"%s\", got \"%s\"", strings.Join(str, "\", or \""), parser.tok.Value())
	fmt.Print("Syntax error: " + msg + "\n")
	parser.listing.AddSyntaxError(msg)
}

func (parser *Parser) sync(t ...interface{}) {
	for {
		if parser.tok.Type() == EOF {
			return
		}

		for _, v := range t {
			if parser.accept(v) {
				return
			}
		}

		parser.nextTok()
	}
}

func (parser *Parser) program() {
	parser.nextTok()
	parser.expect(PROG)
	parser.expect(ID)

	// AddGreenNode(id, PGNAME, nil)
	// offset = 0
	// GetPtr(id) -> identifier_list

	parser.expect(LEFT_PAREN)
	parser.identifier_list()
	parser.expect(RIGHT_PAREN)
	parser.expect(SEMI)

	// identifier list -> program_prime

	parser.program_prime()

	// PopGreenStack()
}

func (parser *Parser) program_prime() {
	if parser.accept(VAR) {
		// program_prime -> declarations
		parser.declarations()

		// declarations -> program_double_prime
		parser.program_double_prime()
	} else if parser.accept(PROC) {
		// program_prime -> subprogram_declarations
		parser.subprogram_declarations()
		parser.compound_statement()
		parser.expect(END)
	} else if parser.accept(BEGIN) {
		parser.compound_statement()
		parser.expect(END)
	} else {
		// ERROR
		parser.printError("var", "procedure", "begin")
		parser.sync(EOF)
	}
}

func (parser *Parser) program_double_prime() {
	if parser.accept(PROC) {
		// program_double_prime > subprogram_declarations
		parser.subprogram_declarations()
		parser.compound_statement()
		parser.expect(END)
	} else if parser.accept(BEGIN) {
		parser.compound_statement()
		parser.expect(END)
	} else {
		// ERROR
		parser.printError("procedure", "begin")
		parser.sync(EOF)
	}
}

func (parser *Parser) identifier_list() {
	parser.expect(ID)

	// temp = AddBlueNode(id, PGPARM, identifier_list_param)
	// if temp = nil {
	// 	identifier_list -> identifier_list_prime
	// } else {
	// 	GetPtr(id) -> identifier_list_prime
	// }
	// GreenStack.top.numParams++

	parser.identifier_list_prime()

	//return identifier_list_prime
}

func (parser *Parser) identifier_list_prime() {
	if parser.accept(COMMA) {
		parser.expect(COMMA)
		parser.expect(ID)

		// temp = AddBlueNode
		// if temp == null {
		// 	identifier_list -> identifier_list_prime
		// } else {
		// 	ptr id => identifier_list_prime
		// }
		// greenstack.top.numParams++

		parser.identifier_list_prime()

		// return identifier_list_prime
	} else if parser.accept(RIGHT_PAREN) {
		// NOOP
	} else {
		// ERROR
		parser.printError(",", ")")
		parser.sync(RIGHT_PAREN)
	}
}

func (parser *Parser) declarations() {
	parser.expect(VAR)
	parser.expect(ID)
	parser.expect(COLON)
	parser.type_prod()

	// temp = AddBlueNode(id, type, declarations, offset)
	// if declarations.temp = nil {
	//   declarations -> declarations_prime
	// } else {
	//   offset += type.size
	//   GetPtr(id) -> declarations'
	// }

	parser.expect(SEMI)
	parser.declarations_prime()

	// return declarations_prime
}

func (parser *Parser) declarations_prime() {
	if parser.accept(VAR) {
		parser.expect(VAR)
		parser.expect(ID)
		parser.expect(COLON)
		parser.type_prod()

		// AddBlueNode(GetPtr(id), type, declarations_prime, offset)
		// offset += type.size
		// GetPtr(id) -> declarations_prime

		parser.expect(SEMI)
		parser.declarations_prime()

		// return declarations_prime
	} else if parser.accept(PROC | BEGIN) {
		// NOOP
	} else {
		// ERROR
		parser.printError("var", "procedure", "begin")
		parser.sync(PROC | BEGIN)
	}
}

func (parser *Parser) type_prod() {
	if parser.accept(INT_DEC | REAL_DEC) {
		parser.standard_type()
		// return {standard_type.size, standard_type.type}
	} else if parser.accept(ARRAY) {
		parser.expect(ARRAY)
		parser.expect(LEFT_BRACKET)
		parser.expect(NUM)

		// num = NUM

		parser.expect(RANGE)
		parser.expect(NUM)

		// num_prime = NUM

		parser.expect(RIGHT_BRACKET)
		parser.expect(OF)

		parser.standard_type()

		// if reflect.TypeOf(num) == 'integer' && reflect.TypeOf(num_prime) == 'integer' && num_prime > num {
		// 	myType = MakeArray(standard_type, num, num - num_prime + 1)
		// 	size = (num_prime - num + 1) * sizeof(standard_type)
		// } else {
		// 	myType = ERR_STAR
		// 	size = 0;
		// }
	} else {
		// ERROR
		parser.printError("integer", "real", "array")
		parser.sync(ARRAY)
	}
}

func (parser *Parser) standard_type() {
	if parser.accept(INT_DEC) {
		parser.expect(INT_DEC)
		// return {INT, INTSIZE}
	} else if parser.accept(REAL_DEC) {
		parser.expect(REAL_DEC)
		// return {REAL, REALSIZE}
	} else {
		// ERROR
		parser.printError("integer", "real")
		parser.sync(REAL_DEC)
	}
}

func (parser *Parser) subprogram_declarations() {
	// subprogram_declarations -> subprogram_declaration

	parser.subprogram_declaration()
	parser.expect(SEMI)

	// subprogram_declaration -> subprogrm_declarations_prime

	parser.subprogram_declarations_prime()

	// return subprogram_declarations_prime
}

func (parser *Parser) subprogram_declarations_prime() {
	if parser.accept(PROC) {
		parser.expect(PROC)
		parser.expect(SEMI)
		// subprogram_declarations_prime -> subprogram_declaration
		parser.subprogram_declarations_prime()
	} else if parser.accept(BEGIN) {
		// NOOP
	} else {
		// ERROR
		parser.printError("procedure", "begin")
		parser.sync(BEGIN)
	}
}

func (parser *Parser) subprogram_declaration() {
	// subprogram_declaration -> subprogram_head
	parser.subprogram_head()
	// subprogram_head -> subprogram_declaration_prime
	parser.subprogram_declaration_prime()
	// return subprogram_declaration_prime
}

func (parser *Parser) subprogram_declaration_prime() {
	if parser.accept(VAR) {
		// subprogram_declaration_prime -> declarations
		parser.declarations()
		// declarations -> subprogram_declarations_double_prime
		parser.subprogram_declaration_double_prime()
		// return subprogram_declaration_double_prime
	} else if parser.accept(BEGIN) {
		parser.compound_statement()
	} else if parser.accept(PROC) {
		parser.subprogram_declarations()
	} else {
		// ERROR
		parser.printError("var", "begin", "procedure")
		parser.sync(BEGIN | PROC)
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
		parser.printError("(", ";")
		parser.sync(SEMI)
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
	} else if parser.accept(RIGHT_PAREN) {
		parser.expect(RIGHT_PAREN)
	}
}

func (parser *Parser) compound_statement() {
	parser.expect(BEGIN)
	parser.compound_statement_prime()
}

func (parser *Parser) compound_statement_prime() {
	if parser.accept(ID) || parser.accept(CALL|BEGIN|IF|WHILE|END_DEC) {
		parser.optional_statements()
		parser.expect(END_DEC)
	} else if parser.accept(END_DEC) {
		parser.expect(END_DEC)
	} else {
		// ERROR
		parser.printError("a identifier", "call", "begin", "if", "while", "end")
		parser.sync(END_DEC)
	}
}

func (parser *Parser) optional_statements() {
	parser.statement_list()
}

func (parser *Parser) statement_list() {
	parser.statement()
	parser.statement_list_prime()
}

func (parser *Parser) statement_list_prime() {
	if parser.accept(SEMI) {
		parser.expect(SEMI)
		parser.statement()
		parser.statement_list_prime()
	} else if parser.accept(END_DEC) {
		// NOOP
	} else {
		// ERROR
		parser.printError(";", "end")
		parser.sync(END_DEC)
	}
}

func (parser *Parser) statement() {
	if parser.accept(ID) {
		parser.variable()
		parser.expect(ASSIGNOP)
		parser.expression()
	} else if parser.accept(CALL) {
		parser.procedure_statement()
	} else if parser.accept(BEGIN) {
		parser.compound_statement()
	} else if parser.accept(IF) {
		parser.expect(IF)
		parser.expression()
		parser.expect(THEN)
		parser.statement()
		parser.statement_prime()
	} else if parser.accept(WHILE) {
		parser.expect(WHILE)
		parser.expression()
		parser.expect(DO)
		parser.statement()
	} else {
		// ERRORÏ
		parser.printError("an identifier", "call", "begin", "if", "while")
		parser.sync(CALL | BEGIN | IF | WHILE)
	}
}

func (parser *Parser) statement_prime() {
	if parser.accept(ELSE) {
		parser.expect(ELSE)
		parser.statement()
	} else if parser.accept(ASSIGNOP) {
		// NOOP
	} else {
		// ERROR
		parser.printError("else", ":=")
		parser.sync(ASSIGNOP)
	}
}

func (parser *Parser) variable() {
	parser.expect(ID)
	parser.variable_prime()
}

func (parser *Parser) variable_prime() {
	if parser.accept(LEFT_BRACKET) {
		parser.expect(LEFT_BRACKET)
	} else if parser.accept(ASSIGNOP) {
		// NOOP
	} else {
		// ERROR
		parser.printError("[", ":=")
		parser.sync(ASSIGNOP)
	}
}

func (parser *Parser) procedure_statement() {
	parser.expect(CALL)
	parser.expect(ID)
	parser.procedure_statement_prime()
}

func (parser *Parser) procedure_statement_prime() {
	if parser.accept(LEFT_PAREN) {
		parser.expect(LEFT_PAREN)
		parser.expression_list()
		parser.expect(RIGHT_PAREN)
	} else if parser.accept(END_DEC | SEMI | ELSE) {
		// NOOP
	} else {
		// ERROR
		parser.printError("(", "end", ";", "else")
		parser.sync(END_DEC | SEMI | ELSE)
	}
}

func (parser *Parser) expression_list() {
	parser.expression()
	parser.expression_list_prime()
}

func (parser *Parser) expression_list_prime() {
	if parser.accept(COMMA) {
		parser.expect(COMMA)
		parser.expression()
		parser.expression_list_prime()
	} else if parser.accept(RIGHT_PAREN) {
		// NOOP
	} else {
		// ERROR
		parser.printError(",", ")")
		parser.sync(RIGHT_PAREN)
	}
}

func (parser *Parser) expression() {
	parser.simple_expression()
	parser.expression_prime()
}

func (parser *Parser) expression_prime() {
	if parser.accept(RELOP) {
		parser.expect(RELOP)
		parser.simple_expression()
	} else if parser.accept(END_DEC | SEMI | ELSE | THEN | DO | RIGHT_BRACKET | RIGHT_PAREN | COMMA) {
		// NOOP
	} else {
		// ERROR
		parser.printError("<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(END_DEC | SEMI | ELSE | THEN | DO | RIGHT_BRACKET | RIGHT_PAREN | COMMA)
	}
}

func (parser *Parser) simple_expression() {
	if parser.accept(ID|NUM) || parser.accept(LEFT_PAREN|NOT) {
		parser.term()
		parser.simple_expression_prime()
	} else if parser.accept(ADD) || parser.accept(SUB) {
		parser.sign()
		parser.simple_expression_prime()
	}
}

func (parser *Parser) simple_expression_prime() {
	if parser.accept(ADDOP) {
		parser.expect(ADDOP)
		parser.term()
		parser.simple_expression_prime()
	} else if parser.accept(RELOP) || parser.accept(END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA) {
		// NOOP
	} else {
		// ERROR
		parser.printError("+", "<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(RELOP, END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA)
	}
}

func (parser *Parser) term() {
	parser.factor()
	parser.term_prime()
}

func (parser *Parser) term_prime() {
	if parser.accept(MULOP) {
		parser.expect(MULOP)
		parser.factor()
		parser.term_prime()
	} else if parser.accept(ADDOP|RELOP) || parser.accept(END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA) {
		// NOOP
	} else {
		// ERROR
		parser.printError("*", "+", "<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(ADDOP|RELOP, END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA)
	}
}

func (parser *Parser) factor() {
	if parser.accept(NUM) {
		parser.expect(NUM)
	} else if parser.accept(LEFT_PAREN) {
		parser.expect(LEFT_PAREN)
		parser.expression()
		parser.expect(RIGHT_PAREN)
	} else if parser.accept(ID) {
		parser.expect(ID)
		parser.factor_prime()
	} else if parser.accept(NOT) {
		parser.expect(NOT)
		parser.factor()
	} else {
		// ERROR
		parser.printError("a number", "(", "an identifier", "not")
		parser.sync(LEFT_PAREN|NOT, ID)
	}
}

func (parser *Parser) factor_prime() {
	if parser.accept(LEFT_BRACKET) {
		parser.expect(LEFT_BRACKET)
		parser.expression()
		parser.expect(RIGHT_BRACKET)
	} else if parser.accept(ADDOP|MULOP|RELOP) || parser.accept(END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA) {
		// NOOP
	} else {
		// ERROR
		parser.printError("[", "*", "+", "<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(ADDOP|MULOP|RELOP, END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA)
	}
}

func (parser *Parser) sign() {
	if parser.accept(ADD) {
		parser.expect(ADD)
	} else if parser.accept(SUB) {
		parser.expect(SUB)
	} else {
		// ERROR
		parser.printError("+", "-")
	}
}
