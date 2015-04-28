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
	memory    *MemoryOffsetList
	tokenFile []byte
	tok       Token
	scope     *ScopeTree
	delay     bool
	newline   bool
}

func NewParser(scanner *Scanner) Parser {
	return Parser{scanner: scanner}
}

func (parser *Parser) Begin(file string) {
	listing := NewListingFile()
	tokenFile := []byte{}
	source := ReadFile(file)

	parser.listing = listing
	parser.memory = NewMemoryOffsetList()
	parser.source = source
	parser.tokenFile = tokenFile

	parser.scope = NewScopeTree()

	parser.program()

	parser.scope.GetRoot().GetMemoryOffset(parser.memory)
	parser.memory.WriteMemoryOffsetFile()

	// ioutil.WriteFile(GenerateTimeString(time.Now())+"_token_file.txt", parser.tokenFile, 0644)
	ioutil.WriteFile("token_file.txt", parser.tokenFile, 0644)
	parser.scanner.SymbolTable().Write()
	parser.memory.WriteMemoryOffsetFile()
	parser.listing.Save()
}

func (parser *Parser) nextTok() {
	tok, err := parser.scanner.NextToken()
	parser.tok = tok

	if tok.Attr() == NEWLINE {
		if parser.newline == true {
			parser.listing.AddLine("")
			parser.newline = false
		} else {
			parser.newline = true
		}
	} else {
		parser.newline = false
	}

	if parser.delay == true {
		parser.listing.AddLine(parser.source.ReadLine(parser.scanner.CurrentLineNumber()))
		parser.delay = false
	}

	if parser.scanner.CurrentLineNumber() >= parser.listing.LineCount() {
		parser.delay = true
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

	parser.expect(LEFT_PAREN)
	parser.identifier_list()
	parser.expect(RIGHT_PAREN)
	parser.expect(SEMI)

	parser.program_prime()

	parser.scope.Pop()
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
		parser.printError("var", "procedure", "begin")
		parser.sync(EOF)
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
		parser.printError("procedure", "begin")
		parser.sync(EOF)
	}
}

func (parser *Parser) identifier_list() {
	progParm := parser.expect(ID)

	symbol := NewSymbol(progParm.Value(), PGPARM)
	parser.scanner.SymbolTable().AddSymbol(symbol)
	parser.scope.GetTop().AddBlueNode(progParm.Value(), symbol, 0)

	parser.identifier_list_prime()
}

