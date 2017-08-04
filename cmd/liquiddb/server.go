package main

import (
	"net/http"
	"sync"

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

		log.WithField("address", conn.String()).Info("New Connection")

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
	Connections []string `json:"connections,omitempty"`
}

func (a App) statsHandler(upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	statsWorkerPool := NewWorkerPool(4, 1*time.Second)

	newConnection := make(chan chan []string)
	removeConnection := make(chan chan []string)
	publishConnectionsInfo := make(chan []string)

	connectionsChannels := make([]chan []string, 0)

	updateConnections := func() {
		clientConnections := clientConnectionsPool.Connections()
		connectionsInfo := make([]string, len(clientConnections))
		for i, c := range clientConnections {
			connectionsInfo[i] = c.String()
		}

		publishConnectionsInfo <- connectionsInfo
	}

	go func() {
		for {
			select {
			case newC := <-newConnection:
				connectionsChannels = append(connectionsChannels, newC)
			case oldC := <-removeConnection:
				index := -1
				for i, ch := range connectionsChannels {
					if ch == oldC {
						index = i
						break
					}
				}

				if index != -1 {
					connectionsChannels = append(connectionsChannels[:index], connectionsChannels[index+1:]...)
				} else {
					log.Error("Stats count channel not found in collection")
				}
			case info := <-publishConnectionsInfo:
				for _, ch := range connectionsChannels {
					//delegating the work to the pool will allow us to
					//not be in a situation where a channel blocks
					//this should not happen, but still, if it does
					//the whole stats section of the application
					//will deadlock. We are also distributing
					//the work to more goroutines
					func(ch chan []string) {
						statsWorkerPool.Schedule(func() {
							ch <- info
						})
					}(ch)
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-clientConnectionsPool.connectionsUpdated:
				updateConnections()
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.WithField("address", ws.RemoteAddr().String()).Info("New Stats Connection")

		connectionsChannel := make(chan []string, 10)
		newConnection <- connectionsChannel

		defer func() {
			removeConnection <- connectionsChannel
			err := ws.Close()
			if err != nil {
				log.WithField("category", "close stats connection").Error(err)
			}
		}()

		go updateConnections()

		for {
			select {
			case c := <-connectionsChannel:
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
