package main

import deadlock "github.com/sasha-s/go-deadlock"

//TODO: abstract this to work with an interface
//TODO: the connection pool will probably later be responsible for distributing work
type ConnectionPool struct {
	connectionsLock    deadlock.RWMutex
	connections        []*clientConnection
	connectionsCount   int
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
	defer p.connectionsLock.Unlock()

	p.connections = append(p.connections, c)
	p.connectionsCount = len(p.connections)

	select {
	case p.connectionsUpdated <- p.connectionsCount:
	default:
	}
}

func (p *ConnectionPool) RemoveConnection(c *clientConnection) {
	p.connectionsLock.Lock()
	defer p.connectionsLock.Unlock()

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

	p.connectionsCount = len(p.connections)

	p.connectionsUpdated <- p.connectionsCount
}

func (p *ConnectionPool) Len() int {
	p.connectionsLock.RLock()
	defer p.connectionsLock.RUnlock()

	return p.connectionsCount
}

func (p *ConnectionPool) Connections() []*clientConnection {
	p.connectionsLock.RLock()
	defer p.connectionsLock.RUnlock()

	return p.connections
}
