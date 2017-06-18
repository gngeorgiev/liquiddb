package store

import (
	"reflect"
	"sync"
	"testing"
)

var (
	data = map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": []byte("foobar"),
		},
	}
)

func TestNotify(t *testing.T) {
	store := New()

	c := make(chan Notification, 1)
	store.Notify(c, OperationInsert)

	var wg sync.WaitGroup
	wg.Add(1)
	value := []byte("foobar")
	go func() {
		defer wg.Done()

		notf := <-c

		if notf.Operation() != OperationInsert {
			t.Errorf("Invalid operation, %s", notf.Operation())
		}

		if notf.Key() != "bar" {
			t.Errorf("Invalid key, %s", notf.Key())
		}

		if !reflect.DeepEqual(notf.Value(), value) {
			t.Errorf("Invalid value, %s", notf.Value())
		}

		if !reflect.DeepEqual(notf.Path(), []string{"foo", "bar"}) {
			t.Errorf("Invalid path, %s", notf.Path())
		}
	}()

	store.Insert(data)
	wg.Wait()
}
