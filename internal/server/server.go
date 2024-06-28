package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/middleware"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	"github.com/begonia-org/begonia/internal/service"
	loadbalance "github.com/begonia-org/go-loadbalancer"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/google/wire"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

var ProviderSet = wire.NewSet(NewGatewayConfig, NewGateway)

func NewGatewayConfig(gw string) *gateway.GatewayConfig {
	_, port, _ := net.SplitHostPort(gw)
	p, _ := strconv.Atoi(port)
	return &gateway.GatewayConfig{
		GrpcProxyAddr: fmt.Sprintf(":%d", p+1),
		GatewayAddr:   gw,
	}
}
func readDesc(conf *config.Config) (gateway.ProtobufDescription, error) {
	desc := conf.GetLocalAPIDesc()
	log.Printf("read desc file:%s", desc)
	bin, err := os.ReadFile(desc)
	if err != nil {
		return nil, fmt.Errorf("read desc file error:%w", err)
	}
	pd, err := gateway.NewDescriptionFromBinary(bin, filepath.Dir(desc))
	if err != nil {
		return nil, err
	}
	err = pd.SetHttpResponse(common.E_HttpResponse)
	if err != nil {
		return nil, err
	}
	return pd, nil
}
func NewGateway(cfg *gateway.GatewayConfig, conf *config.Config, services []service.Service, pluginApply *middleware.PluginsApply) *gateway.GatewayServer {
	// 参数选项
	opts := &gateway.GrpcServerOptions{
		Middlewares:     make([]gateway.GrpcProxyMiddleware, 0),
		Options:         make([]grpc.ServerOption, 0),
		PoolOptions:     make([]loadbalance.PoolOptionsBuildOption, 0),
		HttpMiddlewares: make([]runtime.ServeMuxOption, 0),
		HttpHandlers:    make([]func(http.Handler) http.Handler, 0),
	}
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("application/json", gateway.NewJSONMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("multipart/form-data", gateway.NewFormDataMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("application/x-www-form-urlencoded", gateway.NewFormUrlEncodedMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption(runtime.MIMEWildcard, gateway.NewRawBinaryUnmarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMarshalerOption("application/octet-stream", gateway.NewRawBinaryUnmarshaler()))

	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithMetadata(gateway.IncomingHeadersToMetadata))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithErrorHandler(gateway.HandleErrorWithLogger(gateway.Log)))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithForwardResponseOption(gateway.HttpResponseBodyModify))
	// opts.HttpMiddlewares = append(opts.HttpMiddlewares, runtime.WithRoutingErrorHandler(middleware.HandleRoutingError))
	// 连接池配置
	opts.PoolOptions = append(opts.PoolOptions, loadbalance.WithMaxActiveConns(100))
	opts.PoolOptions = append(opts.PoolOptions, loadbalance.WithPoolSize(128))
	// 中间件配置
	opts.Options = append(opts.Options, grpc.ChainUnaryInterceptor(pluginApply.UnaryInterceptorChains()...))
	opts.Options = append(opts.Options, grpc.ChainStreamInterceptor(pluginApply.StreamInterceptorChains()...))
	pd, err := readDesc(conf)
	if err != nil {
		panic(err)
	}
	cors := &gateway.CorsHandler{
		Cors: conf.GetCorsConfig(),
	}
	opts.HttpHandlers = append(opts.HttpHandlers, cors.Handle)
	gw := gateway.New(cfg, opts)

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
