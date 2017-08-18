package server

import (
	"net/http"

	"github.com/gngeorgiev/liquiddb/cmd/liquiddb/client_connection"
	"github.com/gngeorgiev/liquiddb/cmd/liquiddb/pool"
	log "github.com/sirupsen/logrus"

	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type stats struct {
	Connections []string `json:"connections,omitempty"`
}

func (a App) dbHandler(upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		conn := client_connection.NewWsClientConnection(ws)
		a.dbConnectionHandler(conn)
	}
}

func (a App) statsHandler(upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	statsWorkerPool := pool.NewWorkerPool(4, 1*time.Second)

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
					//not be in a situation where a channel blocks.
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
			case <-clientConnectionsPool.ConnectionsUpdated():
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

func (a App) StartWsServer(serverPort string) error {
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

	log.WithField("port", serverPort).Info("WS server Listening")
	return http.ListenAndServe(serverPort, handler)
}
