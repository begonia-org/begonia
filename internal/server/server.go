package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

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
	_, port, _ := net.SplitHostPort(gw)
	p, _ := strconv.Atoi(port)
	return &dp.GatewayConfig{
		GrpcProxyAddr: fmt.Sprintf(":%d", p+1),
		GatewayAddr:   gw,
	}
}
func NewGateway(cfg *dp.GatewayConfig, conf *config.Config, services []service.Service, validate *validator.APIValidator, logM *middleware.LoggerMiddleware) *dp.GatewayServer {
	// 参数选项
	opts := &dp.GrpcServerOptions{
		Middlewares:     make([]dp.GrpcProxyMiddleware, 0),
		Options:         make([]grpc.ServerOption, 0),
		PoolOptions:     make([]pool.PoolOptionsBuildOption, 0),
		HttpMiddlewares: make([]runtime.ServeMuxOption, 0),
		HttpHandlers:    make([]func(http.Handler) http.Handler, 0),
	}
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("application/json", middleware.NewResponseJSONMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithErrorHandler(middleware.HandleErrorWithLogger(logger.Logger)))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMetadata(middleware.IncomingHeadersToMetadata))
	// 连接池配置
	opts.PoolOptions = append(opts.PoolOptions, pool.WithMaxActiveConns(100))
	opts.PoolOptions = append(opts.PoolOptions, pool.WithPoolSize(128))
	// 中间件配置
	opts.Options = append(opts.Options, grpc.ChainUnaryInterceptor(logM.LoggerUnaryInterceptor, validate.ValidateUnaryInterceptor))
	opts.Options = append(opts.Options, grpc.ChainStreamInterceptor(logM.LoggerStreamInterceptor, validate.ValidateStreamInterceptor))

	cors := &middleware.CorsMiddleware{
		Cors: conf.GetCorsConfig(),
	}
	opts.HttpHandlers = append(opts.HttpHandlers, cors.Handle)
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
	}
	// gw.RegisterService(context.Background(), pd, )
	routersList.LoadAllRouters(pd)
	// // 构建负载均衡器
	// endpoints := make([]loadbalance.Endpoint, 0)
	// pool := dp.NewGrpcConnPool("0.0.0:12139")
	// endpoint := dp.NewGrpcEndpoint("0.0.0.0:12139", pool)
	// endpoints = append(endpoints, endpoint)
	// rr := loadbalance.NewRoundRobinBalance(endpoints)
	// err = gw.RegisterService(context.Background(), pd, rr)
	// if err != nil {
	// 	panic(err)
	// }
	return gw
}
