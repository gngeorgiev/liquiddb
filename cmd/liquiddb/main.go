package main

import (
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/gngeorgiev/liquiddb"
	"github.com/gngeorgiev/liquiddb/cmd/liquiddb/server"
)

//tests for this package will be written when the Go driver is created
func main() {
	app := server.NewApp(liquiddb.New())

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	var serversWg sync.WaitGroup
	//we should exit if any of the servers crashes
	//at least for now
	serversWg.Add(1)

	go func() {
		defer serversWg.Done()

		log.Error(app.StartWsServer(":8082"))
	}()

	go func() {
		defer serversWg.Done()

		log.Error(app.StartTcpServer(":8083"))
	}()

	//TODO: proper server shutdown
	serversWg.Wait()
}
