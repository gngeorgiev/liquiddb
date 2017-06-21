package liquiddb

import (
	"sync"

	"github.com/go-errors/errors"
	"github.com/thoas/go-funk"
)

type Notifier interface {
	Notify(c chan<- OpInfo, op ...Operation)
}

type notifier struct {
	sync.Mutex
	handlers    map[Operation][]chan<- OpInfo
	handlersMap map[chan<- OpInfo][]Operation
}

func newNotifier() *notifier {
	return &notifier{
		handlers:    make(map[Operation][]chan<- OpInfo),
		handlersMap: make(map[chan<- OpInfo][]Operation),
	}
}

func (n *notifier) Notify(c chan<- OpInfo, operations ...Operation) error {
	if c == nil {
		return errors.New("Invalid channel - nil")
	}

	n.Lock()
	defer n.Unlock()

	for _, op := range operations {
		if _, ok := n.handlers[op]; !ok {
			n.handlers[op] = make([]chan<- OpInfo, 0)
		}

		if _, ok := n.handlersMap[c]; !ok {
			n.handlersMap[c] = make([]Operation, 0)
		}

		n.handlers[op] = append(n.handlers[op], c)
		n.handlersMap[c] = append(n.handlersMap[c], op)
	}

	return nil
}

func (n *notifier) StopNotify(c chan<- OpInfo) error {
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

func (n *notifier) notifyInternal(notitifcations ...OpInfo) {
	n.Lock()
	defer n.Unlock()

	for _, notf := range notitifcations {
		for _, c := range n.handlers[notf.Operation] {
			//send to the channel without blocking
			select {
			case c <- notf:
			default:
			}
		}
	}

}
