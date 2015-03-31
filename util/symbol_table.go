package util

import "container/list"
import "fmt"
import "errors"

type SymbolTable struct {
	list *list.List
}

type Symbol struct {
	name     string
	typeName AttributeType
	value    *interface{}
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{list.New()}
}

func NewSymbol(name string) Symbol {
	return Symbol{name: name}
}

func (sym Symbol) GetType() AttributeType {
	return sym.typeName
}

func (sym *Symbol) SetType(typeName AttributeType) {
	sym.typeName = typeName
}

func (st *SymbolTable) GetPtr(name string) (*list.Element, error) {
	for item := st.list.Front(); item != nil; item = item.Next() {
		symbol := item.Value.(Symbol)
		if name == symbol.name {
			return item, nil
		}
	}
	return new(list.Element), errors.New("Symbol not found.")
}

func (st *SymbolTable) AddSymbol(sym Symbol) bool {
	name := sym.name
	_, err := st.GetPtr(name)

	if err != nil {
		st.list.PushBack(sym)
		return true
	} else {
		return false
	}
}

func (st *SymbolTable) AssignType(id string, typeName AttributeType) {
	elem, err := st.GetPtr(id)
	sym := elem.Value.(Symbol)
	if err != nil {
		fmt.Errorf("Error getting %s from symbol table", id)
	} else {
		sym.SetType(typeName)
		st.list.InsertAfter(sym, elem)
		st.list.Remove(elem)
	}
}

func (st *SymbolTable) Print() {
	for item := st.list.Front(); item != nil; item = item.Next() {
		symbol := item.Value.(Symbol)
		fmt.Println(symbol)
	}
	fmt.Println()
}
