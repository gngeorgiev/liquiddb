package liquiddb

import (
	"strings"
)

//LiquidDb provides the means to store data and be notified of changes
type LiquidDb struct {
	//this link will identify a struct's current operation id
	//since non-pointers values are copied and pointer values' pointers
	//are copied, we can safely return a non-pointer value to the database without copying
	//the whole tree
	linkID uint64

	tree   *tree
	linker *linker
	*notifier
}

//New creates new database instance
func New() *LiquidDb {
	return &LiquidDb{
		tree:     newTree(),
		linker:   newLinker(),
		notifier: newNotifier(),
	}
}

//Link links the EventData from the next call to the specified id
func (db LiquidDb) Link(id uint64) LiquidDb {
	db.linkID = id
	db.linker.save(id)
	return db
}

//Set inserts a json in the database
func (db LiquidDb) Set(data map[string]interface{}) ([]EventData, error) {
	op, err := db.tree.Set(data)
	if err != nil {
		return nil, err
	}

	evData := db.linker.link(db.linkID, op...)
	db.notifier.notifyInternal(evData...)
	return op, nil
}

//SetPath sets value by a path, the data can be another json for nested insertion
func (db LiquidDb) SetPath(path []string, data interface{}) ([]EventData, error) {
	//TODO: test
	op, err := db.tree.SetPath(path, data)
	if err != nil {
		return nil, err
	}

	evData := db.linker.link(db.linkID, op...)
	db.notifier.notifyInternal(evData...)
	return evData, nil
}

//Get gets a value out of the store by a path formed by an array of strings
func (db LiquidDb) Get(path []string) (EventData, error) {
	op, err := db.tree.Get(path)
	evData := db.linker.link(db.linkID, op)
	//TODO: Do we want to notify on every get?
	db.notifier.notifyInternal(evData...)
	//TODO: if we have a NotFound, is it okay to ignore it and return nil value
	//i think we don't need this error at all
	return evData[0], err
}

//GetByString gets a value out of the store by a path formed by a string with dots
func (db LiquidDb) GetByString(path string) (interface{}, error) {
	return db.tree.Get(strings.Split(path, "."))
}

//Delete deletes a value from the store by a path
func (db LiquidDb) Delete(path []string) ([]EventData, bool) {
	//TODO: should this return error too, just like Get, or should get not return error?
	//the api must be consistent
	op, ok := db.tree.Delete(path)
	if !ok {
		return nil, false
	}

	evData := db.linker.link(db.linkID, op...)
	db.notifier.notifyInternal(evData...)
	return op, true
}

//DeleteByString deletes a value from the store by a path formed as string separated by dots
func (db LiquidDb) DeleteByString(path string) ([]EventData, bool) {
	return db.Delete(strings.Split(path, "."))
}
