package symbol

import (
	"github.com/Glorforidor/didactic_compiler/types"
)

type SymbolScope int

const (
	GlobalScope SymbolScope = iota
)

type Symbol struct {
	Name  string      // The identifier.
	Type  types.Type  // The type of a symbol.
	Scope SymbolScope // In which scope is the identifier.
}

// Code returns the assembly code for the symbol.
func (s *Symbol) Code() string {
	switch s.Scope {
	case GlobalScope:
		// In the global scope it is always just the name of the symbol as it
		// refers to a label.
		return s.Name
	default:
		return ""
	}
}

type Table struct {
	store map[string]Symbol
}

func NewTable() *Table {
	return &Table{
		store: make(map[string]Symbol),
	}
}

func (st *Table) Define(name string, t types.Type) Symbol {
	s := Symbol{Name: name, Scope: GlobalScope, Type: t}
	st.store[name] = s
	return s
}

func (st *Table) Resolve(name string) (Symbol, bool) {
	s, ok := st.store[name]
	return s, ok
}
