package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gngeorgiev/liquiddb"
	"github.com/gorilla/websocket"
)

//TODO: Use protocol buffers!
func (a App) handleStoreNotify(conn *clientConnection, wg *sync.WaitGroup) {
	defer wg.Done()

	ch := make(chan liquiddb.EventData)
	a.db.Notify(ch, liquiddb.EventOperationDelete, liquiddb.EventOperationInsert,
		liquiddb.EventOperationUpdate, liquiddb.EventOperationGet)
	for {
		//TODO: data must be ordered, is this the case now?
		op := <-ch
		err := conn.WriteJSON(op)
		if err != nil {
			log.Println("write: ", err)
			break
		}
	}
}

type clientOperation string

const (
	clientOperationSet    = clientOperation("set")
	clientOperationDelete = clientOperation("delete")
	clientOperationGet    = clientOperation("get")
)

type clientData struct {
	ID        uint64          `json:"id,omitempty"`
	Operation clientOperation `json:"operation,omitempty"`
	Path      []string        `json:"path,omitempty"`
	Value     interface{}     `json:"value,omitempty"`
}

func (a App) handleClient(conn *clientConnection, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		var data clientData
		err := conn.ReadJSON(&data)
		if err != nil {
			//TODO: try to write one last error to the ws connection before closing it
			log.Println("read: ", err)
			break
		}

		switch data.Operation {
		case clientOperationSet:
			a.db.Link(data.ID).SetPath(data.Path, data.Value)
		case clientOperationDelete:
			a.db.Link(data.ID).Delete(data.Path)
		case clientOperationGet:
			a.db.Link(data.ID).Get(data.Path)
		default:
			//TODO: should we and how to notify the user about this
			log.Println("read: ", fmt.Errorf("Invalid operation type: %s", data.Operation))
		}

	}
}

type clientConnection struct {
	mu sync.Mutex

	ws *websocket.Conn
}

func newClientConnection(ws *websocket.Conn) *clientConnection {
	return &clientConnection{
		sync.Mutex{},
		ws,
	}
}

func (c *clientConnection) WriteJSON(o interface{}) error {
	//TODO: this mutex is not really needed since only one channel is coordinating
	//the writes to the connection at the moment, maybe remove it?
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.ws.WriteJSON(o)
}

func (c *clientConnection) ReadJSON(o interface{}) error {
	return c.ws.ReadJSON(o)
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

		conn := newClientConnection(ws)

		defer conn.ws.Close()

		var wg sync.WaitGroup
		wg.Add(2)

		go a.handleStoreNotify(conn, &wg)
		go a.handleClient(conn, &wg)

		wg.Wait()
	}

	http.HandleFunc("/db", handler)
	return http.ListenAndServe(":8080", nil)
}
