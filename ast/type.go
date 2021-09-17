package ast

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
	return [...]string{"Unknown", "Int", "Float", "String"}[k]
}
