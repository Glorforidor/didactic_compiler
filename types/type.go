package types

type kind int

const (
	Unknown kind = iota
	Int
	Float
	String
	Bool
)

type Type struct {
	Kind kind
}

func (k kind) String() string {
	return [...]string{"unknown", "int", "float", "string", "bool"}[k]
}
