package compiler

import "fmt"

type label struct {
	num int
}

// create returns a unique assembly label name.
func (l *label) create() string {
	l.num++
	return fmt.Sprintf(".L%d", l.num)
}
