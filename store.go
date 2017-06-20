package sstore

import (
	"strings"
)

//Store provides the means to store data and be notified of changes
type Store struct {
	tree *Tree
	*notifier
}

//New creates new Store
func New() *Store {
	return &Store{
		tree:     newTree(),
		notifier: newNotifier(),
	}
}

//Set inserts a json in the store
func (store Store) Set(data map[string]interface{}) ([]OpInfo, error) {
	op, err := store.tree.Set(data)
	if err != nil {
		return nil, err
	}

	store.notifyInternal(op...)
	return op, nil
}

//Get gets a value out of the store by a path formed by an array of strings
func (store Store) Get(path []string) (interface{}, error) {
	return store.tree.Get(path)
}

//GetByString gets a value out of the store by a path formed by a string with dots
func (store Store) GetByString(path string) (interface{}, error) {
	return store.tree.Get(strings.Split(path, "."))
}

//Delete deletes a value from the store by a path
func (store Store) Delete(path []string) (OpInfo, bool) {
	op, ok := store.tree.Delete(path)
	if !ok {
		return OpInfo{}, false
	}

	store.notifyInternal(op)
	return op, true
}

//DeleteByString deletes a value from the store by a path formed as string separated by dots
func (store Store) DeleteByString(path string) (OpInfo, bool) {
	return store.Delete(strings.Split(path, "."))
}
