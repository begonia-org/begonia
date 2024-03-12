package server

import (
	"context"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/gateway"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	"github.com/begonia-org/begonia/internal/pkg/middleware"
	"github.com/begonia-org/begonia/internal/pkg/middleware/validator"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	"github.com/begonia-org/begonia/internal/service"
	dp "github.com/begonia-org/dynamic-proto"
	"github.com/google/wire"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spark-lence/tiga/pool"
	"google.golang.org/grpc"
)

var ProviderSet = wire.NewSet(NewGatewayConfig, NewGateway)

func NewGatewayConfig(gw string) *dp.GatewayConfig {
	return &dp.GatewayConfig{
		GrpcProxyAddr: ":19527",
		GatewayAddr:   gw,
	}
}
func NewGateway(cfg *dp.GatewayConfig, conf *config.Config, services []service.Service, validate *validator.APIValidator) *dp.GatewayServer {
	opts := &dp.GrpcServerOptions{
		Middlewares:     make([]dp.GrpcProxyMiddleware, 0),
		Options:         make([]grpc.ServerOption, 0),
		PoolOptions:     make([]pool.PoolOptionsBuildOption, 0),
		HttpMiddlewares: make([]runtime.ServeMuxOption, 0),
	}
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithErrorHandler(middleware.HandleErrorWithLogger(logger.Logger)))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMetadata(middleware.IncomingHeadersToMetadata))
	opts.PoolOptions = append(opts.PoolOptions, pool.WithMaxActiveConns(100))
	opts.PoolOptions = append(opts.PoolOptions, pool.WithPoolSize(128))
	opts.Options = append(opts.Options, grpc.ChainUnaryInterceptor(validate.ValidateUnaryInterceptor))
	opts.Options = append(opts.Options, grpc.ChainStreamInterceptor(validate.ValidateStreamInterceptor))

	runtime.WithMetadata(middleware.IncomingHeadersToMetadata)
	gw := gateway.New(cfg, opts)
	protos := conf.GetProtosDir()
	pd, err := dp.NewDescription(protos)
	if err != nil {
		panic(err)
	}
	// api.RegisterFileServiceServer()
	routersList := routers.Get()
	for _, srv := range services {
		err := gw.RegisterLocalService(context.Background(), pd, srv.Desc(), srv)
		if err != nil {
			panic(err)
		}
		routersList.LoadAllRouters(pd)
	}
	return gw
}
