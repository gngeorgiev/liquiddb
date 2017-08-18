package server

import "github.com/gngeorgiev/liquiddb"

//App is the app that will be ran when the CLI is started
type App struct {
	db *liquiddb.LiquidDb
}

func NewApp(db *liquiddb.LiquidDb) *App {
	return &App{db}
}
