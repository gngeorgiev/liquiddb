package main

import (
	"net/http"
	"sync"

	deadlock "github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"

	"time"

	"github.com/gngeorgiev/liquiddb"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
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

type clientInterest struct {
	id        uint64
	operation liquiddb.EventOperation
	timestamp time.Time
}

var clientConnectionsPool = NewConnectionPool()

func (a App) dbHandler(upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		conn := newClientConnection(ws)
		clientConnectionsPool.AddConnection(conn)

		defer func() {
			err := conn.Close()
			clientConnectionsPool.RemoveConnection(conn)
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
	var connectionsCountsMutex deadlock.Mutex
	connectionsCounts := make([]chan int, 0)

	go func() {
		for {
			count := <-clientConnectionsPool.connectionsUpdated
			connectionsCountsMutex.Lock()

			for _, ch := range connectionsCounts {
				ch <- count
			}

			connectionsCountsMutex.Unlock()
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
			defer connectionsCountsMutex.Unlock()

			index := -1
			for i, ch := range connectionsCounts {
				if ch == countCh {
					index = i
				}
			}

			if index != -1 {
				connectionsCounts = append(connectionsCounts[:index], connectionsCounts[index+1:]...)
			} else {
				log.Error("Stats count channel not found in collection")
			}

			close(countCh)
		}()

		connectionsCountsMutex.Lock()
		connectionsCounts = append(connectionsCounts, countCh)
		connectionsCountsMutex.Unlock()

		connectionsCount := clientConnectionsPool.Len()

		//send the stats on the initial connection
		if err := ws.WriteJSON(stats{connectionsCount}); err != nil {
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
