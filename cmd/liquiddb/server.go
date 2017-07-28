package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"time"

	"github.com/gngeorgiev/liquiddb"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	funk "github.com/thoas/go-funk"
)

type clientOperation string

const (
	clientOperationSet          = clientOperation("set")
	clientOperationDelete       = clientOperation("delete")
	clientOperationGet          = clientOperation("get")
	clientOperationSubscribe    = clientOperation("subscribe")
	clientOperationUnSubscribe  = clientOperation("unsubscribe")
	hearthbeatOperation         = "hearthbeat"
	hearthbeatResponseOperation = "hearthbeatResponse"
)

type operationClientData struct {
	ID        uint64          `json:"id,omitempty"`
	Operation clientOperation `json:"operation,omitempty"`
	Path      []string        `json:"path,omitempty"`
	Value     interface{}     `json:"value,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
}

type clientConnections struct {
	sync.Mutex

	connectionAdded   chan *clientConnection
	connectionRemoved chan *clientConnection
	connections       []*clientConnection
}

var conns = clientConnections{
	sync.Mutex{},

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

	latencyMutex       sync.Mutex
	latencyHistory     [3]int32
	latency            int32
	hearthbeatResponse chan struct{}

	ws *websocket.Conn
}

func (c *clientConnection) close() error {
	index := funk.IndexOf(conns.connections, c)
	if index != -1 {
		conns.Lock()
		conns.connections = append(conns.connections[:index], conns.connections[index+1:]...)
		conns.Unlock()
	}

	conns.connectionRemoved <- c

	return c.ws.Close()
}

func newClientConnection(ws *websocket.Conn) *clientConnection {
	c := &clientConnection{
		mu:        sync.Mutex{},
		interests: map[string][]*clientInterest{},

		latencyMutex:       sync.Mutex{},
		latencyHistory:     [3]int32{},
		latency:            0,
		hearthbeatResponse: make(chan struct{}),

		ws: ws,
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
		interestHasValidTimestamp := o.Timestamp.After(interest.timestamp) || o.Timestamp.Equal(interest.timestamp)
		if interest.operation == op && interestHasValidTimestamp {
			return true, c.ws.WriteJSON(o)
		}
	}

	log.Println("Didn't send, no matching time")
	log.Printf("Operation timestamp %s", o.Timestamp)
	for _, interest := range interests {
		log.Printf("Interest %d - %s - %s", interest.id, interest.operation, interest.timestamp)
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

	//substract the latency from the interest, theoretically if an event happens
	//at a certain time and the client has a X latency, he might not receive the event because
	//the interest will be send later due to the latency
	t = t.Add(time.Duration(-c.latency) * time.Millisecond)

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

			if data.Operation != hearthbeatResponseOperation {
				log.Printf("Received data: %+v", data)
			}

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
			case hearthbeatResponseOperation:
				conn.hearthbeatResponse <- struct{}{}
			default:
				//TODO: should we and how to notify the user about this
				log.Println("read: ", fmt.Errorf("Invalid operation type: %s", data.Operation))
			}

			if data.Operation != hearthbeatResponseOperation {
				log.Printf("Processed data: %+v", data)
			}
		}
	}
}

func (a App) handleSocketHearthbeat(conn *clientConnection, terminate chan struct{}) error {
	//TODO: refactor this method a bit as it has become too large
	//also refactor the whole file as it has also become too large
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

	getSingleHearthbeatLatency := func() (int32, error) {
		sendTime := time.Now()
		latencyResult := make(chan int32)
		timeout := make(chan struct{})
		t := time.NewTimer(30 * time.Second)
		defer t.Stop()

		go func() {
			select {
			case <-conn.hearthbeatResponse:
			case <-t.C:
				timeout <- struct{}{}
				return
			}

			now := time.Now()

			difference := now.Sub(sendTime).Nanoseconds()
			latency := (difference / int64(time.Millisecond)) / 2
			latencyResult <- int32(latency)
		}()

		if err := sendHearthbeat(); err != nil {
			return 0, err
		}

		select {
		case latency := <-latencyResult:
			return latency, nil
		case <-timeout:
			return 0, errors.New("Hearthbeat timeout")
		}
	}

	//send 3 quick hearthbeats after a connection is established
	//to calculate latency asap
	for i := 0; i < 3; i++ {
		conn.latencyMutex.Lock()
		latency, err := getSingleHearthbeatLatency()
		if err != nil {
			log.Printf("Initial hearthbeat #%d: %s", i, err)
			return err
		}

		conn.latencyHistory[i] = latency
		conn.latencyMutex.Unlock()
	}

	calculateLatency := func() {
		conn.latency = (conn.latencyHistory[0] + conn.latencyHistory[1] + conn.latencyHistory[2]) / 3
	}

	conn.latencyMutex.Lock()
	calculateLatency()
	conn.latencyMutex.Unlock()

	const pingInterval = 500 * time.Millisecond

	timer := time.NewTimer(pingInterval)
	defer timer.Stop()

	for {
		select {
		case <-terminate:
			return nil
		case <-timer.C:
			latency, err := getSingleHearthbeatLatency()
			if err != nil {
				log.Printf("hearthbeat: %s", err)
				return err
			}

			conn.latencyMutex.Lock()
			//move the history of latencies to the left, leaving the last index to the new latency
			conn.latencyHistory[0] = conn.latencyHistory[1]
			conn.latencyHistory[1] = conn.latencyHistory[2]
			conn.latencyHistory[2] = latency
			calculateLatency()
			conn.latencyMutex.Unlock()

			timer.Stop()
			timer.Reset(pingInterval)
		}
	}
}

func (a App) handleSocketClose(conn *clientConnection, terminate chan struct{}) error {
	ch := make(chan error)

	go func() {
		conn.ws.SetCloseHandler(func(code int, text string) error {
			ch <- fmt.Errorf("Socket close %d %s", code, text)
			return nil
		})
	}()

	select {
	case <-terminate:
		return nil
	case err := <-ch:
		return err
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

		handlers := []func(*clientConnection, chan struct{}) error{
			a.handleSocketStoreNotify,
			a.handleSocketClient,
			a.handleSocketHearthbeat,
			a.handleSocketClose,
		}

		terminateHandlers := make([]chan struct{}, len(handlers))
		closeConnection := make(chan error)

		var wg sync.WaitGroup
		wg.Add(len(handlers))
		allHandlersTerminated := make(chan struct{})
		go func() {
			wg.Wait()
			allHandlersTerminated <- struct{}{}
		}()

		for i, handler := range handlers {
			terminateHandlers[i] = make(chan struct{}, 1)
			go func(handler func(*clientConnection, chan struct{}) error, terminateHandler chan struct{}) {
				defer wg.Done()
				err := handler(conn, terminateHandler)
				if err != nil {
					closeConnection <- err
				}
			}(handler, terminateHandlers[i])
		}

		go func() {
			terminationSend := false
			for err := range closeConnection {
				if !terminationSend {
					terminationSend = true
					//terminate all handlers
					for _, h := range terminateHandlers {
						h <- struct{}{}
					}
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
	ConnectionsCount int `json:"connectionsCount,omitempty"`
}

func (a App) statsHandler(upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	var connectionsCountsMutex sync.Mutex
	connectionsCounts := make([]chan int, 0)

	go func() {
		updateConnectionsCount := func() {
			l := len(conns.connections)
			connectionsCountsMutex.Lock()
			defer connectionsCountsMutex.Unlock()
			for _, ch := range connectionsCounts {
				ch <- l
			}
		}

		for {
			select {
			case <-conns.connectionAdded:
				updateConnectionsCount()
			case <-conns.connectionRemoved:
				updateConnectionsCount()
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		countCh := make(chan int)

		defer func() {
			close(countCh)

			connectionsCountsMutex.Lock()
			index := funk.IndexOf(connectionsCounts, countCh)
			connectionsCounts = append(connectionsCounts[:index], connectionsCounts[index+1:]...)
			connectionsCountsMutex.Unlock()
		}()

		connectionsCountsMutex.Lock()
		connectionsCounts = append(connectionsCounts, countCh)
		connectionsCountsMutex.Unlock()

		//send the stats on the initial connection
		if err := ws.WriteJSON(stats{len(conns.connections)}); err != nil {
			log.Println(err)
			return
		}

		for {
			select {
			case c := <-countCh:
				if err := ws.WriteJSON(stats{c}); err != nil {
					log.Println(err)
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

	mux := http.NewServeMux()

	mux.HandleFunc("/db", a.dbHandler(upgrader))
	mux.HandleFunc("/stats", a.statsHandler(upgrader))

	handler := cors.AllowAll().Handler(mux)

	log.Printf("Listening on port %s", serverPort)
	return http.ListenAndServe(serverPort, handler)
}
