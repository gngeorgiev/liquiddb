package sstore

type nodeChildren map[string]*Node

//Node is a node in the tree
type Node struct {
	Key      string
	Value    []byte
	Parent   *Node
	Children nodeChildren

	pristine bool
}

//newNode creates a new node in the tree
func newNode(key string, parent *Node) *Node {
	return &Node{
		key,
		nil,
		parent,
		nodeChildren{},
		true,
	}
}
