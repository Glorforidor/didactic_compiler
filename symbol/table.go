package symbol

import (
	"fmt"

	"github.com/Glorforidor/didactic_compiler/types"
)

type SymbolScope int

const (
	GlobalScope SymbolScope = iota
	LocalScope
)

type Symbol struct {
	Name  string      // The identifier.
	Type  types.Type  // The type of a symbol.
	Scope SymbolScope // In which scope is the identifier.
	Which int         // The ordinal position of variable (local or param)
}

// Code returns the assembly code for the symbol.
func (s *Symbol) Code() string {
	switch s.Scope {
	case GlobalScope:
		// In the global scope it is always just the name of the symbol as it
		// refers to a label.
		return s.Name
	case LocalScope:
		return fmt.Sprintf("%d(sp)", s.Which*8)
	default:
		panic("Symbol did not have a scope!")
	}
}

type Table struct {
	Outer *Table
	store map[string]Symbol

	NumDefinitions int
}

func NewTable() *Table {
	return &Table{
		store: make(map[string]Symbol),
	}
}

func NewEnclosedTable(outer *Table) *Table {
	s := NewTable()
	s.Outer = outer
	return s
}

func (st *Table) Define(name string, t types.Type) (Symbol, error) {
	if s, ok := st.store[name]; ok {
		// TODO: better error message - what scope? maybe just say the variable
		// is already declared.
		return s, fmt.Errorf("identifier: %q already defined in scope", name)
	}

	s := Symbol{Name: name, Type: t, Which: st.NumDefinitions}
	if st.Outer == nil {
		s.Scope = GlobalScope
	} else {
		s.Scope = LocalScope
	}
	st.store[name] = s
	st.NumDefinitions++

	return s, nil
}

func (st *Table) Resolve(name string) (Symbol, bool) {
	s, ok := st.store[name]
	if !ok && st.Outer != nil {
		s, ok := st.Outer.Resolve(name)
		return s, ok
	}
	return s, ok
}
