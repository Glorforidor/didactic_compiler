package compiler

import "fmt"

type label struct {
	num int
}

// Create creates a new label.
func (l *label) Create() {
	l.num++
}

// Name returns the assembly name of the label.
func (l *label) Name() string {
	return fmt.Sprintf(".L%d", l.num)
}