func (parser *Parser) identifier_list_prime() {
	if parser.accept(COMMA) {
		parser.expect(COMMA)

		progParm := parser.expect(ID)
		symbol := NewSymbol(progParm.Value(), PGPARM)
		parser.scanner.SymbolTable().AddSymbol(symbol)
		parser.scope.GetTop().AddBlueNode(progParm.Value(), symbol, 0)

		parser.identifier_list_prime()
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
	id := parser.expect(ID)
	parser.expect(COLON)

	typeName, length := parser.type_prod(id.Value())
	symbol := NewSymbol(id.Value(), typeName)
	parser.scanner.SymbolTable().AddSymbol(symbol)
	err := parser.scope.GetTop().AddBlueNode(id.Value(), symbol, length)
	if err != nil {
		parser.listing.AddScopeError("Variable " + id.Value() + " already declared")
	}

	parser.expect(SEMI)
	parser.declarations_prime()
}

func (parser *Parser) declarations_prime() {
	if parser.accept(VAR) {
		parser.expect(VAR)
		id := parser.expect(ID)
		parser.expect(COLON)

		typeName, length := parser.type_prod(id.Value())
		symbol := NewSymbol(id.Value(), typeName)
		parser.scanner.SymbolTable().AddSymbol(symbol)
		err := parser.scope.GetTop().AddBlueNode(id.Value(), symbol, length)
		if err != nil {
			parser.listing.AddScopeError("Variable " + id.Value() + " already declared")
		}

		parser.expect(SEMI)
		parser.declarations_prime()
	} else if parser.accept(PROC | BEGIN) {
		// NOOP
	} else {
		// ERROR
		parser.printError("var", "procedure", "begin")
		parser.sync(PROC | BEGIN)
	}
}

func (parser *Parser) type_prod(id string) (AttributeType, int) {
	if parser.accept(INT_DEC | REAL_DEC) {
		standardType := parser.standard_type(id)
		var size int
		switch standardType {
		case INT:
			size = 4
		case REAL:
			size = 8
		}
		return standardType, size
	} else if parser.accept(ARRAY) {
		parser.expect(ARRAY)
		parser.expect(LEFT_BRACKET)

		num1 := parser.expect(NUM)
		if parser.CheckType(num1.Attr(), INT, "Array index type mismatch") {
			return ERR, 0
		}

		parser.expect(RANGE)

		num2 := parser.expect(NUM)
		if parser.CheckType(num2.Attr(), INT, "Array index type mismatch") {
			return ERR, 0
		}

		num1Val, _ := strconv.Atoi(num1.Value())
		num2Val, _ := strconv.Atoi(num2.Value())

		length := num2Val - num1Val + 1

		parser.expect(RIGHT_BRACKET)
		parser.expect(OF)

		standard_type := parser.standard_type(id)

		if standard_type == PPINT {
			return PPAINT, 0
		} else if standard_type == PPREAL {
			return PPAREAL, 0
		} else if standard_type == INT {
			return AINT, 4 * length
		} else if standard_type == REAL {
			return AREAL, 8 * length
		} else {
			return ERR, 0
		}
	} else {
		// ERROR
		parser.printError("integer", "real", "array")
		parser.sync(ARRAY)
		return ERR, 0
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
	parser.subprogram_declaration()
	parser.expect(SEMI)
	parser.subprogram_declarations_prime()
}

func (parser *Parser) subprogram_declarations_prime() {
	if parser.accept(PROC) {
		parser.subprogram_declaration()
		parser.expect(SEMI)
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
	parser.subprogram_head()
	parser.subprogram_declaration_prime()
	parser.scope.Pop()
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

	procName := parser.expect(ID)
	greenNode := parser.scope.GetTop().FindGreenNode(procName.Value())

	if greenNode != nil {
		parser.listing.AddScopeError("Procedure " + procName.Value() + " already exists")
	}

	symbol := NewSymbol(procName.Value(), PROC)
	parser.scanner.SymbolTable().AddSymbol(symbol)
	parser.scope.AddGreenNode(procName.Value(), symbol)

	parser.subprogram_head_prime()
}

func (parser *Parser) subprogram_head_prime() {
	if parser.accept(LEFT_PAREN) {
		parser.arguments()
		parser.expect(SEMI)
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
	id := parser.expect(ID)
	parser.expect(COLON)
	typeName, length := parser.type_prod(id.Value())

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

	greenNode := parser.scope.GetTop()
	greenNode.AddBlueNode(id.Value(), symbol, length)
	greenNode.IncParam()

	parser.parameter_list_prime()
}

func (parser *Parser) parameter_list_prime() {
	if parser.accept(SEMI) {
		parser.expect(SEMI)
		id := parser.expect(ID)
		parser.expect(COLON)
		typeName, length := parser.type_prod(id.Value())

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

		greenNode := parser.scope.GetTop()
		greenNode.AddBlueNode(id.Value(), symbol, length)
		greenNode.IncParam()

		parser.parameter_list_prime()
	} else if parser.accept(RIGHT_PAREN) {
		// NOOP
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
		// NOOP
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

		if variable == ERR || expression == ERR {
			return ERR
		}

		if parser.CheckType(variable, expression, "ASSIGNOP type mismatch") {
			return ERR
		}
	} else if parser.accept(CALL) {
		parser.procedure_statement()
	} else if parser.accept(BEGIN) {
		parser.compound_statement()
	} else if parser.accept(IF) {
		parser.expect(IF)

		expression := parser.expression()
		if expression != ERR {
			parser.CheckType(expression, BOOL, "Only boolean expressions are allowed in if statements")
		}

		parser.expect(THEN)
		parser.statement()
		parser.statement_prime()
	} else if parser.accept(WHILE) {
		parser.expect(WHILE)

		expression := parser.expression()
		if expression != ERR {
			parser.CheckType(expression, BOOL, "Only boolean expressions are allowed in while statements")
		}

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

	blueNode, err := parser.scope.GetTop().FindBlueNode(id.Value())
	if err != nil {
		parser.listing.AddScopeError("Could not find variable " + id.Value())
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
	id := parser.expect(ID)

	greenNode := parser.scope.GetTop()
	calledProc := greenNode.FindGreenNode(id.Value())

	if calledProc == nil {
		parser.listing.AddScopeError("Procedure " + id.Value() + " not found")
	}

	parser.procedure_statement_prime(calledProc)
}

func (parser *Parser) procedure_statement_prime(proc *GreenNode) {
	if parser.accept(LEFT_PAREN) {
		parser.expect(LEFT_PAREN)
		parser.expression_list(proc)
		parser.expect(RIGHT_PAREN)
	} else if parser.accept(END_DEC | SEMI | ELSE) {
		// NOOP
	} else {
		// ERROR
		parser.printError("(", "end", ";", "else")
		parser.sync(END_DEC | SEMI | ELSE)
	}
}

func (parser *Parser) expression_list(proc *GreenNode) AttributeType {
	// if proc == nil {
	// 	parser.expression()
	// 	parser.expression_list_prime(proc, 0)
	// 	return ERR
	// }

	if proc == nil {
		return ERR
	}

	vars := proc.GetVars()
	count := 0

	varType := NULL
	checkVar := vars[count].GetSymbol().GetType()

	expression := parser.expression()
	expression_list_prime := parser.expression_list_prime(proc, count)

	switch checkVar {
	case PPINT:
		varType = INT
	case PPREAL:
		varType = REAL
	case PPAINT:
		varType = AINT
	case PPAREAL:
		varType = AREAL
	}

	if parser.CheckType(varType, expression, "Types for parameter "+strconv.Itoa(count)+" in call to "+proc.GetName()+" do not match") {
		return ERR
	}

	if expression == NULL {
		params := proc.GetNumParams()
		if params != 0 {
			parser.listing.AddTypeError("Too few parameters for call to " + proc.GetName())
			return ERR
		}
	}

	if expression_list_prime == NULL {
		params := proc.GetNumParams()
		fmt.Println(params)
		if params > 1 {
			parser.listing.AddTypeError("Too few parameters for call to " + proc.GetName())
			return ERR
		} else if params < 1 {
			parser.listing.AddTypeError("Too many parameters for call to " + proc.GetName())
			return ERR
		}
	}

	// if expression_list_prime == ERR {
	// 	return ERR
	// }

	return checkVar
}

func (parser *Parser) expression_list_prime(proc *GreenNode, count int) AttributeType {
	if parser.accept(COMMA) {
		parser.expect(COMMA)

		count += 1

		vars := proc.GetVars()
		varType := NULL
		checkVar := vars[count].GetSymbol().GetType()

		expression := parser.expression()
		expression_list_prime := parser.expression_list_prime(proc, count)

		switch checkVar {
		case PPINT:
			varType = INT
		case PPREAL:
			varType = REAL
		case PPAINT:
			varType = AINT
		case PPAREAL:
			varType = AREAL
		}

		if expression_list_prime == NULL {
			params := proc.GetNumParams()
			fmt.Println(params)
			fmt.Println()
			fmt.Println(count + 1)
			if params > (count + 1) {
				parser.listing.AddTypeError("Too few parameters for call to " + proc.GetName())
				return ERR
			} else if params < (count + 1) {
				parser.listing.AddTypeError("Too many parameters for call to " + proc.GetName())
				return ERR
			}
		}

		if parser.CheckType(varType, expression, "Types for parameter "+strconv.Itoa(count)+" in call to "+proc.GetName()+" do not match") {
			return ERR
		}

		return varType
	} else if parser.accept(RIGHT_PAREN) {
		// NOOP
		return NULL
	} else {
		// ERROR
		parser.printError(",", ")")
		parser.sync(RIGHT_PAREN)
		return ERR
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
		termType := parser.term(NULL, MULOP)
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

		termType := parser.term(NULL, ADDOP)
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
	}

	return ERR
}

func (parser *Parser) simple_expression_prime(typeName AttributeType) AttributeType {
	if parser.accept(ADDOP) {
		parser.expect(ADDOP)

		termType := parser.term(typeName, ADDOP)
		nextExprType := parser.simple_expression_prime(termType)

		if nextExprType == NULL {
			return typeName
		}

		return termType
	} else if parser.accept(RELOP) || parser.accept(END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA) {
		// NOOP
		return typeName
	} else {
		// ERROR
		parser.printError("+", "<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(RELOP, END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA)
		return ERR
	}
}

func (parser *Parser) term(typeName AttributeType, op TokenType) AttributeType {
	factorType := parser.factor(typeName, op)
	termType := parser.term_prime(typeName, op)

	if termType == NULL {
		return factorType
	}

	if factorType == ERR || termType == ERR {
		return ERR
	}

	if op == MULOP {
		if parser.CheckType(factorType, termType, "MULOP type mismatch") {
			return ERR
		}
	} else {
		return ERR
	}

	return termType
}

func (parser *Parser) term_prime(typeName AttributeType, op TokenType) AttributeType {
	if parser.accept(MULOP) {
		parser.expect(MULOP)
		factorType := parser.factor(typeName, op)
		return factorType
	} else if parser.accept(ADDOP|RELOP) || parser.accept(END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA) {
		// NOOP
		return NULL
	} else {
		// ERROR
		parser.printError("*", "+", "<", "<=", ">", ">=", "=", "end", ";", "else", "then", "do", "]", ")", ",")
		parser.sync(ADDOP|RELOP, END_DEC|SEMI|ELSE|THEN|DO|RIGHT_BRACKET|RIGHT_PAREN|COMMA)
		return ERR
	}
}

func (parser *Parser) factor(typeName AttributeType, op TokenType) AttributeType {
	if parser.accept(NUM) {
		num := parser.expect(NUM)

		if op == ADDOP && parser.CheckType(typeName, num.Attr(), "ADDOP type mismatch") {
			return ERR
		}

		return num.Attr()
	} else if parser.accept(LEFT_PAREN) {
		parser.expect(LEFT_PAREN)
		expression := parser.expression()
		parser.expect(RIGHT_PAREN)

		return expression
	} else if parser.accept(ID) {
		id := parser.expect(ID)

		blueNode, err := parser.scope.GetTop().FindBlueNode(id.Value())
		if err != nil {
			parser.listing.AddScopeError("Could not find variable " + id.Value())
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
		factor := parser.factor(typeName, op)

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
