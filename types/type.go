package types

type kind int

const (
	Unknown kind = iota
	Int
	Float
	String
	Bool
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
	Int:     {kind: Int, name: "int"},
	Float:   {kind: Float, name: "float"},
	String:  {kind: String, name: "string"},
	Bool:    {kind: Bool, name: "bool"},
}

// TODO: fix the kind of signature and struct

type Signature struct{}

func (s *Signature) Kind() kind     { return -1 }
func (s *Signature) String() string { return "" }

type Struct struct{}

func (s *Struct) Kind() kind     { return -1 }
func (s *Struct) String() string { return "" }
