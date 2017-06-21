package liquiddb

import (
	"testing"
)

func Test_linker_save(t *testing.T) {
	l := newLinker()
	l.save(4)

	if l.links[4] == false {
		t.Fatal("Linker save not working")
	}
}

func Test_linker_link(t *testing.T) {
	l := newLinker()
	l.save(4)

	d := l.link(4, EventData{})

	if d[0].ID != 4 {
		t.Fatal("Linker link not working")
	}
}
