package symbol

import (
	"fmt"
	"sort"
)

type SymbolScope int

const (
	GlobalScope SymbolScope = iota
	LocalScope
	TypeScope
	FuncScope
)

func (ss SymbolScope) String() string {
	return [...]string{"Global", "Local"}[ss]
}

type Symbol struct {
	Name        string      // The identifier name.
	Type        interface{} // The token type assoicated to the identifier.
	Scope       SymbolScope // In which scope is the identifier.
	which       int         // The ordinal position of variable (local or param)
	stackPoint  int         // Where the symbol resides on the stack.
	stackOffset int         // stackOffset is used for referencing variables from the previous scope
}

func (s *Symbol) String() string {
	return fmt.Sprintf("Name: %s, Type: %T", s.Name, s.Type)
}

// variablesSize is the byte size of a variable. As we target RV64 then it is 8
// bytes.
const variableSize = 8

func (s *Symbol) Code() interface{} {
	// Code returns the assembly code for the symbol.
	switch s.Scope {
	case GlobalScope:
		// In the global scope it is always just the name of the symbol as it
		// refers to a label.
		return s.Name
	case LocalScope:
		return s.stackOffset + s.stackPoint
	default:
		panic("Symbol did not have a scope!")
	}
}

type Table struct {
	Outer *Table

	store map[string]*Symbol

	numDefinitions int

	stackSpace int
}

func (st *Table) String() string {
	return fmt.Sprintf("%+v", st.store)
}

func NewTable() *Table {
	return &Table{
		store: make(map[string]*Symbol),
	}
}

func NewEnclosedTable(outer *Table) *Table {
	s := NewTable()
	s.Outer = outer
	return s
}

func (st *Table) DefineType(name string, t interface{}) (*Symbol, error) {
	if s, ok := st.store[name]; ok {
		// TODO: better error message - what scope? maybe just say the variable
		// is already declared.
		return s, fmt.Errorf("identifier: %q already defined in scope", name)
	}

	s := &Symbol{Name: name, Type: t, which: 0, Scope: TypeScope}
	st.store[name] = s

	return s, nil
}

func (st *Table) DefineFunc(name string, t interface{}) (*Symbol, error) {
	if s, ok := st.store[name]; ok {
		// TODO: better error message - what scope? maybe just say the variable
		// is already declared.
		return s, fmt.Errorf("identifier: %q already defined in scope", name)
	}

	s := &Symbol{Name: name, Type: t, which: 0, Scope: FuncScope}
	st.store[name] = s

	return s, nil
}

// DefineInFunc is used for defining the function parameter. Which will always
// have the local scope.
func (st *Table) DefineInFunc(name string, t interface{}) *Symbol {
	s := &Symbol{Name: name, Type: t, which: st.numDefinitions, Scope: LocalScope}
	st.store[name] = s
	st.numDefinitions++

	return s
}

// Define defines the name with type t into the symbol table. It will check
// that the variable does not over shadow a symbol with the same name.
func (st *Table) Define(name string, t interface{}) (*Symbol, error) {
	if s, ok := st.store[name]; ok {
		// TODO: better error message - what scope? maybe just say the variable
		// is already declared.
		return s, fmt.Errorf("identifier: %q already defined in scope", name)
	}

	s := &Symbol{Name: name, Type: t, which: st.numDefinitions}
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
	st.numDefinitions++

	return s, nil
}

// ComputeStack how much stack space each block will accomendate and also sets
// the symbols stack offset.
func (st *Table) ComputeStack() {
	keys := make([]string, 0, len(st.store))
	for k := range st.store {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return st.store[keys[i]].which < st.store[keys[j]].which
	})

	// The first element on the stack is always at 8(sp). Writing to 0(sp)
	// should always be safe and should never overwrite other data.
	x := variableSize

	for _, k := range keys {
		v := st.store[k]
		if v.Scope == TypeScope {
			continue
		}

		v.stackPoint = x
		x += variableSize
	}

	// The stack space always need to be 16 byte alligned. We subtract 8 from x
	// since we add that extra 8 as offset.
	if (x-variableSize)%16 == 0 {
		st.stackSpace = x - variableSize
	} else {
		st.stackSpace = x
	}
}

func (st *Table) Resolve(name string) (*Symbol, bool) {
	return st.resolve(name, 0)
}

func (st *Table) resolve(name string, stackOffset int) (*Symbol, bool) {
	s, ok := st.store[name]
	if !ok && st.Outer != nil {
		s, ok := st.Outer.resolve(name, stackOffset+st.StackSpace())
		return s, ok
	}

	// No need to add stack offset when it is global.
	if s != nil && s.Scope != GlobalScope {
		s.stackOffset = stackOffset
	}
	return s, ok
}

func (st *Table) StackSpace() int {
	return st.stackSpace
}
