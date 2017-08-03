package liquiddb

import (
	"github.com/sasha-s/go-deadlock"
)

type linker struct {
	mu deadlock.Mutex

	links map[uint64]bool
}

func newLinker() *linker {
	return &linker{
		mu:    deadlock.Mutex{},
		links: map[uint64]bool{},
	}
}

//TODO: the saved links must be
func (l *linker) save(id uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.links[id] = true
}

func (l *linker) link(id uint64, data ...EventData) []EventData {
	res := make([]EventData, len(data))
	for i, d := range data {
		d.ID = id
		res[i] = d
	}

	return res
}
