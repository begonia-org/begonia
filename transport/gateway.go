package transport

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	loadbalance "github.com/begonia-org/go-loadbalancer"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)



type GrpcServerOptions struct {
	Middlewares     []GrpcProxyMiddleware
	Options         []grpc.ServerOption
	PoolOptions     []loadbalance.PoolOptionsBuildOption
	HttpMiddlewares []runtime.ServeMuxOption
	HttpHandlers    []func(http.Handler) http.Handler
}
type GatewayConfig struct {
	GatewayAddr   string
	GrpcProxyAddr string
}
type GatewayServer struct {
	grpcServer  *grpc.Server
	httpGateway HttpEndpoint
	proxyLB     *GrpcLoadBalancer
	gatewayMux  *runtime.ServeMux
	addr        string
	proxyAddr   string
	opts        *GrpcServerOptions
	mux         *sync.Mutex
}


func NewGrpcServer(opts *GrpcServerOptions, lb *GrpcLoadBalancer) *grpc.Server {

	proxy := NewGrpcProxy(lb, opts.Middlewares...)

	opts.Options = append(opts.Options, grpc.UnknownServiceHandler(proxy.Handler))
	return grpc.NewServer(opts.Options...)
}

func NewHttpServer(addr string, poolOpt ...loadbalance.PoolOptionsBuildOption) (HttpEndpoint, error) {
	pool := NewGrpcConnPool(addr, poolOpt...)

	endpoint := NewEndpoint(pool)

	return NewHttpEndpoint(endpoint)


}
func NewGateway(cfg *GatewayConfig, opts *GrpcServerOptions) *GatewayServer {
	lb := NewGrpcLoadBalancer()
	grpcServer := NewGrpcServer(opts, lb)
	_, port, _ := net.SplitHostPort(cfg.GrpcProxyAddr)
	proxy := fmt.Sprintf("127.0.0.1:%s", port)

	mux := runtime.NewServeMux(
		opts.HttpMiddlewares...,
	)
	httpGateway, err := NewHttpServer(proxy, opts.PoolOptions...)
	if err != nil {
		panic(err)
	}
	gatewayS := &GatewayServer{
		grpcServer:  grpcServer,
		httpGateway: httpGateway,
		proxyLB:     lb,
		gatewayMux:  mux,
		addr:        cfg.GatewayAddr,
		proxyAddr:   cfg.GrpcProxyAddr,
		opts:        opts,
		mux:         &sync.Mutex{},
	}
	// })
	return gatewayS
}

func (g *GatewayServer) RegisterService(ctx context.Context, pd ProtobufDescription, lb loadbalance.LoadBalance) error {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.proxyLB.Register(lb, pd)

	return g.httpGateway.RegisterHandlerClient(ctx, pd, g.gatewayMux)
}
func (g *GatewayServer) RegisterLocalService(ctx context.Context, pd ProtobufDescription, sd *grpc.ServiceDesc, ss any) error {
	g.grpcServer.RegisterService(sd, ss)
	return g.httpGateway.RegisterHandlerClient(ctx, pd, g.gatewayMux)
}
func (g *GatewayServer) DeleteLocalService(pd ProtobufDescription) {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.proxyLB.Delete(pd)
	_= g.DeleteHandlerClient(context.Background(), pd)
}
func (g *GatewayServer) GetLoadbalanceName() loadbalance.BalanceType {
	return g.proxyLB.Name()
}
func (g *GatewayServer) RegisterHandlerClient(ctx context.Context, pd ProtobufDescription) error {
	return g.httpGateway.RegisterHandlerClient(ctx, pd, g.gatewayMux)
}
func (g *GatewayServer) Start() {
	handler := h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			g.grpcServer.ServeHTTP(w, r)
		} else {
			g.gatewayMux.ServeHTTP(w, r)
		}
	}), &http2.Server{})

	for _, h := range g.opts.HttpHandlers {
		handler = h(handler)

	}
	s := &http.Server{
		Addr:    g.addr,
		Handler: handler,
	}

	lis, err := net.Listen("tcp", g.proxyAddr)
	if err != nil {
		panic(err)
	}
	go func() {
		err = g.grpcServer.Serve(lis)
		if err != nil {
			panic(err)

		}
	}()
	time.Sleep(1 * time.Second)
	log.Printf("Start on %s\n", g.addr)
	if err := s.ListenAndServe(); err != nil {
		panic(err)
		// return fmt.Errorf("start server error:%w", err)
	}
}
func (g *GatewayServer) GetOptions() *GrpcServerOptions {
	return g.opts
}
func (g *GatewayServer) DeleteLoadBalance(pd ProtobufDescription) {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.proxyLB.Delete(pd)
	// g.httpGateway.DeleteEndpoint(ctx, pd, mux)
}

func (g *GatewayServer) DeleteHandlerClient(ctx context.Context, pd ProtobufDescription) error {

	return g.httpGateway.DeleteEndpoint(ctx, pd, g.gatewayMux)
}

func (g *GatewayServer) UpdateLoadbalance(pd ProtobufDescription, lb loadbalance.LoadBalance) {
	g.proxyLB.Register(lb, pd)
}
