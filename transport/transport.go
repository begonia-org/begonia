package transport

import "sync"

var onceGW sync.Once

var gateway *GatewayServer

func New(cfg *GatewayConfig, opts *GrpcServerOptions) *GatewayServer {
	onceGW.Do(func() {

		gateway = NewGateway(cfg, opts)
	})
	return gateway
}
func Get() *GatewayServer {
	return gateway
}
