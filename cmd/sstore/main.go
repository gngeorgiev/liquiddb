package main

import (
	"fmt"
	"github.com/gngeorgiev/sstore/store"
)

func main() {
	s := store.New()

	data := map[string]interface{}{
		"gosho": map[string]interface{}{
			"data": []byte("data"),
		},
	};

	t.Insert(data)
	d, _ := t.GetByString("gosho.data")
	fmt.Println(string(d))
}
