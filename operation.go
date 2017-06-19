package sstore

type Operation string

const (
	OperationInsert = Operation("insert")
	OperationDelete = Operation("delete")
	OperationUpdate = Operation("update")
)

type OpInfo struct {
	Operation Operation
	Path      []string
	Key       string
	Value     []byte
}
