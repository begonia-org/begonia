package internal

import (
	"context"
	"time"

	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/daemon"
)

type GatewayWorker interface {
	// Start the worker
	Start()
}

type GatewayWorkerImpl struct {
	// data
	daemon daemon.Daemon

	server *gateway.GatewayServer
}

func NewGatewayWorkerImpl(daemon daemon.Daemon, server *gateway.GatewayServer) GatewayWorker {
	return &GatewayWorkerImpl{
		daemon: daemon,
		server: server,
	}
}

func (g *GatewayWorkerImpl) Start() {

	g.daemon.Start(context.Background())
	time.Sleep(time.Second * 2)
	g.server.Start()
	
}
