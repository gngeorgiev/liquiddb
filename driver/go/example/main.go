package main

import (
	"fmt"

	"github.com/gngeorgiev/liquiddb/driver/go"
)

func main() {
	l := liquidgo.NewDefault()

	if err := l.Connect(); err != nil {
		panic(err)
	}

	r := l.Root()
	dataCh, errCh := r.Value()

	select {
	case data := <-dataCh:
		fmt.Println(data)
	case err := <-errCh:
		fmt.Println(err)
	}
}
