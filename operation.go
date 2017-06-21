package liquiddb

type Operation string

const (
	OperationInsert = Operation("insert")
	OperationDelete = Operation("delete")
	OperationUpdate = Operation("update")
	OperationGet = Operation("get")
)

type OpInfo struct {
	Operation Operation   `json:"operation,omitempty"`
	Path      []string    `json:"path,omitempty"`
	Key       string      `json:"key,omitempty"`
	Value     interface{} `json:"value,omitempty"`
}
