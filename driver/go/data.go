package liquidgo

import (
	"github.com/gngeorgiev/liquiddb/cmd/liquiddb/operations"
)

type ClientData struct {
	ID        uint64                     `json:"id,omitempty"`
	Timestamp string                     `json:"timestamp,omitempty"`
	Operation operations.ClientOperation `json:"operation,omitempty"`
	Path      []string                   `json:"path,omitempty"`
	Value     interface{}                `json:"value,omitempty"`
}
