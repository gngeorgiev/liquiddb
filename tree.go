package sstore

import (
	"fmt"

	"github.com/go-errors/errors"
)

const (
	TreeRoot string = "root"
)

var (
	NotFoundErr = errors.New("Not found")
)

type normalizedData struct {
	key   []string
	value []byte
}

type Tree struct {
	Root *Node
}

func newTree() *Tree {
	return &Tree{
		Root: newNode(TreeRoot, nil),
	}
}

func (tree Tree) normalize(data map[string]interface{}) ([]*normalizedData, error) {
	res := make([]*normalizedData, 0)

	var n func(interface{}, []string) error
	n = func(value interface{}, path []string) error {
		switch v := value.(type) {
		case []byte:
			res = append(res, &normalizedData{path, v})
		case map[string]interface{}:
			for k, v := range v {
				n(v, append(path, k))
			}
		default:
			return fmt.Errorf("Invalid datatype %s", v)
		}

		return nil
	}

	for k, v := range data {
		if err := n(v, []string{k}); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (tree Tree) findNode(path []string, autoCreate bool) *Node {
	node := tree.Root
	for {
		if len(path) == 0 {
			break
		}

		key := path[0]
		path = path[1:]
		if _, ok := node.Children[key]; !ok {
			if autoCreate {
				node.Children[key] = newNode(key, node)
			} else {
				return nil
			}
		}

		node = node.Children[key]
	}

	return node
}

func (tree Tree) performOnNodes(data []*normalizedData, f func(*normalizedData, *Node) OpInfo) []OpInfo {
	ops := make([]OpInfo, 0) //TODO: optimize

	for _, d := range data {
		node := tree.findNode(d.key[:], true)
		node.Value = d.value
		ops = append(ops, f(d, node))
		node.pristine = false
	}

	return ops
}

func (tree Tree) do(data map[string]interface{}, f func(*normalizedData, *Node) OpInfo) ([]OpInfo, error) {
	normalizedData, err := tree.normalize(data)
	if err != nil {
		return nil, err
	}

	return tree.performOnNodes(normalizedData, f), nil
}

func (tree Tree) Set(data map[string]interface{}) ([]OpInfo, error) {
	return tree.do(data, func(d *normalizedData, n *Node) OpInfo {
		var op Operation
		if n.pristine {
			op = OperationInsert
		} else {
			op = OperationUpdate
		}

		return OpInfo{
			Key:       n.Key,
			Operation: op,
			Path:      d.key,
			Value:     d.value,
		}
	})
}

func (tree Tree) Delete(path []string) (OpInfo, bool) {
	node := tree.findNode(path, false)
	if node == nil {
		return OpInfo{}, false
	}

	delete(node.Parent.Children, node.Key)
	val := node.Value
	node.Value = nil
	node.Parent = nil

	return OpInfo{
		Key:       node.Key,
		Operation: OperationDelete,
		Path:      path,
		Value:     val,
	}, true
}

func (tree Tree) Get(path []string) ([]byte, error) {
	node := tree.findNode(path, false)
	if node == nil {
		return nil, NotFoundErr
	}

	return node.Value, nil
}
