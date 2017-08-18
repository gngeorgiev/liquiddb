package operations

import (
	"time"

	"github.com/gngeorgiev/liquiddb"
)

type ClientOperation string

const (
	ClientOperationSet          = ClientOperation("set")
	ClientOperationDelete       = ClientOperation("delete")
	ClientOperationGet          = ClientOperation("get")
	ClientOperationSubscribe    = ClientOperation("subscribe")
	ClientOperationUnSubscribe  = ClientOperation("unsubscribe")
	HearthbeatOperation         = "hearthbeat"
	HearthbeatResponseOperation = "hearthbeatResponse"
)

type OperationClientData struct {
	ID        uint64          `json:"id,omitempty"`
	Operation ClientOperation `json:"operation,omitempty"`
	Path      []string        `json:"path,omitempty"`
	Value     interface{}     `json:"value,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
}

type ClientInterest struct {
	Id        uint64
	Operation liquiddb.EventOperation
	Timestamp time.Time
}
