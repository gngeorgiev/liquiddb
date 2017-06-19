package sstore

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	b    = []byte("foobar")
	p    = []string{"foo", "bar"}
	data = map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": b,
		},
	}
)

func TestNotify(t *testing.T) {
	ops := []Operation{OperationInsert, OperationDelete, OperationUpdate}

	for _, op := range ops {
		fn := func(op Operation) func(t *testing.T) {
			return func(t *testing.T) {
				store := New()

				c := make(chan OpInfo, 1)

				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer wg.Done()

					notf := <-c

					if notf.Operation != op {
						t.Errorf("Invalid operation, %s", notf.Operation)
					}

					if notf.Key != "bar" {
						t.Errorf("Invalid key, %s", notf.Key)
					}

					if !reflect.DeepEqual(notf.Value, b) {
						t.Errorf("Invalid value, %s", notf.Value)
					}

					if !reflect.DeepEqual(notf.Path, []string{"foo", "bar"}) {
						t.Errorf("Invalid path, %s", notf.Path)
					}
				}()

				switch op {
				case OperationInsert:
					store.Notify(c, op)
					store.Set(data)
				case OperationUpdate:
					store.Set(data)
					store.Notify(c, op)
					store.Set(data)
				case OperationDelete:
					store.Set(data)
					store.Notify(c, op)
					store.Delete(p)
				}

				wg.Wait()
			}
		}(op)
		t.Run(fmt.Sprintf("TestNotify-%s", op), fn)
	}
}

func TestStopNotify(t *testing.T) {
	store := New()

	var wg sync.WaitGroup

	ch := make(chan OpInfo, 1)
	store.Notify(ch, OperationInsert, OperationUpdate)

	wg.Add(1)
	go func() {
		defer wg.Done()

		notf := <-ch
		if notf.Operation != OperationInsert {
			t.Fatalf("Invalid operation %s", notf.Operation)
		}
	}()
	store.Set(data)

	wg.Wait()
	wg.Add(1)
	go func() {
		defer wg.Done()

		timer := time.After(time.Millisecond * 500)

		select {
		case notf := <-ch:
			t.Fatalf("Should not receive notification %+s", notf)
		case <-timer:
		}
	}()

	store.StopNotify(ch)
	store.Set(data)

	wg.Wait()
}
