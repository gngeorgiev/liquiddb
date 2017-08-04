package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gngeorgiev/liquiddb"
	log "github.com/sirupsen/logrus"
)

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
	dataCh := make(chan operationClientData, 10)
	errorCh := make(chan error)

	go func() {
		for {
			var data operationClientData
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

		case err := <-errorCh:
			return err
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
