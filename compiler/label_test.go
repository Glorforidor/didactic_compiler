package compiler

import (
	"testing"
)

func TestCreateAndName(t *testing.T) {
	var l label

	if l.create() != ".L1" {
		t.Fatalf("label have wrong name: expected=%q, got=%q", ".L1", l.create())
	}

	if l.create() != ".L2" {
		t.Fatalf("label have wrong name: expected=%q, got=%q", ".L2", l.create())
	}
}
