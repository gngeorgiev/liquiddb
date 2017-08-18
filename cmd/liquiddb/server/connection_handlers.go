package server

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gngeorgiev/liquiddb"
	"github.com/gngeorgiev/liquiddb/cmd/liquiddb/client_connection"
	"github.com/gngeorgiev/liquiddb/cmd/liquiddb/operations"
	log "github.com/sirupsen/logrus"
)

//TODO: Use protocol buffers!
func (a App) handleSocketStoreNotify(conn client_connection.ClientConnection, terminate chan struct{}) error {
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
				err = conn.WriteJSON(op)
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

func (a App) handleSocketClient(conn client_connection.ClientConnection, terminate chan struct{}) error {
	dataCh := make(chan operations.OperationClientData, 10)
	errorCh := make(chan error)

	go func() {
		for {
			var data operations.OperationClientData
			err := conn.ReadJSON(&data)
			if err != nil {
				//TODO: try to write one last error to the ws connection before closing it
				log.WithField("category", "read").Error(err)
				errorCh <- err
				return
			}

			dataCh <- data
		}
	}()

	for {
		select {
		case <-terminate:
			return nil
		case data := <-dataCh:
			if data.Operation != operations.HearthbeatResponseOperation {
				log.WithField("data", data).Debug("Received data")
			}

			switch data.Operation {
			case operations.ClientOperationSet:
				a.db.Link(data.ID).SetPath(data.Path, data.Value)
			case operations.ClientOperationDelete:
				a.db.Link(data.ID).Delete(data.Path)
			case operations.ClientOperationGet:
				a.db.Link(data.ID).Get(data.Path)
			case operations.ClientOperationSubscribe:
				op := liquiddb.EventOperation(data.Value.(string))
				//TODO: can we optimize this strings join?
				if err := conn.AddInterest(strings.Join(data.Path, "."), op, data); err != nil {
					log.WithField("category", "add interest").Error(err)
					return err
				}
			case operations.ClientOperationUnSubscribe:
				op := liquiddb.EventOperation(data.Value.(string))
				//TODO: can we optimize this strings join?
				conn.RemoveInterest(strings.Join(data.Path, "."), op, data)
			case operations.HearthbeatResponseOperation:
				conn.HearthbeatResponse() <- struct{}{}
			default:
				//TODO: should we and how to notify the user about this
				log.WithFields(log.Fields{
					"category":  "read",
					"operation": data.Operation,
				}).Error("Invalid operation type")
			}

		case err := <-errorCh:
			return err
		}
	}
}

func (a App) handleSocketHearthbeat(conn client_connection.ClientConnection, terminate chan struct{}) error {
	//TODO: refactor this method a bit as it has become too large
	//also refactor the whole file as it has also become too large
	sendHearthbeat := func() error {
		err := conn.WriteJSON(struct {
			Operation string `json:"operation,omitempty"`
			Timestamp string `json:"timestamp,omitempty"`
		}{
			Operation: operations.HearthbeatOperation,
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
			case <-conn.HearthbeatResponse():
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
				latencyHistory := conn.GetLatencyHistory()
				//move the history of latencies to the left, leaving the last index to the new latency
				latencyHistory[0] = latencyHistory[1]
				latencyHistory[1] = latencyHistory[2]
				latencyHistory[2] = latency
				conn.SetLatencyHistory(latencyHistory)
				conn.SetLatency((latencyHistory[0] + latencyHistory[1] + latencyHistory[2]) / 3)

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

func (a App) dbConnectionHandler(conn client_connection.ClientConnection) {
	clientConnectionsPool.AddConnection(conn)

	log.WithField("address", conn.String()).Info("New Connection")

	defer func() {
		err := conn.Close()
		clientConnectionsPool.RemoveConnection(conn)
		if err != nil {
			log.WithField("category", "close ws connection").Error(err)
		}
	}()

	handlers := map[string]func(client_connection.ClientConnection, chan struct{}) error{
		"handleSocketStoreNotify": a.handleSocketStoreNotify,
		"handleSocketClient":      a.handleSocketClient,
		"handleSocketHearthbeat":  a.handleSocketHearthbeat,
	}

	terminateHandler := make(chan struct{}, len(handlers))
	closeConnection := make(chan struct{}, len(handlers))

	var wg sync.WaitGroup
	wg.Add(len(handlers))

	for handlerName, handler := range handlers {
		go func(handlerName string, handler func(client_connection.ClientConnection, chan struct{}) error, terminateHandler chan struct{}) {
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
