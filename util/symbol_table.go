package util

import _ "container/list"
import "fmt"
import "errors"
import "os"

type SymbolTable struct {
	list []*Symbol
}

type Symbol struct {
	name     string
	typeName AttributeType
	size     int
	value    *interface{}
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{make([]*Symbol, 0)}
}

func NewSymbol(name string, typeName AttributeType) *Symbol {
	return &Symbol{name: name, typeName: typeName}
}

// SYMBOL TABLE

func (st *SymbolTable) GetPtr(name string, typeAttr AttributeType) (*Symbol, error) {
	for _, symbol := range st.list {
		if symbol == nil {
			return &Symbol{}, errors.New("Symbol not found.")
		} else if name == symbol.name && symbol.GetType() == typeAttr {
			return symbol, nil
		}
	}

	return &Symbol{}, errors.New("Error")
}

func (st *SymbolTable) AddSymbol(sym *Symbol) error {
	name := sym.name
	symType := sym.GetType()
	_, err := st.GetPtr(name, symType)

	if err != nil {
		st.list = append(st.list, sym)
		return nil
	} else {
		return fmt.Errorf("Variable already declared")
	}
}

// func (st *SymbolTable) AssignType(id string, typeName AttributeType) {
// 	sym, err := st.GetPtr(id, typeName)
// 	if err != nil {
// 		fmt.Errorf("Error getting %s from symbol table", id)
// 	} else {
// 		sym.SetType(typeName)
// 		st.list.InsertAfter(sym, elem)
// 		st.list.Remove(elem)
// 	}
// }

func (st *SymbolTable) Print() {
	for _, symbol := range st.list {
		fmt.Println(symbol)
	}
	fmt.Println()
}

func (st *SymbolTable) String() string {
	var st_string string
	for _, symbol := range st.list {
		if symbol == nil {
			break
		}

		st_string += symbol.String()
	}
	st_string += "\n"
	return st_string
}

func (st *SymbolTable) Write() string {
	file := "symbol_file.txt"
	newFile, err := os.Create(file)
	if err != nil {
		return ""
	}
	defer newFile.Close()

	newFile.Write([]byte(st.String()))
	return file
}

// SYMBOL

func (sym *Symbol) GetType() AttributeType {
	return sym.typeName
}

func (sym *Symbol) SetType(typeName AttributeType) {
	sym.typeName = typeName
}

func (sym *Symbol) GetSize() int {
	return sym.size
}

func (sym *Symbol) SetSize(newSize int) {
	sym.size = newSize
}

func (sym *Symbol) String() string {
	return fmt.Sprintln(sym.name, sym.typeName, sym.value)
}
