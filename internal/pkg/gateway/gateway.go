package gateway

import (
	"sync"

	"github.com/begonia-org/begonia/transport"
)

var onceGW sync.Once

var gateway *transport.GatewayServer

func New(cfg *transport.GatewayConfig, opts *transport.GrpcServerOptions) *transport.GatewayServer {
	onceGW.Do(func() {

		gateway = transport.NewGateway(cfg, opts)
	})
	return gateway
}
func Get() *transport.GatewayServer {
	return gateway
}
