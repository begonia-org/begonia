package server

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/wetrycode/begonia/internal/pkg/logger"
	"github.com/wetrycode/begonia/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GatewayServer struct {
	http *http.Server
	grpc *grpc.Server
}

func NewDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
}
func NewGrpcServerOptions()[]grpc.ServerOption{
	return []grpc.ServerOption{
		grpc.Creds(insecure.NewCredentials()),
	}
}
func NewServiceOptions(user *service.UsersService, file *service.FileService, endpoint string, opts []grpc.DialOption) []service.ServiceOptions {
	return []service.ServiceOptions{
		service.WithFileService(file, opts),
		service.WithUserService(user, opts),
	}
}
func New(mux *runtime.ServeMux, handle http.Handler,grpcServer *grpc.Server, endpoint string, opts []service.ServiceOptions) *GatewayServer {
	server := NewHttpServer(endpoint, handle)
	for _, opt := range opts {
		err := opt(grpcServer, mux, endpoint)
		if err != nil {
			logger.Logger.Error(err)
			panic(err)
		}
	}
	return &GatewayServer{
		http: server,
		grpc: grpcServer,
	}
}
func (g *GatewayServer) Start() error {
	logger.Logger.Info("Gateway Server Start on ", g.http.Addr, " ...")
	return g.http.ListenAndServe()
}
