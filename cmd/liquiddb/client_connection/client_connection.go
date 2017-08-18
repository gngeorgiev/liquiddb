package client_connection

import (
	"time"

	"github.com/gngeorgiev/liquiddb"
	"github.com/gngeorgiev/liquiddb/cmd/liquiddb/operations"
	deadlock "github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"
)

type ClientConnection interface {
	WriteInterested(path string, o liquiddb.EventData) (bool, error)
	AddInterest(interest string, op liquiddb.EventOperation, o operations.OperationClientData) error
	RemoveInterest(interest string, op liquiddb.EventOperation, o operations.OperationClientData)

	GetLatencyHistory() [3]int32
	SetLatencyHistory([3]int32)
	GetLatency() int32
	SetLatency(int32)

	HearthbeatResponse() chan struct{}

	WriteJSON(o interface{}) error
	ReadJSON(o interface{}) error
	Close() error

	String() string
}

type clientConnection struct {
	interestsMutex deadlock.Mutex
	interests      map[string][]*operations.ClientInterest

	latencyHistoryMutex deadlock.Mutex
	latencyHistory      [3]int32

	latencyMutex deadlock.Mutex
	latency      int32

	hearthbeatResponse chan struct{}
}

func newClientConnection() *clientConnection {
	c := &clientConnection{
		interestsMutex: deadlock.Mutex{},
		interests:      map[string][]*operations.ClientInterest{},

		latencyHistoryMutex: deadlock.Mutex{},
		latencyHistory:      [3]int32{},

		latencyMutex:       deadlock.Mutex{},
		latency:            0,
		hearthbeatResponse: make(chan struct{}),
	}

	return c
}

func (c *clientConnection) HearthbeatResponse() chan struct{} {
	return c.hearthbeatResponse
}

func (c *clientConnection) GetLatencyHistory() [3]int32 {
	c.latencyHistoryMutex.Lock()
	defer c.latencyHistoryMutex.Unlock()

	return c.latencyHistory
}

func (c *clientConnection) SetLatencyHistory(l [3]int32) {
	c.latencyHistoryMutex.Lock()
	c.latencyHistory = l
	c.latencyHistoryMutex.Unlock()
}

func (c *clientConnection) GetLatency() int32 {
	c.latencyMutex.Lock()
	defer c.latencyMutex.Unlock()

	return c.latency
}

func (c *clientConnection) SetLatency(l int32) {
	c.latencyMutex.Lock()
	c.latency = l
	c.latencyMutex.Unlock()
}

func (c *clientConnection) WriteInterested(path string, o liquiddb.EventData) (bool, error) {
	c.interestsMutex.Lock()
	defer c.interestsMutex.Unlock()

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
			"id":        interest.Id,
			"operation": interest.Operation,
			"timestamp": interest.Timestamp,
		}).Debug("Interest")

		interestHasValidTimestamp := o.Timestamp.After(interest.Timestamp) || o.Timestamp.Equal(interest.Timestamp)
		if interest.Operation == op && interestHasValidTimestamp {
			return true, nil
		}
	}

	return false, nil
}

func (c *clientConnection) AddInterest(interest string, op liquiddb.EventOperation, o operations.OperationClientData) error {
	c.interestsMutex.Lock()
	defer c.interestsMutex.Unlock()

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

	cInterest := &operations.ClientInterest{
		o.ID,
		op,
		t,
	}
	if interests == nil {
		c.interests[interest] = []*operations.ClientInterest{cInterest}
	} else {
		c.interests[interest] = append(interests, cInterest)
	}

	return nil
}

func (c *clientConnection) RemoveInterest(interest string, op liquiddb.EventOperation, o operations.OperationClientData) {
	c.interestsMutex.Lock()
	defer c.interestsMutex.Unlock()

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
		if cInterest.Operation == op && cInterest.Id == o.ID {
			c.interests[interest] = append(interests[:i], interests[i+1:]...)
		}
	}

	if len(c.interests[interest]) == 0 {
		delete(c.interests, interest)
	}
}
