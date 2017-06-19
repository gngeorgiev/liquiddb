package sstore

import (
	"testing"
)

func TestNewNode(t *testing.T) {
	n := newNode(TreeRoot, &Node{
		Key:      "ParentTest",
		Children: nil,
		Parent:   nil,
		pristine: false,
		Value:    nil,
	})

	if n.Key != TreeRoot {
		t.Fatalf("Invalid node key %s", n.Key)
	}

	if n.Parent == nil || n.Parent.Key != "ParentTest" {
		t.Fatalf("Invalid parent %s", n.Parent)
	}
}
