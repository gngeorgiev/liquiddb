package main

import (
	"encoding/json"
	"io"
	"net"
	"time"
)

type tcpClientConnection struct {
	*clientConnection

	conn net.Conn
}

func newTcpClientConnection(conn net.Conn) ClientConnection {
	cc := newClientConnection()

	return &tcpClientConnection{
		cc,
		conn,
	}
}

func (c *tcpClientConnection) String() string {
	return c.conn.RemoteAddr().String()
}

func (c *tcpClientConnection) Close() error {
	return c.conn.Close()
}

func (c *tcpClientConnection) WriteJSON(o interface{}) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}

	if deadlineErr := c.conn.SetWriteDeadline(time.Now().Add(1 * time.Second)); deadlineErr != nil {
		return deadlineErr
	}

	_, writeErr := c.conn.Write(b)
	return writeErr
}

func (c *tcpClientConnection) ReadJSON(o interface{}) error {
	data := make([]byte, 0, 4096)
	buf := make([]byte, 0, 1024)

	if deadlineErr := c.conn.SetReadDeadline(time.Now().Add(1 * time.Second)); deadlineErr != nil {
		return deadlineErr
	}

	for {
		n, err := c.conn.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
	}

	return json.Unmarshal(data, o)
}
