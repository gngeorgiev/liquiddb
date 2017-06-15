package store

import (
	"github.com/gngeorgiev/sstore/tree"
	"strings"
)

type Store struct {
	tree *tree.Tree
}

func New() *Store {
	return &Store{
		tree: tree.New(),
	}
}

func (store Store) Insert(data map[string]interface{}) error {
	return store.tree.Insert(data)
}

func (store Store) Get(path []string) ([]byte, error) {
	return store.tree.Get(path)
}

func (store Store) GetByString(path string) ([]byte, error) {
	return store.tree.Get(strings.Split(path, "."))
}