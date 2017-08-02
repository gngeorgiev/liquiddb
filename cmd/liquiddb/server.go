package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

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
	sync.Mutex

	interests map[string][]*clientInterest

	latencyMutex       sync.Mutex
	latencyHistory     [3]int32
	latency            int32
	hearthbeatResponse chan struct{}

	ws *websocket.Conn
}

func (c *clientConnection) String() string {
	return c.ws.RemoteAddr().String()
}

func (c *clientConnection) close() error {
	c.Lock()
	defer c.Unlock()

	conns.Lock()
	defer conns.Unlock()

	index := -1
	for i, conn := range conns.connections {
		if conn == c {
			index = i
			break
		}
	}

	// log.WithFields(log.Fields{
	// 	"index": index,
	// 	"len":   len(conns.connections),
	// })
	if index != -1 {
		conns.connections = append(conns.connections[:index], conns.connections[index+1:]...)
		conns.connectionRemoved <- c
	}

	return c.ws.Close()
}

func newClientConnection(ws *websocket.Conn) *clientConnection {
	conns.Lock()
	defer conns.Unlock()

	c := &clientConnection{
		Mutex: sync.Mutex{},

		interests: map[string][]*clientInterest{},

		latencyMutex:       sync.Mutex{},
		latencyHistory:     [3]int32{},
		latency:            0,
		hearthbeatResponse: make(chan struct{}),

		ws: ws,
	}

	log.WithField("address", ws.RemoteAddr().String()).Info("New Connection")

	conns.connectionAdded <- c
	conns.connections = append(conns.connections, c)

	return c
}

func (c *clientConnection) WriteInterested(path string, o liquiddb.EventData) (bool, error) {
	c.Lock()
	defer c.Unlock()

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
		log.WithFields(log.Fields{
			"id":        interest.id,
			"operation": interest.operation,
			"timestamp": interest.timestamp,
		}).Debug("Interest")

		interestHasValidTimestamp := o.Timestamp.After(interest.timestamp) || o.Timestamp.Equal(interest.timestamp)
		if interest.operation == op && interestHasValidTimestamp {
			return true, c.ws.WriteJSON(o)
		}
	}

	return false, nil
}

func (c *clientConnection) AddInterest(interest string, op liquiddb.EventOperation, o operationClientData) error {
	c.Lock()
	defer c.Unlock()

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
	c.Lock()
	defer c.Unlock()

	return c.ws.WriteJSON(o)
}

func (c *clientConnection) ReadJSON(o interface{}) error {
	return c.ws.ReadJSON(o)
}

