package store

import (
	"github.com/gngeorgiev/sstore/tree"
)

type Operation string

const (
	OperationInsert = Operation("insert")
	OperationDelete = Operation("delete")
	OperationUpdate = Operation("update")
)

type Notification interface {
	Operation() Operation
	Path() []string
	Key() string
	Value() []byte
}

type notification struct {
	operation Operation
	path      []string
	key       string
	value     []byte
}

func (n notification) Operation() Operation {
	return n.operation
}

func (n notification) Path() []string {
	return n.path
}

func (n notification) Key() string {
	return n.key
}

func (n notification) Value() []byte {
	return n.value
}

func NewNotificationFromOp(info tree.OpInfo, op Operation) Notification {
	return &notification{
		operation: op,
		key:       info.Key,
		path:      info.Path,
		value:     info.Value,
	}
}
