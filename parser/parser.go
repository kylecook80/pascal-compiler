package parser

import (
	. "compiler/scanner"
	. "compiler/util"
	"fmt"
	"io/ioutil"
	_ "reflect"
	"strconv"
	"strings"
	_ "time"
)

type Parser struct {
	scanner   *Scanner
	listing   *ListingFile
	source    *Buffer
	tokenFile []byte
	tok       Token
	scope     *ScopeTree
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

	parser.scope = NewScopeTree()

	parser.program()

	// parser.scanner.SymbolTable().Print()
	// ioutil.WriteFile(GenerateTimeString(time.Now())+"_token_file.txt", parser.tokenFile, 0644)
	ioutil.WriteFile("token_file.txt", parser.tokenFile, 0644)
	parser.scanner.SymbolTable().Write()
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
			return true
		}
	case AttributeType:
		if parser.tok.Attr() == sym&parser.tok.Attr() {
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
	// fmt.Print("Syntax error: " + msg + "\n")
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

func (parser *Parser) CheckType(value AttributeType, checked AttributeType, msg string) bool {
	if value != value&checked {
		parser.listing.AddTypeError(msg)
		return true
	} else {
		return false
	}
}

func (parser *Parser) program() {
	parser.nextTok()
	parser.expect(PROG)
	programName := parser.expect(ID)

	newSymbol := NewSymbol(programName.Value(), PGNAME)
	parser.scanner.SymbolTable().AddSymbol(newSymbol)
	parser.scope.CreateRoot(programName.Value(), newSymbol)

	// AddGreenNode(id, PGNAME, nil)
	// offset = 0
	// GetPtr(id) -> identifier_list

	parser.expect(LEFT_PAREN)
	parser.identifier_list()
	parser.expect(RIGHT_PAREN)
	parser.expect(SEMI)

	// identifier list -> program_prime

	parser.program_prime()

	parser.scope.Pop()
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
	progParm := parser.expect(ID)
	symbol := NewSymbol(progParm.Value(), PGPARM)
	parser.scanner.SymbolTable().AddSymbol(symbol)
	parser.scope.GetTop().AddBlueNode(progParm.Value(), symbol)

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

		progParm := parser.expect(ID)
		symbol := NewSymbol(progParm.Value(), PGPARM)
		parser.scanner.SymbolTable().AddSymbol(symbol)
		parser.scope.GetTop().AddBlueNode(progParm.Value(), symbol)

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

	typeName := parser.type_prod(id.Value())
	symbol := NewSymbol(id.Value(), typeName)
	parser.scanner.SymbolTable().AddSymbol(symbol)
	parser.scope.GetTop().AddBlueNode(id.Value(), symbol)

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

		typeName := parser.type_prod(id.Value())
		symbol := NewSymbol(id.Value(), typeName)
		parser.scanner.SymbolTable().AddSymbol(symbol)
		parser.scope.GetTop().AddBlueNode(id.Value(), symbol)

		// parser.scanner.SymbolTable().AssignType(id.Value(), typeName)

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

func (parser *Parser) type_prod(id string) AttributeType {
	if parser.accept(INT_DEC | REAL_DEC) {
		standard_type := parser.standard_type(id)
		return standard_type
	} else if parser.accept(ARRAY) {
		parser.expect(ARRAY)
		parser.expect(LEFT_BRACKET)

		num1 := parser.expect(NUM)
		if parser.CheckType(num1.Attr(), INT, "Array index type mismatch") {
			return ERR
		}

		parser.expect(RANGE)

		num2 := parser.expect(NUM)
		if parser.CheckType(num2.Attr(), INT, "Array index type mismatch") {
			return ERR
		}

		parser.expect(RIGHT_BRACKET)
		parser.expect(OF)

		standard_type := parser.standard_type(id)

		if standard_type == PPINT {
			return PPAINT
		} else if standard_type == PPREAL {
			return PPAREAL
		} else if standard_type == INT {
			return AINT
		} else if standard_type == REAL {
			return AREAL
		} else {
			return ERR
		}

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
		return ERR
	}
}

func (parser *Parser) standard_type(id string) AttributeType {
	if parser.accept(INT_DEC) {
		parser.expect(INT_DEC)
		return INT
	} else if parser.accept(REAL_DEC) {
		parser.expect(REAL_DEC)
		return REAL
	} else {
		// ERROR
		parser.printError("integer", "real")
		parser.sync(REAL_DEC)
		return ERR
	}
}

func (parser *Parser) subprogram_declarations() {
	// subprogram_declarations -> subprogram_declaration

	parser.subprogram_declaration()
	parser.expect(SEMI)

	// subprogram_declaration -> subprogram_declarations_prime

	parser.subprogram_declarations_prime()

	// return subprogram_declarations_prime
}

func (parser *Parser) subprogram_declarations_prime() {
	if parser.accept(PROC) {
		parser.subprogram_declaration()
		parser.expect(SEMI)
		parser.subprogram_declarations_prime()
		// subprogram_declarations_prime -> subprogram_declaration
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
	parser.scope.Pop()
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

	procName := parser.expect(ID)
	symbol := NewSymbol(procName.Value(), PROC)
	parser.scanner.SymbolTable().AddSymbol(symbol)
	parser.scope.AddGreenNode(procName.Value(), symbol)

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
	typeName := parser.type_prod(id.Value())

	var symbol *Symbol
	if typeName == INT {
		symbol = NewSymbol(id.Value(), PPINT)
	} else if typeName == REAL {
		symbol = NewSymbol(id.Value(), PPREAL)
	} else if typeName == AINT {
		symbol = NewSymbol(id.Value(), PPAINT)
	} else if typeName == AREAL {
		symbol = NewSymbol(id.Value(), PPAREAL)
	}

	parser.scanner.SymbolTable().AddSymbol(symbol)
	parser.scope.GetTop().AddBlueNode(id.Value(), symbol)

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
		typeName := parser.type_prod(id.Value())

		var symbol *Symbol
		if typeName == INT {
			symbol = NewSymbol(id.Value(), PPINT)
		} else if typeName == REAL {
			symbol = NewSymbol(id.Value(), PPREAL)
		} else if typeName == AINT {
			symbol = NewSymbol(id.Value(), PPAINT)
		} else if typeName == AREAL {
			symbol = NewSymbol(id.Value(), PPAREAL)
		}

		parser.scanner.SymbolTable().AddSymbol(symbol)
		parser.scope.GetTop().AddBlueNode(id.Value(), symbol)

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
	if parser.accept(ID) || parser.accept(CALL|BEGIN|IF|WHILE) {
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

func (parser *Parser) statement() AttributeType {
	if parser.accept(ID) {
		variable := parser.variable()
		parser.expect(ASSIGNOP)
		expression := parser.expression()

		switch variable {
		case PPINT:
			variable = INT
		case PPREAL:
			variable = REAL
		case PPAINT:
			variable = AINT
		case PPAREAL:
			variable = AREAL
		}

		if parser.CheckType(variable, expression, "ASSIGNOP type mismatch") {
			return ERR
		}

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

		// if expression != BOOL {
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
		// ERROR
		parser.printError("an identifier", "call", "begin", "if", "while")
		parser.sync(CALL | BEGIN | IF | WHILE)
		return ERR
	}

	return NULL
}

func (parser *Parser) statement_prime() {
	if parser.accept(ELSE) {
		parser.expect(ELSE)
		parser.statement()
	} else if parser.accept(END_DEC | SEMI | ELSE) {
		// NOOP
	} else {
		// ERROR
		parser.printError("end", ";", "else")
		parser.sync(ASSIGNOP)
	}
}

func (parser *Parser) variable() AttributeType {
	id := parser.expect(ID)
	// sym, _ := parser.scanner.SymbolTable().GetPtr(id.Value())

	blueNode, err := parser.scope.GetTop().FindBlueNode(id.Value())
	if err != nil {
		parser.listing.AddTypeError("Variable not declared")
		return ERR
	}

	sym := blueNode.GetSymbol()

	variable_prime := parser.variable_prime(sym.GetType())
	return variable_prime
}

func (parser *Parser) variable_prime(id AttributeType) AttributeType {
	if parser.accept(LEFT_BRACKET) {
		parser.expect(LEFT_BRACKET)
		expression := parser.expression()
		parser.expect(RIGHT_BRACKET)

		if parser.CheckType(expression, INT, "Only use integers as array indices") {
			return ERR
		}

		return expression

		// if expression == INT {
		// 	if GetType(variable_prime) == AINT or GetType(variable_prime) == PPAINT {
		// 		variable_prime = INT
		// 	}

		// 	if GetType(variable_prime) == AREAL or GetType(variable_prime) == PPAREAL {
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
		return id
	} else {
		// ERROR
		parser.printError("[", ":=")
		parser.sync(ASSIGNOP)
		return ERR
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

func (parser *Parser) expression() AttributeType {
	simple_expression := parser.simple_expression()
	expression_prime := parser.expression_prime(simple_expression)

	return expression_prime
}

func (parser *Parser) expression_prime(expr AttributeType) AttributeType {
	if parser.accept(RELOP) {
		parser.expect(RELOP)
		simple_expression := parser.simple_expression()

		errMsg := "RELOP type mismatch"
		if parser.CheckType(simple_expression, INT|REAL, errMsg) {
			return ERR
		}

		if parser.CheckType(expr, simple_expression, errMsg) {
			return ERR
		}

		return BOOL
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
		return expr
	} else {
		// ERROR
		parser.printError("<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(END_DEC | SEMI | ELSE | THEN | DO | RIGHT_BRACKET | RIGHT_PAREN | COMMA)
		return ERR
	}
}

func (parser *Parser) simple_expression() AttributeType {
	if parser.accept(ID|NUM) || parser.accept(LEFT_PAREN|NOT) {
		termType := parser.term()
		nextExprType := parser.simple_expression_prime(termType)

		if nextExprType == NULL {
			return termType
		}

		if nextExprType == ERR {
			return ERR
		}

		return termType
	} else if parser.accept(ADD) || parser.accept(SUB) {
		parser.sign()

		termType := parser.term()
		nextExprType := parser.simple_expression_prime(termType)

		errMsg := "Cannot use a sign on non-integers or non-reals"
		if parser.CheckType(termType, INT|REAL, errMsg) {
			return ERR
		}

		if nextExprType == NULL {
			return termType
		}

		if nextExprType == ERR {
			return ERR
		}

		return termType

		// if term == INT || term == REAL {
		// 	term -> simple_expression_prime
		// } else {
		// 	ERR_STAR -> simple_expression_prime
		// 	fmt.Print("A sign can only be used with an integer or real")
		// }
		// return simple_expression_prime
	}
	return ERR
}

func (parser *Parser) simple_expression_prime(typeName AttributeType) AttributeType {
	if parser.accept(ADDOP) {
		parser.expect(ADDOP)

		termType := parser.term()
		nextExprType := parser.simple_expression_prime(termType)

		if nextExprType == NULL {
			return termType
		}

		return termType

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
		// return simple_express_prime
	} else if parser.accept(RELOP) || parser.accept(END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA) {
		// NOOP
	} else {
		// ERROR
		parser.printError("+", "<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(RELOP, END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA)
	}
	return NULL
}

func (parser *Parser) term() AttributeType {
	factorType := parser.factor()
	termType := parser.term_prime()

	if termType == NULL {
		return factorType
	}

	if factorType != termType {
		parser.listing.AddTypeError("MULOP type mismatch")
		return ERR
	}

	return termType

	// factor -> term_prime
	// return term_prime
}

func (parser *Parser) term_prime() AttributeType {
	if parser.accept(MULOP) {
		parser.expect(MULOP)
		factorType := parser.factor()
		return factorType

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
		// return factor
	} else if parser.accept(ADDOP|RELOP) || parser.accept(END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA) {
		// NOOP
		return NULL
		// return term_prime
	} else {
		// ERROR
		parser.printError("*", "+", "<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(ADDOP|RELOP, END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA)
		return ERR
	}
}

func (parser *Parser) factor() AttributeType {
	if parser.accept(NUM) {
		num := parser.expect(NUM)
		return num.Attr()
	} else if parser.accept(LEFT_PAREN) {
		parser.expect(LEFT_PAREN)
		expression := parser.expression()
		parser.expect(RIGHT_PAREN)
		// return NULL
		return expression
	} else if parser.accept(ID) {
		id := parser.expect(ID)

		blueNode, err := parser.scope.GetTop().FindBlueNode(id.Value())
		if err != nil {
			parser.listing.AddTypeError("Variable already exists with same type in scope")
			return ERR
		}

		sym := blueNode.GetSymbol()
		symType := sym.GetType()

		switch symType {
		case PPINT:
			symType = INT
		case PPREAL:
			symType = REAL
		}

		factor_prime := parser.factor_prime(symType)

		if symType == factor_prime {
			return factor_prime
		}

		if symType == symType&(INT|AINT) {
			if factor_prime != factor_prime&(INT|AINT) {
				return ERR
			}
			return INT
		} else if symType == symType&(REAL|AREAL) {
			if factor_prime != factor_prime&(REAL|AREAL) {
				return ERR
			}
			return REAL
		}

		return ERR
	} else if parser.accept(NOT) {
		parser.expect(NOT)
		factor := parser.factor()

		if factor == BOOL {
			return BOOL
		} else {
			return ERR
		}
	} else {
		// ERROR
		parser.printError("a number", "(", "an identifier", "not")
		parser.sync(LEFT_PAREN|NOT, ID)
		return ERR
	}
}

func (parser *Parser) factor_prime(prevType AttributeType) AttributeType {
	if parser.accept(LEFT_BRACKET) {
		parser.expect(LEFT_BRACKET)
		expression := parser.expression()
		parser.expect(RIGHT_BRACKET)

		return expression
	} else if parser.accept(ADDOP|MULOP|RELOP) || parser.accept(END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA) {
		// NOOP

		return prevType

		// if sym.GetType() == PPINT {
		// 	return INT
		// } else if sym.GetType() == PPREAL {
		// 	return REAL
		// } else if sym.GetType() == PPAINT {
		// 	return AINT
		// } else if sym.GetType() == PPAREAL {
		// 	return AREAL
		// } else {
		// 	return sym.GetType()
		// }
	} else {
		// ERROR
		parser.printError("[", "*", "+", "<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(ADDOP|MULOP|RELOP, END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA)
		return ERR
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
