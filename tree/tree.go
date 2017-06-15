package tree

import (
	"github.com/go-errors/errors"
	"fmt"
)

const (
	TREE_ROOT string = "root"
)

type Tree struct {
	Root *Node
}

func New() *Tree {
	return &Tree{
		Root: NewNode(TREE_ROOT, nil),
	}
}

func (tree Tree) insertPath(path []string, data map[string]interface{}) ([]opInfo, error) {
	res := make([]opInfo, 0) //the initial size can be optimized

	for k, v := range data {
		switch t := v.(type) {
		case []byte:
			info, err := tree.Root.Insert(append(path, k), t)
			res = append(res, info)
			if err != nil {
				return res, err
			}
		case map[string]interface{}:
			infos, err := tree.insertPath(append(path, k), t)
			res = append(res, infos...)
			if err != nil {
				return res, err
			}
		default:
			return res, errors.New(fmt.Sprintf("Invalid datatype %s", t))
		}
	}

	return res, nil
}

func (tree Tree) Insert(data map[string]interface{}) ([]opInfo, error) {
	return tree.insertPath([]string{}, data)
}

func (tree Tree) Get(path []string) ([]byte, error) {
	return tree.Root.Get(path)
}