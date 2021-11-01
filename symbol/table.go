package symbol

import (
	"fmt"
	"math"
)

type SymbolScope int

const (
	GlobalScope SymbolScope = iota
	LocalScope
)

func (ss SymbolScope) String() string {
	return [...]string{"Global", "Local"}[ss]
}

type Symbol struct {
	Name        string      // The identifier name.
	Type        interface{} // The token type assoicated to the identifier.
	Scope       SymbolScope // In which scope is the identifier.
	stackOffset int         // stackOffset is used for referencing variables from the previous scope
	which       int         // The ordinal position of variable (local or param)
}

// variablesSize is the byte size of a variable. As we target RV64 then it is 8
// bytes.
const variableSize = 8

// Code returns the assembly code for the symbol.
func (s *Symbol) Code() string {
	switch s.Scope {
	case GlobalScope:
		// In the global scope it is always just the name of the symbol as it
		// refers to a label.
		return s.Name
	case LocalScope:
		return fmt.Sprintf("%d(sp)", s.stackOffset+(s.which+1)*variableSize)
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

func (st *Table) Define(name string, t interface{}) (Symbol, error) {
	if s, ok := st.store[name]; ok {
		// TODO: better error message - what scope? maybe just say the variable
		// is already declared.
		return s, fmt.Errorf("identifier: %q already defined in scope", name)
	}

	s := Symbol{Name: name, Type: t, which: st.NumDefinitions}
	if st.Outer == nil {
		s.Scope = GlobalScope
	} else {
		// Do not allow variable shadowing.
		if s, ok := st.Resolve(name); ok {
			return s, fmt.Errorf("identifier: %q would over shadow existing identfier", name)
		}
		s.Scope = LocalScope
	}
	st.store[name] = s
	st.NumDefinitions++

	return s, nil
}

func (st *Table) Resolve(name string) (Symbol, bool) {
	return st.resolve(name, 0)
}

func (st *Table) resolve(name string, stackOffset int) (Symbol, bool) {
	s, ok := st.store[name]
	if !ok && st.Outer != nil {
		s, ok := st.Outer.resolve(name, stackOffset+st.StackSpace())
		return s, ok
	}

	// No need to add stack offset when it is global.
	if s.Scope != GlobalScope {
		s.stackOffset = stackOffset
	}
	return s, ok
}

func (st *Table) StackSpace() int {
	x := math.Round(float64(st.NumDefinitions) / 2)
	y := x * 16
	return int(y)
}
