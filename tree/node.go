package tree

import (
	"github.com/go-errors/errors"
	"github.com/thoas/go-funk"
)

type nodeChildren map[string]*Node

//Node is a node in the tree
type Node struct {
	Key      string
	Value    []byte
	Parent   *Node
	Children nodeChildren
}

//NewNode creates a new node in the tree
func NewNode(key string, parent *Node) *Node {
	return &Node{
		key,
		nil,
		parent,
		nodeChildren{},
	}
}

func (n *Node) buildPathUpwards() []string {
	path := []string{n.Key}
	currentNode := n
	for {
		currentNode = currentNode.Parent
		if currentNode.Key == TREE_ROOT || currentNode.Parent == nil {
			break
		}

		p := currentNode.Key
		path = append(path, p)
	}

	return funk.ReverseStrings(path)
}

//Insert inserts a data into a specific key path - e.g. ["foo", "bar"] - "foo" is the same as "foo": { "bar": "foo" }
//Returns an OpInfo that holds info about the passed operation, if succesful.
func (n *Node) Insert(keys []string, data []byte) (OpInfo, error) {
	if len(keys) == 0 {
		n.Value = data
		return OpInfo{
			n.buildPathUpwards(),
			n.Key,
			n.Value,
		}, nil
	}

	key := keys[0]
	if existingNode, ok := n.Children[key]; ok {
		return existingNode.Insert(keys[1:], data)
	}

	newNode := NewNode(key, n)
	n.Children[key] = newNode
	return newNode.Insert(keys[1:], data)
}

//Get gets data from a path
func (n Node) Get(path []string) ([]byte, error) {
	if len(path) == 0 {
		return n.Value, nil
	}

	key := path[0]
	if existingNode, ok := n.Children[key]; ok {
		return existingNode.Get(path[1:])
	}

	return nil, errors.New("Not found")
}
