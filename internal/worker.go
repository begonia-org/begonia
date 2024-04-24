package internal

import (
	"context"
	"time"

	"github.com/begonia-org/begonia/internal/daemon"
	"github.com/begonia-org/begonia/transport"
)

type GatewayWorker interface {
	// Start the worker
	Start() error
}

type GatewayWorkerImpl struct {
	// data
	daemon daemon.Daemon

	server *transport.GatewayServer
}

func NewGatewayWorkerImpl(daemon daemon.Daemon, server *transport.GatewayServer) GatewayWorker {
	return &GatewayWorkerImpl{
		daemon: daemon,
		server: server,
	}
}

func (g *GatewayWorkerImpl) Start() error {

	g.daemon.Start(context.Background())
	time.Sleep(time.Second * 2)
	return g.server.Start()
}
