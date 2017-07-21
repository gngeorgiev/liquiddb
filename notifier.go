package liquiddb

import (
	"sync"

	"time"

	"github.com/go-errors/errors"
	"github.com/thoas/go-funk"
)

//EventOperation is a db operation
type EventOperation string

const (
	//EventOperationInsert is an insert db operation
	EventOperationInsert = EventOperation("insert")
	//EventOperationDelete is a delete db operation
	EventOperationDelete = EventOperation("delete")
	//EventOperationUpdate is an update db operation
	EventOperationUpdate = EventOperation("update")
	//EventOperationGet is a get db operation
	EventOperationGet = EventOperation("get")
)

//EventData is a whole db event holding data and metadata
type EventData struct {
	ID        uint64         `json:"id,omitempty"`
	Operation EventOperation `json:"operation,omitempty"`
	Path      []string       `json:"path,omitempty"`
	Key       string         `json:"key,omitempty"`
	Value     interface{}    `json:"value,omitempty"`
	Timestamp time.Time
}

type EventsSortedByTimestamp []EventData

func (e EventsSortedByTimestamp) Len() int {
	return len(e)
}

func (e EventsSortedByTimestamp) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e EventsSortedByTimestamp) Less(i, j int) bool {
	return e[i].Timestamp.Before(e[j].Timestamp)
}

type notifier struct {
	sync.Mutex
	handlers    map[EventOperation][]chan<- EventData
	handlersMap map[chan<- EventData][]EventOperation
}

func newNotifier() *notifier {
	return &notifier{
		handlers:    make(map[EventOperation][]chan<- EventData),
		handlersMap: make(map[chan<- EventData][]EventOperation),
	}
}

func (n *notifier) Notify(c chan<- EventData, operations ...EventOperation) error {
	if c == nil {
		return errors.New("Invalid channel - nil")
	}

	n.Lock()
	defer n.Unlock()

	for _, op := range operations {
		if _, ok := n.handlers[op]; !ok {
			n.handlers[op] = make([]chan<- EventData, 0)
		}

		if _, ok := n.handlersMap[c]; !ok {
			n.handlersMap[c] = make([]EventOperation, 0)
		}

		n.handlers[op] = append(n.handlers[op], c)
		n.handlersMap[c] = append(n.handlersMap[c], op)
	}

	return nil
}

func (n *notifier) StopNotify(c chan<- EventData) error {
	if c == nil {
		return errors.New("Invalid channel - nil")
	}

	n.Lock()
	defer n.Unlock()

	operations := n.handlersMap[c]
	delete(n.handlersMap, c)
	for _, op := range operations {
		channels := n.handlers[op]
		//TODO: check performance since this uses reflection,
		//shouldn't be a problem tho, there won't be many elements
		index := funk.IndexOf(channels, c)
		n.handlers[op] = append(channels[:index], channels[index+1:]...)
	}

	return nil
}

func (n *notifier) notifyInternal(notifications ...EventData) {
	n.Lock()
	defer n.Unlock()

	for _, notification := range notifications {
		notification.Timestamp = time.Now().UTC()
		for _, c := range n.handlers[notification.Operation] {
			c <- notification
		}
	}

}
