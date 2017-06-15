package tree

import "github.com/go-errors/errors"

type Node struct {
	Key string
	Value []byte
	Parent *Node
	Children map[string]*Node
}

func NewNode(key string, parent *Node) *Node {
	return &Node{
		key,
		nil,
		parent,
		map[string]*Node{},
	}
}

func (n *Node) buildPathUpwards() []string {
	path := []string{n.Key}
	currentNode := n
	for {
		if currentNode.Key == TREE_ROOT {
			break
		}

		p := currentNode.Parent.Key
		path = append(path, p)
		currentNode = currentNode.Parent
	}

	return path
}

func (n *Node) Insert(keys []string, data []byte) (opInfo, error) {
	if len(keys) == 0 {
		n.Value = data
		return opInfo{
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

func (n Node) Get(path[] string) ([]byte, error) {
	if len(path) == 0 {
		return n.Value, nil
	}

	key := path[0]
	if existingNode, ok := n.Children[key]; ok {
		return existingNode.Get(path[1:])
	}

	return nil, errors.New("Not found")
}