package main

import (
	"net/http"

	"log"

	"sync"

	"fmt"

	"github.com/gngeorgiev/liquiddb"
	"github.com/gorilla/websocket"
)

//App is the app that will be ran when the CLI is started
type App struct {
	store *liquiddb.Store
}

func main() {
	app := App{
		store: liquiddb.New(),
	}

	log.Fatal(app.startServer())
}

func (a App) handleStoreNotify(ws *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	ch := make(chan liquiddb.OpInfo)
	a.store.Notify(ch, liquiddb.OperationDelete, liquiddb.OperationInsert, liquiddb.OperationUpdate)
	for {
		//TODO: add ID's of messages, this way clients can reason for send and received messages
		//we should return the same id which belonged to the request
		//TODO: data must be ordered, is this the case now?
		op := <-ch
		err := ws.WriteJSON(op)
		if err != nil {
			log.Println("write: ", err)
			break
		}
	}
}

type clientOpType string

const (
	opTypeSet    = clientOpType("set")
	opTypeDelete = clientOpType("delete")
	opTypeGet    = clientOpType("get")
)

type clientOp struct {
	OperationType clientOpType `json:"operation,omitempty"`
	Path          []string     `json:"path,omitempty"`
	//TODO: should this value be json, or we can alter the store api instead
	Value interface{} `json:"value,omitempty"`
}

func (a App) handleClient(ws *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		var data clientOp
		err := ws.ReadJSON(&data)
		if err != nil {
			log.Println("read: ", err)
			break
		}

		switch data.OperationType {
		case opTypeSet:
			if _, err := a.store.SetPath(data.Path, data.Value); err != nil {
				log.Println("op set: ", err)
				break
			}
		case opTypeDelete:
			a.store.Delete(data.Path)
		case opTypeGet:
			value, err := a.store.Get(data.Path)
			d := liquiddb.OpInfo{
				Key:       data.Path[len(data.Path)-1],
				Operation: liquiddb.OperationGet,
				Path:      data.Path,
				Value:     value,
			}
			//TODO: notify user properly of error
			if err != nil && err != liquiddb.NotFoundErr {
				log.Println("op get: ", err)
				break
			}

			if err = ws.WriteJSON(d); err != nil {
				log.Println("op get: ", err)
				break
			}
		default:
			//TODO: should we and how to notify the user about this
			log.Println("read: ", fmt.Errorf("Invalid operation type: %s", data.OperationType))
		}

	}
}

func (a App) startServer() error {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		defer ws.Close()

		var wg sync.WaitGroup
		wg.Add(2)

		go a.handleStoreNotify(ws, &wg)
		go a.handleClient(ws, &wg)

		wg.Wait()
	}

	http.HandleFunc("/store", handler)
	return http.ListenAndServe(":8080", nil)
}
