package liquidgo

import (
	"testing"
)

func TestRef_Value(t *testing.T) {
	l := connect(t)

	r := l.Ref("foo.bar")

	valCh, _ := r.Value()
	val := <-valCh
	res := val.(int)

	if res != 5 {
		t.Fatalf("%d is different from 5", res)
	}
}
