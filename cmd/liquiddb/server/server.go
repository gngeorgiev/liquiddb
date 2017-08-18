package server

import (
	"github.com/gngeorgiev/liquiddb/cmd/liquiddb/pool"
)

var clientConnectionsPool = pool.NewConnectionPool()
