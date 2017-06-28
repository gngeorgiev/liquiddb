package liquiddb

import (
	"reflect"
	"testing"
)

func TestNewNode(t *testing.T) {
	n := newNode(TreeRoot, &Node{
		Key:      "ParentTest",
		Children: map[string]*Node{},
		Parent:   nil,
		pristine: false,
		Value:    nil,
	})

	if n.Key != TreeRoot {
		t.Fatalf("Invalid node key %s", n.Key)
	}

	if n.Parent == nil || n.Parent.Key != "ParentTest" {
		t.Fatalf("Invalid parent %+v", n.Parent)
	}
}

func TestNew(t *testing.T) {
	tree := newTree()

	if tree.root.Key != TreeRoot {
		t.Fatalf("Invalid tree root %s", tree.root.Key)
	}
}

func TestTree_SetInsert(t *testing.T) {
	tree := New()

	opInfos, err := tree.Set(data)
	if err != nil {
		t.Fatal(err)
	}

	op := opInfos[1]

	if op.Key != "bar" || op.Operation != EventOperationInsert ||
		!reflect.DeepEqual(op.Path, p) || !reflect.DeepEqual(op.Value, b) {
		t.Fatalf("Invalid op %+v", op)
	}
}

func TestTree_SetUpdate(t *testing.T) {
	tree := New()

	tree.Set(data)
	opInfos, err := tree.Set(data)
	if err != nil {
		t.Fatal(err)
	}

	op := opInfos[1]

	if op.Key != "bar" || op.Operation != EventOperationUpdate ||
		!reflect.DeepEqual(op.Path, p) || !reflect.DeepEqual(op.Value, b) {
		t.Fatalf("Invalid op %+v", op)
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
			t.Fatalf("Invalid op %+v", op)
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

	node := tree.root.Children["foo"]

	tree.Delete([]string{})

	children := len(tree.root.Children)
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
		t.Fatalf("Invalid data %+v", data)
	}
}

func TestTree_GetNonExisting(t *testing.T) {
	tree := New()

	v, _ := tree.Get(p)
	if v.Value != nil {
		t.Fatal("Wrong non existing value")
	}
}

func TestTree_GetJson(t *testing.T) {
	tree := New()
	//settings another value fist, to make sure we are overriding it correctly
	tree.Set(map[string]interface{}{
		"foo": true,
	})

	tree.Set(data)

	data, err := tree.Get([]string{"foo"})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data.Value, map[string]interface{}{
		"bar": []byte("foobar"),
	}) {
		t.Fatal("Incorrect json value")
	}
}
