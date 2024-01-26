package server

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"github.com/wetrycode/begonia/internal/pkg/logger"
	"github.com/wetrycode/begonia/internal/pkg/middlerware"
)

type GatewayServer struct {
	server *http.Server
}

func NewGatewayServer(mux *runtime.ServeMux, config *config.Config, endpoint string, vildator *middlerware.APIVildator) *GatewayServer {
	cors := config.GetCorsConfig()
	server := NewHttpServer(endpoint, cors, vildator, mux)
	return &GatewayServer{
		server: server,
	}
}

func (g *GatewayServer) Start() error {
	logger.Logger.Info("Gateway Server Start on ", g.server.Addr, " ...")
	return g.server.ListenAndServe()
}
