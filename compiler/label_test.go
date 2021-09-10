package compiler

import (
	"testing"
)

func TestCreateAndName(t *testing.T) {
	var l label

	l.Create()
	if l.Name() != ".L1" {
		t.Fatalf("label have wrong name: expected=%q, got=%q", ".L1", l.Name())
	}

	l.Create()
	if l.Name() != ".L2" {
		t.Fatalf("label have wrong name: expected=%q, got=%q", ".L2", l.Name())
	}
}