func (c *clientConnection) RemoveInterest(interest string, op liquiddb.EventOperation, o operationClientData) {
	c.Lock()
	defer c.Unlock()

	if interest == "" {
		interest = liquiddb.TreeRoot
	}

	interests := c.interests[interest]
	if interests == nil || len(interests) == 0 {
		log.WithFields(log.Fields{
			"interest":  interest,
			"operation": op,
		}).Warn("Trying to remove unexisting interest")
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
				log.WithField("data", op).Debug("Sending data")
			} else {
				log.WithField("operation", op).Debug("Did not send data because not interested")
			}

			if err != nil {
				log.WithField("category", "write").Error(err)
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
				log.WithField("category", "read").Error(err)
				return err
			}

			if data.Operation != hearthbeatResponseOperation {
				log.WithField("data", data).Debug("Received data")
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
					log.WithField("category", "add interest").Error(err)
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
				log.WithFields(log.Fields{
					"category":  "read",
					"operation": data.Operation,
				}).Error("Invalid operation type")
			}

			if data.Operation != hearthbeatResponseOperation {
				log.WithField("data", data).Debug("Processed data")
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

	getSingleHearthbeatLatency := func() (<-chan int32, <-chan error) {
		sendTime := time.Now()

		latency := make(chan int32)
		err := make(chan error)

		go func() {
			t := time.NewTimer(10 * time.Second)
			defer t.Stop()

			if hearthBeatError := sendHearthbeat(); hearthBeatError != nil {
				err <- hearthBeatError
				return
			}

			select {
			case <-conn.hearthbeatResponse:
				now := time.Now()
				difference := now.Sub(sendTime).Nanoseconds()
				latencyResult := (difference / int64(time.Millisecond)) / 2
				latency <- int32(latencyResult)
			case <-t.C:
				err <- errors.New("Hearthbeat timeout")
			}
		}()

		return latency, err
	}

	const pingInterval = 500 * time.Millisecond
	const initialPingsInterval = 5 * time.Millisecond

	pingsSend := 0

	timer := time.NewTimer(initialPingsInterval)
	defer timer.Stop()

	for {
		select {
		case <-terminate:
			return nil
		case <-timer.C:
			latencyResult, errResult := getSingleHearthbeatLatency()
			select {
			case <-terminate:
				return nil
			case latency := <-latencyResult:
				conn.latencyMutex.Lock()
				//move the history of latencies to the left, leaving the last index to the new latency
				conn.latencyHistory[0] = conn.latencyHistory[1]
				conn.latencyHistory[1] = conn.latencyHistory[2]
				conn.latencyHistory[2] = latency
				conn.latency = (conn.latencyHistory[0] + conn.latencyHistory[1] + conn.latencyHistory[2]) / 3
				conn.latencyMutex.Unlock()

				timer.Stop()

				//send 3 quick hearthbeats after a connection is established
				//to calculate latency asap
				if pingsSend < 3 {
					pingsSend++
					timer.Reset(initialPingsInterval)
				} else {
					timer.Reset(pingInterval)
				}
			case err := <-errResult:
				if err != nil {
					log.WithField("category", "heartbeat").Error(err)
				}

				return err
			}
		}
	}
}

func (a App) handleSocketClose(conn *clientConnection, terminate chan struct{}) error {
	ch := make(chan error)

	conn.Lock()

	conn.ws.SetCloseHandler(func(code int, text string) error {
		go func() {
			socketCloseErr := fmt.Errorf("Socket close %d %s", code, text)
			log.WithField("category", "socket close").Error(socketCloseErr)
			ch <- socketCloseErr
		}()

		return nil
	})

	conn.Unlock()

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

		defer func() {
			err := conn.close()
			if err != nil {
				log.WithField("category", "close ws connection").Error(err)
			}
		}()

		handlers := map[string]func(*clientConnection, chan struct{}) error{
			"handleSocketStoreNotify": a.handleSocketStoreNotify,
			"handleSocketClient":      a.handleSocketClient,
			"handleSocketHearthbeat":  a.handleSocketHearthbeat,
			"handleSocketClose":       a.handleSocketClose,
		}

		terminateHandler := make(chan struct{}, len(handlers))
		closeConnection := make(chan struct{}, len(handlers))

		var wg sync.WaitGroup
		wg.Add(len(handlers))

		for handlerName, handler := range handlers {
			go func(handlerName string, handler func(*clientConnection, chan struct{}) error, terminateHandler chan struct{}) {
				defer wg.Done()
				err := handler(conn, terminateHandler)
				log.WithFields(log.Fields{
					"category":    "handler returned",
					"handlerName": handlerName,
					"connection":  conn.String(),
				}).Info(err)

				closeConnection <- struct{}{}
			}(handlerName, handler, terminateHandler)
		}

		go func() {
			<-closeConnection
			//terminate all handlers
			for i := 0; i < len(handlers); i++ {
				terminateHandler <- struct{}{}
			}
		}()

		wg.Wait()
		close(closeConnection)
		log.WithField("connection", conn.String()).Info("Closed connection")
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
			connectionsCountsMutex.Lock()
			defer connectionsCountsMutex.Unlock()

			conns.Lock()
			defer conns.Unlock()

			l := len(conns.connections)
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

		defer ws.Close()

		countCh := make(chan int)

		defer func() {
			connectionsCountsMutex.Lock()
			index := funk.IndexOf(connectionsCounts, countCh)
			connectionsCounts = append(connectionsCounts[:index], connectionsCounts[index+1:]...)
			close(countCh)
			connectionsCountsMutex.Unlock()
		}()

		connectionsCountsMutex.Lock()
		connectionsCounts = append(connectionsCounts, countCh)
		connectionsCountsMutex.Unlock()

		//send the stats on the initial connection
		if err := ws.WriteJSON(stats{len(conns.connections)}); err != nil {
			log.WithField("category", "write initial stats").Error(err)
			return
		}

		for {
			select {
			case c := <-countCh:
				if err := ws.WriteJSON(stats{c}); err != nil {
					log.WithField("category", "write stats").Error(err)
					return
				}
			}
		}
	}
}

const serverPort = ":8082"

func (a App) startServer() error {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/db", a.dbHandler(upgrader))
	mux.HandleFunc("/stats", a.statsHandler(upgrader))

	handler := cors.AllowAll().Handler(mux)

	log.WithField("port", serverPort).Info("Server Listening")
	return http.ListenAndServe(serverPort, handler)
}
