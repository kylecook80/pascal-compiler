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
str << "type TokenType int"
str << "type Attribute int"
str << ""
str << "// Tokens"
str << "const ("
str << "_ TokenType = iota"

tokens.each do |token|
  str << "#{token}"
end

str << ")"
str << ""
str << "// Attributes"
str << "const ("
str << "_ Attribute = iota"

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

File.open("tokens.go", "w") do |f|
  f.puts str
end

`go fmt`
