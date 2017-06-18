package store

import (
	"strings"

	"github.com/gngeorgiev/sstore/tree"
)

//Store provides the means to store data and be notified of changes
type Store struct {
	tree *tree.Tree
	*notifier
}

//New creates new Store
func New() *Store {
	return &Store{
		tree:     tree.New(),
		notifier: newNotifier(),
	}
}

func (store Store) buildNotify(infos []tree.OpInfo, op Operation) {
	go func() {
		notifications := make([]Notification, len(infos))
		for i, info := range infos {
			notifications[i] = NewNotificationFromOp(info, op)
		}

		store.notifyInternal(notifications...)
	}()
}

//Insert inserts a json in the store
func (store Store) Insert(data map[string]interface{}) ([]tree.OpInfo, error) {
	op, err := store.tree.Insert(data)
	if err != nil {
		return nil, err
	}

	store.buildNotify(op, OperationInsert)
	return op, nil
}

//Get gets a value out of the store by a path formed by an array of strings
func (store Store) Get(path []string) ([]byte, error) {
	return store.tree.Get(path)
}

//GetByString gets a value out of the store by a path formed by a string with dots
func (store Store) GetByString(path string) ([]byte, error) {
	return store.tree.Get(strings.Split(path, "."))
}
