package main

import (
	deadlock "github.com/sasha-s/go-deadlock"
)

//TODO: abstract this to work with an interface
//TODO: the connection pool will probably later be responsible for distributing work
type ConnectionPool struct {
	connectionsLock    deadlock.RWMutex
	connections        []*clientConnection
	connectionsUpdated chan int
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connectionsLock:    deadlock.RWMutex{},
		connections:        make([]*clientConnection, 0),
		connectionsUpdated: make(chan int, 10),
	}
}

func (p *ConnectionPool) AddConnection(c *clientConnection) {
	p.connectionsLock.Lock()
	p.connections = append(p.connections, c)
	p.connectionsUpdated <- len(p.connections)
	p.connectionsLock.Unlock()
}

func (p *ConnectionPool) RemoveConnection(c *clientConnection) {
	p.connectionsLock.Lock()

	index := -1
	for i, conn := range p.connections {
		if conn == c {
			index = i
			break
		}
	}

	if index != -1 {
		p.connections = append(p.connections[:index], p.connections[index+1:]...)
	}

	p.connectionsUpdated <- len(p.connections)
	p.connectionsLock.Unlock()
}

func (p *ConnectionPool) Len() int {
	p.connectionsLock.RLock()
	defer p.connectionsLock.RUnlock()

	return len(p.connections)
}

func (p *ConnectionPool) Connections() []*clientConnection {
	p.connectionsLock.RLock()
	defer p.connectionsLock.RUnlock()

	connectionsBuffer := make([]*clientConnection, len(p.connections))
	copy(connectionsBuffer, p.connections)

	return connectionsBuffer
}
