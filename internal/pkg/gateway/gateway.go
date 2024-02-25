package gateway

import (
	"sync"

	dp "github.com/begonia-org/dynamic-proto"
)

var onceGW sync.Once

var gateway *dp.GatewayServer

func New(cfg *dp.GatewayConfig, opts *dp.GrpcServerOptions) *dp.GatewayServer {
	onceGW.Do(func() {

		gateway = dp.NewGateway(cfg, opts)
	})
	return gateway
}
func Get() *dp.GatewayServer {
	return gateway
}
