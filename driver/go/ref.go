package liquidgo

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gngeorgiev/liquiddb/cmd/liquiddb/operations"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type Ref struct {
	liquid *LiquidGo

	path string

	valuesChMutex sync.Mutex
	valuesCh      []chan interface{}
}

func newRef(l *LiquidGo, path string) *Ref {
	return &Ref{
		liquid:        l,
		path:          path,
		valuesChMutex: sync.Mutex{},
		valuesCh:      make([]chan interface{}, 0),
	}
}

func (r *Ref) Value() (<-chan interface{}, <-chan error) {
	ch := make(chan interface{})
	errCh := make(chan error)

	go func() {
		clientData := ClientData{
			ID:        rand.Uint64(),
			Operation: operations.ClientOperationGet,
			Path:      strings.Split(r.path, "."),
			Timestamp: time.Now().Format(time.RFC3339),
			Value:     nil,
		}

		if err := r.liquid.Write(clientData); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}

		ev := r.liquid.ReadId(clientData.ID)
		ch <- ev.Value
	}()

	return ch, errCh
}

//TODO:
// func (r *Ref) ValueInt() (<-chan int, <-chan error) {

// }
