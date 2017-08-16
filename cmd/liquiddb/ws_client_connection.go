package main

import (
	"github.com/gorilla/websocket"
	"github.com/sasha-s/go-deadlock"
)

type wsClientConnection struct {
	*clientConnection

	wsMutex deadlock.Mutex
	ws      *websocket.Conn
}

func newWsClientConnection(ws *websocket.Conn) ClientConnection {
	cc := newClientConnection()

	return &wsClientConnection{
		cc,
		deadlock.Mutex{},
		ws,
	}
}

func (c *wsClientConnection) String() string {
	return c.ws.RemoteAddr().String()
}

func (c *wsClientConnection) Close() error {
	c.wsMutex.Lock()
	defer c.wsMutex.Unlock()

	return c.ws.Close()
}

func (c *wsClientConnection) WriteJSON(o interface{}) error {
	c.wsMutex.Lock()
	defer c.wsMutex.Unlock()

	return c.ws.WriteJSON(o)
}

func (c *wsClientConnection) ReadJSON(o interface{}) error {
	return c.ws.ReadJSON(o)
}

func (c *wsClientConnection) SetCloseHandler(handler func(code int, text string) error) {
	c.wsMutex.Lock()
	c.ws.SetCloseHandler(handler)
	c.wsMutex.Unlock()
}
