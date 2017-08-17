package main

import (
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
)

func (a App) startTcpServer(serverPort string) error {
	server, err := net.Listen("tcp", serverPort)
	if err != nil {
		return err
	}

	log.WithField("port", serverPort).Info("TCP server Listening")

	var connectionsWg sync.WaitGroup
	for {
		c, err := server.Accept()
		if err != nil {
			return err
		}

		connectionsWg.Add(1)
		go func() {
			defer connectionsWg.Done()

			conn := newTcpClientConnection(c)
			a.dbConnectionHandler(conn)
		}()
	}

	//let's try to wait for the connections to properly close
	//before closing the whole server
	connectionsWg.Wait()

	return server.Close()
}
