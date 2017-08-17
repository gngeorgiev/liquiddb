package main

import (
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/gngeorgiev/liquiddb"
)

//App is the app that will be ran when the CLI is started
type App struct {
	db *liquiddb.LiquidDb
}

//tests for this package will be written when the Go driver is created
func main() {
	app := App{
		db: liquiddb.New(),
	}

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	var serversWg sync.WaitGroup
	//we should exit if any of the servers crashes
	//at least for now
	serversWg.Add(1)

	go func() {
		defer serversWg.Done()

		log.Fatal(app.startWsServer(":8082"))
	}()

	go func() {
		defer serversWg.Done()

		log.Fatal(app.startTcpServer(":8083"))
	}()

	serversWg.Wait()
}
