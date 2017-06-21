package liquiddb

import (
	"reflect"
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

func TestNew(t *testing.T) {
	tree := newTree()

	if tree.Root.Key != TreeRoot {
		t.Fatalf("Invalid tree root %s", tree.Root.Key)
	}
}

func TestTree_SetInsert(t *testing.T) {
	tree := New()

	opInfos, err := tree.Set(data)
	if err != nil {
		t.Fatal(err)
	}

	op := opInfos[0]

	if op.Key != "bar" || op.Operation != EventOperationInsert ||
		!reflect.DeepEqual(op.Path, p) || !reflect.DeepEqual(op.Value, b) {
		t.Fatalf("Invalid op %+s", op)
	}
}

func TestTree_SetUpdate(t *testing.T) {
	tree := New()

	tree.Set(data)
	opInfos, err := tree.Set(data)
	if err != nil {
		t.Fatal(err)
	}

	op := opInfos[0]

	if op.Key != "bar" || op.Operation != EventOperationUpdate ||
		!reflect.DeepEqual(op.Path, p) || !reflect.DeepEqual(op.Value, b) {
		t.Fatalf("Invalid op %+s", op)
	}
}

func TestTree_Delete(t *testing.T) {
	tree := New()
	tree.Set(data)

	op, ok := tree.Delete(p)
	if !ok {
		t.Fatalf("Invalid result after delete, should be true")
	}

	for _, op := range op {
		if op.Operation != EventOperationDelete || op.Key != "bar" ||
			!reflect.DeepEqual(op.Path, p) || !reflect.DeepEqual(op.Value, b) {
			t.Fatalf("Invalid op %+s", op)
		}
	}

	op, ok = tree.Delete(p)
	if ok {
		t.Fatalf("Invalid result after delete, should be false")
	}
}

func TestTree_DeleteAll(t *testing.T) {
	tree := newTree()

	tree.Set(data)

	node := tree.Root.Children["foo"]

	tree.Delete([]string{})

	children := len(tree.Root.Children)
	if children > 0 {
		t.Fatal("Invalid amount of children")
	}

	if node.Parent != nil || node.Value != nil {
		t.Fatal("Node is not deleted properly")
	}
}

func TestTree_Get(t *testing.T) {
	tree := New()
	tree.Set(data)

	data, err := tree.Get(p)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data.Value, b) {
		t.Fatalf("Invalid data %s", data)
	}
}

func TestTree_GetNonExisting(t *testing.T) {
	tree := New()

	_, err := tree.Get(p)
	if err.Error() != ErrNotFound.Error() {
		t.Fatal(err)
	}
}
