package liquiddb

type linker struct {
	links map[uint64]bool
}

func newLinker() *linker {
	return &linker{map[uint64]bool{}}
}

func (l *linker) save(id uint64) {
	l.links[id] = true
}

func (l linker) link(id uint64, data ...EventData) []EventData {
	res := make([]EventData, len(data))
	for i, d := range data {
		d.ID = id
		res[i] = d
	}

	return res
}
