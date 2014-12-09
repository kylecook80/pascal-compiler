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
	ID
	WS
	NUM
	RANGE
	ASSIGNOP
	RELOP
	ADDOP
	MULOP
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
	PROG
	VAR
	OF
	INT_DEC
	REAL_DEC
	PROC
	BEGIN
	END_DEC
	LONG_REAL
	REAL
	INT
	IF
	THEN
	ELSE
	WHILE
	DO
	AND
	OR
	NOT
	MOD
	ARRAY
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACKET
	RIGHT_BRACKET
	COMMA
	SEMI
	COLON
	END
)

var TokenStrings []string = []string{
	RES:      "RES",
	ID:       "ID",
	WS:       "WS",
	NUM:      "NUM",
	RANGE:    "RANGE",
	ASSIGNOP: "ASSIGNOP",
	RELOP:    "RELOP",
	ADDOP:    "ADDOP",
	MULOP:    "MULOP",
	LEXERR:   "LEXERR",
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
	PROG:            "PROG",
	VAR:             "VAR",
	OF:              "OF",
	INT_DEC:         "INT_DEC",
	REAL_DEC:        "REAL_DEC",
	PROC:            "PROC",
	BEGIN:           "BEGIN",
	END_DEC:         "END_DEC",
	LONG_REAL:       "LONG_REAL",
	REAL:            "REAL",
	INT:             "INT",
	IF:              "IF",
	THEN:            "THEN",
	ELSE:            "ELSE",
	WHILE:           "WHILE",
	DO:              "DO",
	AND:             "AND",
	OR:              "OR",
	NOT:             "NOT",
	MOD:             "MOD",
	ARRAY:           "ARRAY",
	LEFT_PAREN:      "LEFT_PAREN",
	RIGHT_PAREN:     "RIGHT_PAREN",
	LEFT_BRACKET:    "LEFT_BRACKET",
	RIGHT_BRACKET:   "RIGHT_BRACKET",
	COMMA:           "COMMA",
	SEMI:            "SEMI",
	COLON:           "COLON",
	END:             "END",
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

func (tok Token) Type() TokenType {
	return tok.id
}

func (tok Token) Attr() Attribute {
	return tok.attr
}

func (tok Token) Value() string {
	return tok.lexeme
}

func (tokType TokenType) String() string {
	return TokenStrings[tokType]
}

func (attr Attribute) String() string {
	return AttrStrings[attr]
}
