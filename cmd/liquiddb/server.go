package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gngeorgiev/liquiddb"
	"github.com/gorilla/websocket"
	funk "github.com/thoas/go-funk"
)

//TODO: Use protocol buffers!
func (a App) handleStoreNotify(conn *clientConnection, stop chan struct{}) {
	ch := make(chan liquiddb.EventData, 10)
	a.db.Notify(ch, liquiddb.EventOperationDelete, liquiddb.EventOperationInsert,
		liquiddb.EventOperationUpdate, liquiddb.EventOperationGet)
	for {
		//TODO: data must be ordered, is this the case now?
		select {
		case op := <-ch:
			//TODO: more strings.Join to optimize....
			//I should probably just keep path in both forms - string and slice
			send, err := conn.WriteInterested(strings.Join(op.Path, "."), op.Operation, op)
			if send {
				log.Printf("Sending data: %+v", op)
			} else {
				log.Printf("Did not send data because not interested: %+v", op)
			}

			if err != nil {
				log.Println("write: ", err)
				close(stop)
				break
			}
		case <-stop:
			a.db.StopNotify(ch)
			return
		}
	}
}

type clientOperation string

const (
	clientOperationSet         = clientOperation("set")
	clientOperationDelete      = clientOperation("delete")
	clientOperationGet         = clientOperation("get")
	clientOperationSubscribe   = clientOperation("subscribe")
	clientOperationUnSubscribe = clientOperation("unsubscribe")
)

type clientData struct {
	ID        uint64          `json:"id,omitempty"`
	Operation clientOperation `json:"operation,omitempty"`
	Path      []string        `json:"path,omitempty"`
	Value     interface{}     `json:"value,omitempty"`
}

func (a App) handleClient(conn *clientConnection, stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
			var data clientData
			err := conn.ReadJSON(&data)
			if err != nil {
				//TODO: try to write one last error to the ws connection before closing it
				log.Println("read: ", err)
				close(stop)
				break
			}

			log.Printf("Received data: %+v", data)

			switch data.Operation {
			case clientOperationSet:
				a.db.Link(data.ID).SetPath(data.Path, data.Value)
			case clientOperationDelete:
				a.db.Link(data.ID).Delete(data.Path)
			case clientOperationGet:
				a.db.Link(data.ID).Get(data.Path)
			case clientOperationSubscribe:
				op := liquiddb.EventOperation(data.Value.(string))
				//TODO: can we optimize this strings join?
				conn.AddInterest(strings.Join(data.Path, "."), op)
			case clientOperationUnSubscribe:
				op := liquiddb.EventOperation(data.Value.(string))
				//TODO: can we optimize this strings join?
				conn.RemoveInterest(strings.Join(data.Path, "."), op)
			default:
				//TODO: should we and how to notify the user about this
				log.Println("read: ", fmt.Errorf("Invalid operation type: %s", data.Operation))
			}
		}
	}
}

type clientConnections struct {
	connectionAdded   chan *clientConnection
	connectionRemoved chan *clientConnection
	connections       []*clientConnection
}

var conns = clientConnections{
	make(chan *clientConnection),
	make(chan *clientConnection),
	make([]*clientConnection, 0),
}

type clientConnection struct {
	mu        sync.Mutex
	interests map[string][]liquiddb.EventOperation

	ws *websocket.Conn
}

func (c *clientConnection) close() error {
	index := funk.IndexOf(conns.connections, c)
	if index != -1 {
		conns.connections = append(conns.connections[:index], conns.connections[index+1:]...)
	}

	conns.connectionRemoved <- c

	return c.ws.Close()
}

func newClientConnection(ws *websocket.Conn) *clientConnection {
	c := &clientConnection{
		sync.Mutex{},
		map[string][]liquiddb.EventOperation{},
		ws,
	}

	log.Printf("New Connection: %s", ws.RemoteAddr().String())

	conns.connectionAdded <- c
	conns.connections = append(conns.connections, c)

	return c
}

func (c *clientConnection) WriteInterested(path string, op liquiddb.EventOperation, o liquiddb.EventData) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	//TODO: operations including root should be optimized and cleaned up
	interests := c.interests[path]
	if interests == nil {
		interests = c.interests[liquiddb.TreeRoot]
		if interests == nil {
			return false, nil
		}
	}

	for _, ev := range interests {
		if ev == op {
			return true, c.ws.WriteJSON(o)
		}
	}

	return false, nil
}

func (c *clientConnection) AddInterest(interest string, op liquiddb.EventOperation) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if interest == "" {
		interest = liquiddb.TreeRoot
	}

	interests := c.interests[interest]
	if interests == nil {
		c.interests[interest] = []liquiddb.EventOperation{op}
	} else {
		c.interests[interest] = append(interests, op)
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

func (c *clientConnection) RemoveInterest(interest string, op liquiddb.EventOperation) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if interest == "" {
		interest = liquiddb.TreeRoot
	}

	interests := c.interests[interest]
	if interests == nil || len(interests) == 0 {
		log.Printf("Trying to remove unexisting interest: %s, %s", interest, op)
		return
	}

	index := funk.IndexOf(interests, op)
	if index != -1 {
		c.interests[interest] = append(interests[:index], interests[index+1:]...)
	}
}

func (a App) dbHandler(upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		conn := newClientConnection(ws)

		defer conn.close()

		close := make(chan struct{})

		go a.handleStoreNotify(conn, close)
		go a.handleClient(conn, close)

		<-close
	}
}

func (a App) statsHandler(upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	connectionsCount := make(chan int)

	go func() {
		for {
			select {
			case <-conns.connectionAdded:
				select {
				case connectionsCount <- len(conns.connections):
				default:
				}
			case <-conns.connectionRemoved:
				select {
				case connectionsCount <- len(conns.connections):
				default:
				}
			}
		}
	}()

	type stats struct {
		connectionsCount int
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for {
			select {
			case c := <-connectionsCount:
				if err := ws.WriteJSON(stats{c}); err != nil {
					return
				}
			}
		}

	}
}

func (a App) startServer() error {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	http.HandleFunc("/db", a.dbHandler(upgrader))
	http.HandleFunc("/stats", a.statsHandler(upgrader))

	return http.ListenAndServe(":8080", nil)
}
