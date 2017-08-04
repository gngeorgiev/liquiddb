package main

import (
	"time"

	"github.com/gorilla/websocket"

	"github.com/gngeorgiev/liquiddb"
	deadlock "github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"
)

type clientConnection struct {
	deadlock.Mutex

	interests map[string][]*clientInterest

	latencyMutex       deadlock.Mutex
	latencyHistory     [3]int32
	latency            int32
	hearthbeatResponse chan struct{}

	ws *websocket.Conn
}

func newClientConnection(ws *websocket.Conn) *clientConnection {
	c := &clientConnection{
		Mutex: deadlock.Mutex{},

		interests: map[string][]*clientInterest{},

		latencyMutex:       deadlock.Mutex{},
		latencyHistory:     [3]int32{},
		latency:            0,
		hearthbeatResponse: make(chan struct{}),

		ws: ws,
	}

	return c
}

func (c *clientConnection) String() string {
	return c.ws.RemoteAddr().String()
}

func (c *clientConnection) Close() error {
	c.Lock()
	defer c.Unlock()

	return c.ws.Close()
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
	c.latencyMutex.Lock()
	t = t.Add(time.Duration(-c.latency) * time.Millisecond)
	c.latencyMutex.Unlock()

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
