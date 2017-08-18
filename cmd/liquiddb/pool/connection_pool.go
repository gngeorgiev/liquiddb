package pool

import (
	"github.com/gngeorgiev/liquiddb/cmd/liquiddb/client_connection"
	deadlock "github.com/sasha-s/go-deadlock"
)

//TODO: abstract this to work with an interface
//TODO: the connection pool will probably later be responsible for distributing work
type ConnectionPool struct {
	connectionsLock    deadlock.RWMutex
	connections        []client_connection.ClientConnection
	connectionsUpdated chan int
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connectionsLock:    deadlock.RWMutex{},
		connections:        make([]client_connection.ClientConnection, 0),
		connectionsUpdated: make(chan int, 10),
	}
}

func (p *ConnectionPool) ConnectionsUpdated() <-chan int {
	return p.connectionsUpdated
}

func (p *ConnectionPool) AddConnection(c client_connection.ClientConnection) {
	p.connectionsLock.Lock()
	p.connections = append(p.connections, c)
	p.connectionsUpdated <- len(p.connections)
	p.connectionsLock.Unlock()
}

func (p *ConnectionPool) RemoveConnection(c client_connection.ClientConnection) {
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

func (p *ConnectionPool) Connections() []client_connection.ClientConnection {
	p.connectionsLock.RLock()
	defer p.connectionsLock.RUnlock()

	connectionsBuffer := make([]client_connection.ClientConnection, len(p.connections))
	copy(connectionsBuffer, p.connections)

	return connectionsBuffer
}
