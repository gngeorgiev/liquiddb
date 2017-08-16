package main

import (
	"time"

	"github.com/gngeorgiev/liquiddb"
	deadlock "github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"
)

type ClientConnection interface {
	WriteInterested(path string, o liquiddb.EventData) (bool, error)
	AddInterest(interest string, op liquiddb.EventOperation, o operationClientData) error
	RemoveInterest(interest string, op liquiddb.EventOperation, o operationClientData)

	GetLatencyHistory() [3]int32
	SetLatencyHistory([3]int32)
	GetLatency() int32
	SetLatency(int32)

	HearthbeatResponse() chan struct{}

	WriteJSON(o interface{}) error
	ReadJSON(o interface{}) error
	Close() error
	SetCloseHandler(func(code int, text string) error)

	String() string
}

type clientConnection struct {
	interestsMutex deadlock.Mutex
	interests      map[string][]*clientInterest

	latencyHistoryMutex deadlock.Mutex
	latencyHistory      [3]int32

	latencyMutex deadlock.Mutex
	latency      int32

	hearthbeatResponse chan struct{}
}

func newClientConnection() *clientConnection {
	c := &clientConnection{
		interestsMutex: deadlock.Mutex{},
		interests:      map[string][]*clientInterest{},

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
			"id":        interest.id,
			"operation": interest.operation,
			"timestamp": interest.timestamp,
		}).Debug("Interest")

		interestHasValidTimestamp := o.Timestamp.After(interest.timestamp) || o.Timestamp.Equal(interest.timestamp)
		if interest.operation == op && interestHasValidTimestamp {
			return true, nil
		}
	}

	return false, nil
}

func (c *clientConnection) AddInterest(interest string, op liquiddb.EventOperation, o operationClientData) error {
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

func (c *clientConnection) RemoveInterest(interest string, op liquiddb.EventOperation, o operationClientData) {
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
		if cInterest.operation == op && cInterest.id == o.ID {
			c.interests[interest] = append(interests[:i], interests[i+1:]...)
		}
	}

	if len(c.interests[interest]) == 0 {
		delete(c.interests, interest)
	}
}
