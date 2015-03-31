package parser

import (
	. "compiler/scanner"
	. "compiler/util"
	"fmt"
	"io/ioutil"
	_ "reflect"
	"strconv"
	"strings"
	"time"
)

type Parser struct {
	scanner   *Scanner
	listing   *ListingFile
	source    *Buffer
	tokenFile []byte
	tok       Token
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

	parser.scanner.SymbolTable().Print()
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

func (parser *Parser) expect(t interface{}) Token {
	if parser.accept(t) {
		currentToken := parser.tok
		parser.nextTok()
		return currentToken
	} else {
		parser.printError(t)
		parser.sync(t)
		parser.nextTok()
		return Token{}
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
		// return identifier_list_prime
	} else {
		// ERROR
		parser.printError(",", ")")
		parser.sync(RIGHT_PAREN)
	}
}

func (parser *Parser) declarations() {
	parser.expect(VAR)
	id := parser.expect(ID)
	parser.expect(COLON)
	parser.type_prod(id.Value())

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
		id := parser.expect(ID)
		parser.expect(COLON)
		parser.type_prod(id.Value())

		// AddBlueNode(GetPtr(id), type, declarations_prime, offset)
		// offset += type.size
		// GetPtr(id) -> declarations_prime

		parser.expect(SEMI)
		parser.declarations_prime()

		// return declarations_prime
	} else if parser.accept(PROC | BEGIN) {
		// NOOP
		// return declarations_prime
	} else {
		// ERROR
		parser.printError("var", "procedure", "begin")
		parser.sync(PROC | BEGIN)
	}
}

