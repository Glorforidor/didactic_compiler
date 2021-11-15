package types

import (
	"strings"
)

type kind int

const (
	Unknown kind = iota
	Nil
	Int
	Float
	String
	Bool
	StructKind
	Func
)

type Type interface {
	Kind() kind
	String() string
}

type Basic struct {
	kind kind
	name string
}

func (b *Basic) Kind() kind     { return b.kind }
func (b *Basic) String() string { return b.name }

var Typ = []*Basic{
	Unknown: {kind: Unknown, name: "unknown"},
	Nil:     {kind: Nil, name: "nil"},
	Int:     {kind: Int, name: "int"},
	Float:   {kind: Float, name: "float"},
	String:  {kind: String, name: "string"},
	Bool:    {kind: Bool, name: "bool"},
}

type Signature struct {
	Parameter Type
	Result    Type
}

func (s *Signature) Kind() kind { return Func }
func (s *Signature) String() string {
	var sb strings.Builder

	sb.WriteString("(")
	if s.Parameter != nil {
		sb.WriteString(s.Parameter.String())
	}
	sb.WriteString(",")
	if s.Result != nil {
		sb.WriteString(s.Result.String())
	}
	sb.WriteString(")")

	return sb.String()
}

type Field struct {
	Name string
	Type Type
}

func (f *Field) String() string {
	var sb strings.Builder

	sb.WriteString(f.Name)
	sb.WriteString(" ")
	sb.WriteString(f.Type.String())

	return sb.String()
}

type Struct struct {
	Fields []*Field
}

func (s *Struct) Kind() kind { return StructKind }
func (s *Struct) String() string {
	var sb strings.Builder
	sb.WriteString("struct")
	sb.WriteString("{")
	for i, f := range s.Fields {
		sb.WriteString(f.String())

		if i == len(s.Fields)-1 {
			break
		}

		sb.WriteString(";")
	}
	sb.WriteString("}")

	return sb.String()
}
