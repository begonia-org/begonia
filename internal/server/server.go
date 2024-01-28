package server

import (
	"github.com/google/wire"
	"google.golang.org/grpc"
)

var ProviderSet = wire.NewSet(NewGatewayMux,NewGrpcServerOptions, NewDialOptions, grpc.NewServer, NewHandlers, NewServiceOptions, New)
