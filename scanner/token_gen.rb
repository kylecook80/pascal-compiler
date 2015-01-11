tokens = []
attrs = []

File.open("tokens", "r") do |f|
  f.each_line do |token|
    tokens << token.chomp
  end
end

File.open("attributes", "r") do |f|
  f.each_line do |attribute|
    attrs << attribute.chomp
  end
end

str = []
str << "package scanner"
str << ""
str << "type Token struct {"
str << "  id   TokenType"
str << "  attr Attribute"
str << "  lexeme string"
str << "}"
str << ""
str << "type TokenType uint"
str << "type Attribute uint"
str << ""
str << "// Tokens"
str << "const ("
str << "_ TokenType = 1 << iota"

tokens.each do |token|
  str << "#{token}"
end

str << ")"
str << ""
str << "// Attributes"
str << "const ("
str << "_ Attribute = 1 << iota"

attrs.each do |attribute|
  str << "#{attribute}"
end

str << ")"
str << ""
str << "var TokenStrings []string = []string{"

tokens.each do |token|
  str << "#{token}: \"#{token}\","
end

str << "}"
str << ""
str << "var AttrStrings []string = []string{"

attrs.each do |attribute|
  str << "#{attribute}: \"#{attribute}\","
end

str << "}"
str << ""

str << "func (tok Token) String() string {"
str << "if tok.id < 0 || int(tok.id) >= len(TokenStrings) {"
str << "return \"Unknown\""
str << "}"
str << ""
str << "if tok.attr < 0 || int(tok.attr) >= len(AttrStrings) {"
str << "return \"Unknown\""
str << "}"
str << ""
str << "return \"\\\"\" + tok.lexeme + \"\\\"\" + \" \" + TokenStrings[tok.id] + \" \" + AttrStrings[tok.attr]"
str << "}"
str << ""

str << "func (tok Token) Type() TokenType {"
str << "return tok.id"
str << "}"
str << ""

str << "func (tok Token) Attr() Attribute {"
str << "return tok.attr"
str << "}"
str << ""

str << "func (tok Token) Value() string {"
str << "return tok.lexeme"
str << "}"
str << ""

str << "func (tokType TokenType) String() string {"
str << "return TokenStrings[tokType]"
str << "}"
str << ""

str << "func (attr Attribute) String() string {"
str << "return AttrStrings[attr]"
str << "}"
str << ""

File.open("tokens.go", "w") do |f|
  f.puts str
end

`go fmt`
