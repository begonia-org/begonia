package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/gateway"
	"github.com/begonia-org/begonia/internal/pkg/middleware"
	"github.com/begonia-org/begonia/internal/pkg/middleware/serialization"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	"github.com/begonia-org/begonia/internal/service"
	"github.com/begonia-org/begonia/transport"
	loadbalance "github.com/begonia-org/go-loadbalancer"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	"github.com/google/wire"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

var ProviderSet = wire.NewSet(NewGatewayConfig, NewGateway)

func NewGatewayConfig(gw string) *transport.GatewayConfig {
	_, port, _ := net.SplitHostPort(gw)
	p, _ := strconv.Atoi(port)
	return &transport.GatewayConfig{
		GrpcProxyAddr: fmt.Sprintf(":%d", p+1),
		GatewayAddr:   gw,
	}
}
func readDesc(conf *config.Config) (transport.ProtobufDescription, error) {
	desc := conf.GetLocalAPIDesc()
	bin, err := os.ReadFile(desc)
	if err != nil {
		return nil, fmt.Errorf("read desc file error:%w", err)
	}
	pd, err := transport.NewDescriptionFromBinary(bin,filepath.Dir(desc))
	if err != nil {
		return nil, err
	}
	err = pd.SetHttpResponse(common.E_HttpResponse)
	if err != nil {
		return nil, err
	}
	return pd, nil
}
func NewGateway(cfg *transport.GatewayConfig, conf *config.Config, services []service.Service, pluginApply *middleware.PluginsApply) *transport.GatewayServer {
	// 参数选项
	opts := &transport.GrpcServerOptions{
		Middlewares:     make([]transport.GrpcProxyMiddleware, 0),
		Options:         make([]grpc.ServerOption, 0),
		PoolOptions:     make([]loadbalance.PoolOptionsBuildOption, 0),
		HttpMiddlewares: make([]runtime.ServeMuxOption, 0),
		HttpHandlers:    make([]func(http.Handler) http.Handler, 0),
	}
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("application/json", serialization.NewJSONMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("multipart/form-data", serialization.NewFormDataMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("application/x-www-form-urlencoded", serialization.NewFormUrlEncodedMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption(runtime.MIMEWildcard, serialization.NewRawBinaryUnmarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("application/octet-stream", serialization.NewRawBinaryUnmarshaler()))

	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMetadata(middleware.IncomingHeadersToMetadata))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithErrorHandler(middleware.HandleErrorWithLogger(logger.Log)))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithForwardResponseOption(middleware.HttpResponseBodyModify))
	// opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithRoutingErrorHandler(middleware.HandleRoutingError))
	// 连接池配置
	opts.PoolOptions = append(opts.PoolOptions, loadbalance.WithMaxActiveConns(100))
	opts.PoolOptions = append(opts.PoolOptions, loadbalance.WithPoolSize(128))
	// 中间件配置
	opts.Options = append(opts.Options, grpc.ChainUnaryInterceptor(pluginApply.UnaryInterceptorChains()...))
	opts.Options = append(opts.Options, grpc.ChainStreamInterceptor(pluginApply.StreamInterceptorChains()...))

	cors := &middleware.CorsMiddleware{
		Cors: conf.GetCorsConfig(),
	}
	opts.HttpHandlers = append(opts.HttpHandlers, cors.Handle)
	runtime.WithMetadata(middleware.IncomingHeadersToMetadata)
	gw := gateway.New(cfg, opts)

	pd, err := readDesc(conf)
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
