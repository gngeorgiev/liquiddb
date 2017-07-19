package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"time"

	"github.com/gngeorgiev/liquiddb"
	"github.com/gorilla/websocket"
	funk "github.com/thoas/go-funk"
)

type clientOperation string

const (
	clientOperationSet         = clientOperation("set")
	clientOperationDelete      = clientOperation("delete")
	clientOperationGet         = clientOperation("get")
	clientOperationSubscribe   = clientOperation("subscribe")
	clientOperationUnSubscribe = clientOperation("unsubscribe")
	hearthbeatOperation        = "hearthbeat"
)

type operationClientData struct {
	ID        uint64          `json:"id,omitempty"`
	Operation clientOperation `json:"operation,omitempty"`
	Path      []string        `json:"path,omitempty"`
	Value     interface{}     `json:"value,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
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

type clientInterest struct {
	id        uint64
	operation liquiddb.EventOperation
	timestamp time.Time
}

type clientConnection struct {
	mu        sync.Mutex
	interests map[string][]*clientInterest

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
		map[string][]*clientInterest{},
		ws,
	}

	log.Printf("New Connection: %s", ws.RemoteAddr().String())

	conns.connectionAdded <- c
	conns.connections = append(conns.connections, c)

	return c
}

func (c *clientConnection) WriteInterested(path string, o liquiddb.EventData) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	op := o.Operation

	//TODO: operations including root should be optimized and cleaned up
	interests := c.interests[path]
	if interests == nil {
		interests = c.interests[liquiddb.TreeRoot]
		if interests == nil {
			return false, nil
		}
	}

	for _, interest := range interests {
		log.Println(interest.timestamp)
		log.Println(o.Timestamp)
		if interest.operation == op && o.Timestamp.After(interest.timestamp) {
			return true, c.ws.WriteJSON(o)
		} else if !o.Timestamp.After(interest.timestamp) {
			log.Println("Didnt send, no matcing time")
		}
	}

	return false, nil
}

func (c *clientConnection) AddInterest(interest string, op liquiddb.EventOperation, o operationClientData) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	timestamp := o.Timestamp

	if interest == "" {
		interest = liquiddb.TreeRoot
	}

	interests := c.interests[interest]
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return err
	}

	cInterest := &clientInterest{
		o.ID,
		op,
		t,
	}
	if interests == nil {
		c.interests[interest] = []*clientInterest{cInterest}
	} else {
		c.interests[interest] = append(interests, cInterest)
	}

	return nil
}

func (c *clientConnection) WriteJSON(o interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.ws.WriteJSON(o)
}

func (c *clientConnection) ReadJSON(o interface{}) error {
	return c.ws.ReadJSON(o)
}

func (c *clientConnection) RemoveInterest(interest string, op liquiddb.EventOperation, o operationClientData) {
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

	for i, cInterest := range interests {
		if cInterest.operation == op && cInterest.id == o.ID {
			c.interests[interest] = append(interests[:i], interests[i+1:]...)
		}
	}

	if len(c.interests[interest]) == 0 {
		delete(c.interests, interest)
	}
}

//TODO: Use protocol buffers!
func (a App) handleSocketStoreNotify(conn *clientConnection, terminate chan struct{}) error {
	ch := make(chan liquiddb.EventData, 10)
	a.db.Notify(ch, liquiddb.EventOperationDelete, liquiddb.EventOperationInsert,
		liquiddb.EventOperationUpdate, liquiddb.EventOperationGet)
	defer a.db.StopNotify(ch)

	for {
		//TODO: data must be ordered, is this the case now?
		select {
		case <-terminate:
			return nil
		case op := <-ch:
			//TODO: more strings.Join to optimize....
			//I should probably just keep path in both forms - string and slice
			send, err := conn.WriteInterested(strings.Join(op.Path, "."), op)
			if send {
				log.Printf("Sending data: %+v", op)
			} else {
				log.Printf("Did not send data because not interested: %+v", op)
			}

			if err != nil {
				log.Println("write: ", err)
				return err
			}
		}
	}
}

func (a App) handleSocketClient(conn *clientConnection, terminate chan struct{}) error {
	for {
		select {
		case <-terminate:
			return nil
		default:
			var data operationClientData
			err := conn.ReadJSON(&data)
			if err != nil {
				//TODO: try to write one last error to the ws connection before closing it
				log.Println("read: ", err)
				return err
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
				if err := conn.AddInterest(strings.Join(data.Path, "."), op, data); err != nil {
					log.Println("add interest: ", err)
					return err
				}
			case clientOperationUnSubscribe:
				op := liquiddb.EventOperation(data.Value.(string))
				//TODO: can we optimize this strings join?
				conn.RemoveInterest(strings.Join(data.Path, "."), op, data)
			default:
				//TODO: should we and how to notify the user about this
				log.Println("read: ", fmt.Errorf("Invalid operation type: %s", data.Operation))
			}

			log.Printf("Processed data: %+v", data)
		}
	}
}

func (a App) handleSocketHearthbeat(conn *clientConnection, terminate chan struct{}) error {
	ticker := time.NewTicker(10 * time.Second)

	sendHearthbeat := func() error {
		err := conn.WriteJSON(struct {
			Operation string `json:"operation,omitempty"`
			Timestamp string `json:"timestamp,omitempty"`
		}{
			Operation: hearthbeatOperation,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
		if err != nil {
			return err
		}

		return nil
	}

	//send hearthbeat immediately after connected
	if err := sendHearthbeat(); err != nil {
		return err
	}

	for {
		select {
		case <-terminate:
			return nil
		case <-ticker.C:
			if err := sendHearthbeat(); err != nil {
				ticker.Stop()
				log.Printf("hearthbeat: %s", err)
				return err
			}
		}
	}
}

func (a App) handleSocketClose(conn *clientConnection, terminate chan struct{}) error {
	ch := make(chan error)

	conn.ws.SetCloseHandler(func(code int, text string) error {
		ch <- fmt.Errorf("Socket close %d %s", code, text)
		return nil
	})

	return <-ch
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

		handlers := []func(*clientConnection, chan struct{}) error{
			a.handleSocketStoreNotify,
			a.handleSocketClient,
			a.handleSocketHearthbeat,
			a.handleSocketClose,
		}

		terminateHandler := make(chan struct{})
		closeConnection := make(chan error)

		var wg sync.WaitGroup
		wg.Add(len(handlers))
		allHandlersTerminated := make(chan struct{})
		go func() {
			wg.Wait()
			allHandlersTerminated <- struct{}{}
		}()

		for _, handler := range handlers {
			go func(handler func(*clientConnection, chan struct{}) error) {
				defer wg.Done()
				err := handler(conn, terminateHandler)
				if err != nil {
					closeConnection <- err
				}
			}(handler)
		}

		go func() {
			terminationSend := false
			for err := range closeConnection {
				if !terminationSend {
					terminationSend = true
					//terminate all handlers
					close(terminateHandler)
				}

				if err != nil {
					//log all errors
					log.Println(err)
				} else {
					return
				}
			}
		}()

		<-allHandlersTerminated
		closeConnection <- nil
		log.Printf("Closed connection: %s", ws.RemoteAddr().String())
	}
}

type stats struct {
	connectionsCount int
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

const serverPort = ":8082"

func (a App) startServer() error {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	http.HandleFunc("/db", a.dbHandler(upgrader))
	http.HandleFunc("/stats", a.statsHandler(upgrader))

	log.Printf("Listening on port %s", serverPort)
	return http.ListenAndServe(serverPort, nil)
}
