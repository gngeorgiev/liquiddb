package main

import (
	"log"

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

	log.Fatal(app.startServer())
}
