package scanner

type Token struct {
	id     TokenType
	attr   Attribute
	lexeme string
}

type TokenType int
type Attribute int

// Tokens
const (
	_ TokenType = iota
	RES
	PROG
	VAR
	OF
	INT_DEC
	REAL_DEC
	FUNC
	PROC
	BEGIN
	END_DEC
	IF
	THEN
	ELSE
	WHILE
	DO
	AND
	OR
	NOT
	MOD
	ID
	WS
	LONG_REAL
	REAL
	INT
	ASSIGNOP
	RELOP
	ADDOP
	MULOP
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACKET
	RIGHT_BRACKET
	COMMA
	SEMI
	COLON
	END
	LEXERR
)

// Attributes
const (
	_ Attribute = iota
	NULL
	UNREC
	EXTRA_LONG_INT
	EXTRA_LONG_FRAC
	NOT_EQ
	LESS_EQ
	GREATER_EQ
	EQ
	LESS
	GREATER
	ADD
	SUB
	MUL
	DIV
)

var TokenStrings []string = []string{
	RES:           "RES",
	PROG:          "PROG",
	VAR:           "VAR",
	OF:            "OF",
	INT_DEC:       "INT_DEC",
	REAL_DEC:      "REAL_DEC",
	FUNC:          "FUNC",
	PROC:          "PROC",
	BEGIN:         "BEGIN",
	END_DEC:       "END_DEC",
	IF:            "IF",
	THEN:          "THEN",
	ELSE:          "ELSE",
	WHILE:         "WHILE",
	DO:            "DO",
	AND:           "AND",
	OR:            "OR",
	NOT:           "NOT",
	MOD:           "MOD",
	ID:            "ID",
	WS:            "WS",
	LONG_REAL:     "LONG_REAL",
	REAL:          "REAL",
	INT:           "INT",
	ASSIGNOP:      "ASSIGNOP",
	RELOP:         "RELOP",
	ADDOP:         "ADDOP",
	MULOP:         "MULOP",
	LEFT_PAREN:    "LEFT_PAREN",
	RIGHT_PAREN:   "RIGHT_PAREN",
	LEFT_BRACKET:  "LEFT_BRACKET",
	RIGHT_BRACKET: "RIGHT_BRACKET",
	COMMA:         "COMMA",
	SEMI:          "SEMI",
	COLON:         "COLON",
	END:           "END",
	LEXERR:        "LEXERR",
}

var AttrStrings []string = []string{
	NULL:            "NULL",
	UNREC:           "UNREC",
	EXTRA_LONG_INT:  "EXTRA_LONG_INT",
	EXTRA_LONG_FRAC: "EXTRA_LONG_FRAC",
	NOT_EQ:          "NOT_EQ",
	LESS_EQ:         "LESS_EQ",
	GREATER_EQ:      "GREATER_EQ",
	EQ:              "EQ",
	LESS:            "LESS",
	GREATER:         "GREATER",
	ADD:             "ADD",
	SUB:             "SUB",
	MUL:             "MUL",
	DIV:             "DIV",
}

func (tok Token) Type() TokenType {
	return tok.id
}

func (tok Token) Value() string {
	return tok.lexeme
}

func (tok Token) String() string {
	if tok.id < 0 || int(tok.id) >= len(TokenStrings) {
		return "Unknown"
	}

	if tok.attr < 0 || int(tok.attr) >= len(AttrStrings) {
		return "Unknown"
	}

	return "\"" + tok.lexeme + "\"" + " " + TokenStrings[tok.id] + " " + AttrStrings[tok.attr]
}
