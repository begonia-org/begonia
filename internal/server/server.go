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
	"github.com/begonia-org/begonia/internal/pkg/middleware/serialization"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	"github.com/begonia-org/begonia/internal/service"
	dp "github.com/begonia-org/dynamic-proto"
	common "github.com/begonia-org/go-sdk/common/api/v1"
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
func NewGateway(cfg *dp.GatewayConfig, conf *config.Config, services []service.Service, pluginApply *middleware.PluginsApply) *dp.GatewayServer {
	// 参数选项
	opts := &dp.GrpcServerOptions{
		Middlewares:     make([]dp.GrpcProxyMiddleware, 0),
		Options:         make([]grpc.ServerOption, 0),
		PoolOptions:     make([]pool.PoolOptionsBuildOption, 0),
		HttpMiddlewares: make([]runtime.ServeMuxOption, 0),
		HttpHandlers:    make([]func(http.Handler) http.Handler, 0),
	}
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("application/json", serialization.NewResponseJSONMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("multipart/form-data", serialization.NewFormDataMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("application/x-www-form-urlencoded", serialization.NewFormUrlEncodedMarshaler()))

	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMetadata(middleware.IncomingHeadersToMetadata))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithErrorHandler(middleware.HandleErrorWithLogger(logger.Logger)))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithForwardResponseOption(middleware.HttpResponseBodyModify))
	// 连接池配置
	opts.PoolOptions = append(opts.PoolOptions, pool.WithMaxActiveConns(100))
	opts.PoolOptions = append(opts.PoolOptions, pool.WithPoolSize(128))
	// 中间件配置
	opts.Options = append(opts.Options, grpc.ChainUnaryInterceptor(pluginApply.UnaryInterceptorChains()...))
	opts.Options = append(opts.Options, grpc.ChainStreamInterceptor(pluginApply.StreamInterceptorChains()...))

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
	err = pd.SetHttpResponse(common.E_HttpResponse)
	if err != nil {
		panic(err)
	}
	routersList := routers.Get()
	for _, srv := range services {
		err := gw.RegisterLocalService(context.Background(), pd, srv.Desc(), srv)
		if err != nil {
			panic(err)
		}
		for _, method := range srv.Desc().Methods {
			routersList.AddLocalSrv(fmt.Sprintf("/%s/%s", srv.Desc().ServiceName, method.MethodName))
		}

	}
	routersList.LoadAllRouters(pd)

	return gw
}
