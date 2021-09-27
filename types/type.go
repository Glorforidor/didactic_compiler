package types

type kind int

const (
	Unknown kind = iota
	Int
	Float
	String
)

type Type struct {
	Kind kind
}

func (k kind) String() string {
	return [...]string{"unknown", "int", "float", "string"}[k]
}