func (parser *Parser) type_prod(id string) {
	if parser.accept(INT_DEC | REAL_DEC) {
		parser.standard_type(id)
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

		parser.standard_type(id)

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

func (parser *Parser) standard_type(id string) {
	if parser.accept(INT_DEC) {
		parser.expect(INT_DEC)
		parser.scanner.SymbolTable().AssignType(id, INT)
		// return {INT, INTSIZE}
	} else if parser.accept(REAL_DEC) {
		parser.expect(REAL_DEC)
		parser.scanner.SymbolTable().AssignType(id, REAL)
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
		// return subprogram_decalration_double_prime
	} else if parser.accept(PROC) {
		// subprogram_declaration_double_prime -> subprogram_declarations
		parser.subprogram_declarations()
		parser.compound_statement()
		// return subprogram_declarations
	}
}

func (parser *Parser) subprogram_head() {
	parser.expect(PROC)
	parser.expect(ID)

	// if AddGreenNode(id, PROCNAME, subprogram_head) != NULL {
	// 		PushGreenStack(GetPtr(id))
	//		GetPtr(id) -> subprogram_head_prime
	// }

	parser.subprogram_head_prime()

	// if subprogram_head == 0 {
	// 	PopGreenStack(GetPtr(id))
	// 	return subprogram_head_prime
	// } else {
	// 	return subprogram_head
	// }
}

func (parser *Parser) subprogram_head_prime() {
	if parser.accept(LEFT_PAREN) {
		// subprogram_head_prime -> arguments
		parser.arguments()
		parser.expect(SEMI)
		// return arguments
	} else if parser.accept(SEMI) {
		parser.expect(SEMI)
	} else {
		// ERROR
		parser.printError("(", ";")
		parser.sync(SEMI)
	}
}

func (parser *Parser) arguments() {
	// GreenStack.top.numParams = 0
	// arguments -> parameter_list
	parser.expect(LEFT_PAREN)
	parser.parameter_list()
	parser.expect(RIGHT_PAREN)
	// return parameter_list
}

func (parser *Parser) parameter_list() {
	id := parser.expect(ID)
	parser.expect(COLON)
	parser.type_prod(id.Value())

	// if parameter_list != null {
	// 	if type_prod.type == INT {
	// 		temp = AddBlueNode(id, PPINT, parameter_list)
	// 	} else if type_prod.type == REAL {
	// 		temp = AddBlueNode(id, PPREAL, parameter_list)
	// 	} else if type_prod.type == AINT {
	// 		temp = AddBlueNode(id, PPAINT, parameter_list)
	// 	} else if type_prod.type == AREAL {
	// 		temp = AddBlueNode(id, PPAREAL, parameter_list)
	// 	} else {
	// 		temp = AddBlueNode(id, ERR, parameter_list)
	// 	}

	// 	if temp == NULL {
	// 		fmt.Print("Parameter Error")
	// 	} else {
	// 		GreenStack.top.numParams++
	// 		temp -> parameter_list_prime
	// 	}
	// } else {
	// 	GetPtr(id) -> parameter_list_prime
	// }

	parser.parameter_list_prime()

	// return parameter_list_prime
}

func (parser *Parser) parameter_list_prime() {
	if parser.accept(SEMI) {
		parser.expect(SEMI)
		id := parser.expect(ID)
		parser.expect(COLON)
		parser.type_prod(id.Value())

		// if parameter_list != null {
		// 	if type_prod.type == INT {
		// 		temp = AddBlueNode(id, PPINT, parameter_list_prime)
		// 	} else if type_prod.type == REAL {
		// 		temp = AddBlueNode(id, PPREAL, parameter_list_prime)
		// 	} else if type_prod.type == AINT {
		// 		temp = AddBlueNode(id, PPAINT, parameter_list_prime)
		// 	} else if type_prod.type == AREAL {
		// 		temp = AddBlueNode(id, PPAREAL, parameter_list_prime)
		// 	} else {
		// 		temp = AddBlueNode(id, ERR, parameter_list_prime)
		// 	}

		// 	if temp == NULL {
		// 		fmt.Print("Parameter Error")
		// 	} else {
		// 		GreenStack.top.numParams++
		// 		temp -> parameter_list_prime
		// 	}
		// } else {
		// 	GetPtr(id) -> parameter_list_prime
		// }

		parser.parameter_list_prime()

		// return parameter_list_prime
	} else if parser.accept(RIGHT_PAREN) {
		parser.expect(RIGHT_PAREN)
		// return parameter_list_prime
	} else {
		// ERROR
		parser.printError("(")
		parser.sync(RIGHT_PAREN)
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
		// if variable.type != ERR && variable.type != expression.type {
		// 	fmt.Print("Operand type mismatch")
		// }
	} else if parser.accept(CALL) {
		parser.procedure_statement()
	} else if parser.accept(BEGIN) {
		parser.compound_statement()
	} else if parser.accept(IF) {
		parser.expect(IF)
		parser.expression()
		// if expression.type != BOOL {
		// 	fmt.Print("expression not a boolean expression")
		// }
		parser.expect(THEN)
		parser.statement()
		parser.statement_prime()
	} else if parser.accept(WHILE) {
		parser.expect(WHILE)
		parser.expression()
		// if expression.type != BOOL {
		// 	fmt.Print("expression not a boolean expression")
		// }
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
	// GetPtr(id) -> variable_prime
	parser.variable_prime()
	// return variable_prime
}

func (parser *Parser) variable_prime() {
	if parser.accept(LEFT_BRACKET) {
		parser.expect(LEFT_BRACKET)
		parser.expression()
		parser.expect(RIGHT_BRACKET)

		// if expression == INT {
		// 	if GetType(variable_prime) == AINT or GetType(variable_prime) == FPAINT {
		// 		variable_prime = INT
		// 	}

		// 	if GetType(variable_prime) == AREAL or GetType(variable_prime) == FPAREAL {
		// 		variable_prime = REAL
		// 	}

		// 	if GetType(variable_prime) == ERR {
		// 		variable_prime = ERR
		// 	} else {
		// 		variable_prime = ERR_STAR
		// 		fmt.Print("Error all the things")
		// 	}
		// } else {
		// 	variable_prime = ERR_STAR
		// 	fmt.Print("Error of the other errors");
		// }
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
	// PushGreenStack(expression_list)
	// expression_list_prime = 1
	// if GreenStack.top.numParams == 0 {
	// 	fmt.Print("More than 0 parameters given")
	// 	return false
	// }
	parser.expression_list_prime()
	// if expression == GetParam(0, GreenStack.top) {
	// 	return expression_list_prime
	// } else {
	// 	fmt.Print("Type of parameter 0 does not match")
	// 	PopGreenNodeStack()
	// 	return false
	// }
}

func (parser *Parser) expression_list_prime() {
	if parser.accept(COMMA) {
		parser.expect(COMMA)
		parser.expression()
		// expression_list_prime -> expression_list_prime
		// if GreenStack.top.numParams == expression_list_prime {
		// 	fmt.Print("More than one parameter supplied")
		// }
		// return false
		parser.expression_list_prime()
		// if expression == GetParam(expression_list_prime, GreenStack.top) {
		// 	return expression_list_prime
		// } else {
		// 	return false
		// }
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
	// simple_expression -> expression_prime
	parser.expression_prime()
	// return expression_prime
}

func (parser *Parser) expression_prime() {
	if parser.accept(RELOP) {
		parser.expect(RELOP)
		parser.simple_expression()
		// if expression_prime == simple_expression {
		// 	if expression_prime == INT || expression_prime == REAL {
		// 		expression_prime_type = BOOL
		// 	} else if expression_prime == ERR {
		// 		expression_prime_type = ERR
		// 	} else {
		// 		expression_prime_type = ERR_STAR
		// 		fmt.Print("Operand types do not match")
		// 	}
		// }
	} else if parser.accept(END_DEC | SEMI | ELSE | THEN | DO | RIGHT_BRACKET | RIGHT_PAREN | COMMA) {
		// NOOP
		// return expression_prime
	} else {
		// ERROR
		parser.printError("<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(END_DEC | SEMI | ELSE | THEN | DO | RIGHT_BRACKET | RIGHT_PAREN | COMMA)
	}
}

func (parser *Parser) simple_expression() {
	if parser.accept(ID|NUM) || parser.accept(LEFT_PAREN|NOT) {
		parser.term()
		// term_type -> simple_expression_prime
		parser.simple_expression_prime()
		// return simple_expression_prime
	} else if parser.accept(ADD) || parser.accept(SUB) {
		parser.sign()
		parser.term()
		// if term == INT || term == REAL {
		// 	term -> simple_expression_prime
		// } else {
		// 	ERR_STAR -> simple_expression_prime
		// 	fmt.Print("A sign can only be used with an integer or real")
		// }
		parser.simple_expression_prime()
		// return simple_expression_prime
	}
}

func (parser *Parser) simple_expression_prime() {
	if parser.accept(ADDOP) {
		parser.expect(ADDOP)
		parser.term()
		// if simple_expression_prime == term {
		// 	if addop == OR && simple_expression_prime != BOOL {
		// 		simple_expression_prime = ERR_STAR
		// 		fmt.Print("Types do not match operation")
		// 	}
		//  if addop != OR && simple_expression_prime != INT {
		// 		simple_expression_prime = ERR_STAR
		// 		fmt.Print("Types do not match operation")
		//  }
		// 	if simple_expression_prime == BOOL {
		// 		BOOL -> simple_expression_prime
		// 	} else if simple_expression_prime == INT {
		// 		INT -> simple_expression_prime
		// 	} else if simple_expression_prime == REAL {
		// 		REAL -> simple_expression_prime
		// 	}	else {
		// 		ERR_STAR -> simple_expression_prime
		// 		fmt.Print("Mixed mode operations not allowed")
		// 	}
		// }
		parser.simple_expression_prime()
		// return simple_express_prime
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
	// factor -> term_prime
	parser.term_prime()
	// return term_prime
}

func (parser *Parser) term_prime() {
	if parser.accept(MULOP) {
		parser.expect(MULOP)
		parser.factor()
		// if (mulop == '*' || mulop == '/' || mulop == 'div') || (mulop == 'mod' && term_prime == INT) || (mulop == 'and' && term_prime == BOOL) {
		// 	if term_prime != BOOL {
		// 		factor -> term_prime
		// 	} else {
		// 		ERR_STAR -> term_prime
		// 		fmt.Print("Types do not match operation")
		// 	}
		// } else {
		// 	ERR_STAR -> term_prime
		// 	fmt.Print("Mixed mode operations not allowed")
		// }
		parser.term_prime()
		// return factor
	} else if parser.accept(ADDOP|RELOP) || parser.accept(END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA) {
		// NOOP
		// return term_prime
	} else {
		// ERROR
		parser.printError("*", "+", "<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(ADDOP|RELOP, END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA)
	}
}

func (parser *Parser) factor() AttributeType {
	if parser.accept(NUM) {
		num := parser.expect(NUM)
		return num.Attr()
	} else if parser.accept(LEFT_PAREN) {
		parser.expect(LEFT_PAREN)

		// expression := parser.expression()

		parser.expect(RIGHT_PAREN)
		// return expression
		return NULL
	} else if parser.accept(ID) {
		id := parser.expect(ID)
		name := id.Value()

		// GetPtr(id) -> factor_prime
		elem, err := parser.scanner.SymbolTable().GetPtr(name)
		sym := elem.Value.(Symbol)

		var factor_prime AttributeType
		if err != nil {
			fmt.Errorf("Error!")
		} else {
			factor_prime = parser.factor_prime(sym)
		}

		return factor_prime
	} else if parser.accept(NOT) {
		parser.expect(NOT)
		factor_prime := parser.factor_prime(Symbol{})

		if factor_prime == BOOL {
			return BOOL
		} else if factor_prime == ERR {
			return ERR
		} else {
			fmt.Print("factor_prime error")
			return ERR_STAR
		}
	} else {
		// ERROR
		parser.printError("a number", "(", "an identifier", "not")
		parser.sync(LEFT_PAREN|NOT, ID)
		return NULL
	}
}

func (parser *Parser) factor_prime(sym Symbol) AttributeType {
	if parser.accept(LEFT_BRACKET) {
		parser.expect(LEFT_BRACKET)
		// expression :=
		parser.expression()
		parser.expect(RIGHT_BRACKET)
		// if expression != INT {
		// 	fmt.Print("Error")
		// 	return ERR_STAR
		// }
		return NULL
	} else if parser.accept(ADDOP|MULOP|RELOP) || parser.accept(END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA) {
		// NOOP
		if sym.GetType() == PPINT {
			return INT
		} else if sym.GetType() == PPREAL {
			return REAL
		} else if sym.GetType() == PPAINT {
			return AINT
		} else if sym.GetType() == PPAREAL {
			return AREAL
		} else {
			return NULL
		}
		// return NULL
	} else {
		// ERROR
		parser.printError("[", "*", "+", "<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(ADDOP|MULOP|RELOP, END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA)
		return NULL
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
