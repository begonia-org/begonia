package transport

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	loadbalance "github.com/begonia-org/go-loadbalancer"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type grpcEndpointImpl struct {
	addr string
	pool loadbalance.Pool
}
type EndpointServer struct {
	Addr   string
	Weight int
}

// NewGrpcConnPool 创建一个grpc连接池
func NewGrpcConnPool(addr string, poolOpt ...loadbalance.PoolOptionsBuildOption) loadbalance.Pool {
	opts := loadbalance.NewPoolOptions(func(ctx context.Context) (loadbalance.Connection, error) {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		return loadbalance.NewConnectionImpl(conn, 30*time.Second, 10*time.Second), nil
	})
	opts.ConnectionUsedHook = append(opts.ConnectionUsedHook, loadbalance.ConnUseAt)
	for _, opt := range poolOpt {
		opt(opts)

	}
	pool := loadbalance.NewConnPool(opts)
	return pool

}

func NewGrpcEndpoint(addr string, pool loadbalance.Pool) loadbalance.Endpoint {
	return &grpcEndpointImpl{
		addr: addr,
		pool: pool,
	}
}
func (g *grpcEndpointImpl) AfterTransform(ctx context.Context, cn loadbalance.Connection) {
	g.pool.Release(ctx, cn)
}
func (g *grpcEndpointImpl) Stats() loadbalance.Stats {
	return g.pool.Stats()
}
func (g *grpcEndpointImpl) Addr() string {
	return g.addr
}

func (g *grpcEndpointImpl) Get(ctx context.Context) (interface{}, error) {
	return g.pool.Get(ctx)
}
func (g *grpcEndpointImpl) Close() error {
	return g.pool.Close()
}

type GrpcLoadBalancer struct {
	lb   map[string]loadbalance.LoadBalance
	mu   sync.Mutex
	name loadbalance.BalanceType
}

func NewGrpcLoadBalancer() *GrpcLoadBalancer {
	return &GrpcLoadBalancer{
		lb: make(map[string]loadbalance.LoadBalance),
	}
}

func (g *GrpcLoadBalancer) Register(lb loadbalance.LoadBalance, pd ProtobufDescription) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.name = loadbalance.BalanceType(lb.Name())
	fds := pd.GetFileDescriptorSet()
	for _, file := range fds.GetFile() { // 遍历所有文件描述符
		for _, service := range file.GetService() { // 遍历文件中的所有服务
			for _, method := range service.GetMethod() { // 遍历服务中的所有方法
				key := fmt.Sprintf("/%s.%s/%s", file.GetPackage(), service.GetName(), method.GetName())
				g.lb[strings.ToUpper(key)] = lb
			}
		}
	}
}

func (g *GrpcLoadBalancer) Name() loadbalance.BalanceType {
	return g.name
}
func (g *GrpcLoadBalancer) Delete(pd ProtobufDescription) {
	g.mu.Lock()
	defer g.mu.Unlock()
	fds := pd.GetFileDescriptorSet()
	for _, file := range fds.GetFile() { // 遍历所有文件描述符
		for _, service := range file.GetService() { // 遍历文件中的所有服务
			for _, method := range service.GetMethod() { // 遍历服务中的所有方法
				key := fmt.Sprintf("/%s.%s/%s", file.GetPackage(), service.GetName(), method.GetName())
				// 不直接关闭是为了防止正在使用的连接被关闭
				// 避免共享该负载均衡器的其他路由器出现问题
				// g.lb[strings.ToUpper(key)].Close()
				g.lb[strings.ToUpper(key)] = nil
				delete(g.lb, strings.ToUpper(key))
			}
		}
	}
}
func (g *GrpcLoadBalancer) Select(method string, args ...interface{}) (loadbalance.Endpoint, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if lb, ok := g.lb[strings.ToUpper(method)]; ok {
		endpoint, err := lb.Select(args...)
		return endpoint, err

	}
	return nil, loadbalance.ErrNoEndpoint
}
func (g *GrpcLoadBalancer) Stats() {

}

type GrpcProxyMiddleware func(srv interface{}, serverStream grpc.ServerStream) error
type GrpcProxy struct {
	lb          *GrpcLoadBalancer
	middlewares []GrpcProxyMiddleware
}

func NewGrpcProxy(lb *GrpcLoadBalancer, middlewares ...GrpcProxyMiddleware) *GrpcProxy {
	return &GrpcProxy{
		lb:          lb,
		middlewares: middlewares,
	}
}
func (g *GrpcProxy) getClientIP(ctx context.Context) (string, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return "", loadbalance.ErrNoSourceIP
	}
	peerAddrStr := peer.Addr.String()
	return peerAddrStr, nil
}
func (g *GrpcProxy) getXForward(ctx context.Context) []string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	return md.Get("X-Forwarded-For")
}

// func (g *GrpcProxy) forwardUnaryCall(ctx context.Context, method string, req interface{}, cc *grpc.ClientConn) (interface{}, error) {
// 	// 创建一个新的UnaryInvoker，它是用于发起Unary RPC调用的
// 	invoker := grpc.Invoke

// 	// 准备用于存储RPC响应的变量
// 	var resp interface{}

// 	// 调用实际服务
// 	err := invoker(ctx, method, req, &resp, cc)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return resp, nil
// }

// func (g *GrpcProxy) UnaryProxyInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 	fullMethodName := info.FullMethod
// 	log.Println("准备转发")
// 	// 选择一个端点，获取链接
// 	clientIP, err := g.getClientIP(ctx)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Unavailable, "no source ip")
// 	}
// 	xForwards := g.getXForward(ctx)
// 	if len(xForwards) > 0 {
// 		clientIP = xForwards[0]
// 	} else {
// 		xForwards = make([]string, 0)
// 	}
// 	endpoint, err := g.lb.Select(fullMethodName, clientIP)
// 	// log.Println("选择连接完成")
// 	if err != nil {
// 		return nil, status.Errorf(codes.Unavailable, "no endpoint available to select,%v", err)
// 	}
// 	cn, err := endpoint.Get(ctx)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Unavailable, "no endpoint available from endpoint,%v", err)
// 	}
// 	// 释放链接
// 	defer endpoint.AfterTransform(ctx, cn.((loadbalance.Connection)))

//		conn := cn.(loadbalance.Connection).ConnInstance().(*grpc.ClientConn)
//		// 将请求转发到目标服务
//		// 添加本地ip
//		local, _ := tiga.GetLocalIP()
//		xForwards = append(xForwards, local)
//		clientCtx := metadata.NewIncomingContext(ctx, metadata.Pairs("X-Forwarded-For", strings.Join(xForwards, ",")))
//		resp, err := g.forwardUnaryCall(clientCtx, info.FullMethod, req, conn)
//		if err != nil {
//			return nil, status.Errorf(codes.Internal, "failed to forward call: %v", err)
//		}
//		return resp, nil
//	}
func (g *GrpcProxy) Handler(srv interface{}, serverStream grpc.ServerStream) error {

	// 执行中间件
	for _, middleware := range g.middlewares {
		if err := middleware(srv, serverStream); err != nil {
			return err
		}
	}
	// 获取方法名
	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Internal, "stream not exists in context")
	}
	// 选择一个端点，获取链接
	clientIP, err := g.getClientIP(serverStream.Context())
	if err != nil {
		return status.Errorf(codes.Unavailable, "no source ip")
	}
	xForwards := g.getXForward(serverStream.Context())
	if len(xForwards) > 0 {
		clientIP = xForwards[0]
	} else {
		xForwards = make([]string, 0)
	}
	// 传入ip地址(一致性哈希负载均衡算法)和方法名，选择一个端点
	endpoint, err := g.lb.Select(fullMethodName, clientIP)
	if err != nil {
		return status.Errorf(codes.Unavailable, "no endpoint available to select,%v", err)
	}
	cn, err := endpoint.Get(serverStream.Context())
	if err != nil {
		return status.Errorf(codes.Unavailable, "no endpoint available from endpoint,%v", err)
	}
	// 释放链接
	defer endpoint.AfterTransform(serverStream.Context(), cn.((loadbalance.Connection)))

	conn := cn.(loadbalance.Connection).ConnInstance().(*grpc.ClientConn)
	clientCtx, clientCancel := context.WithCancel(serverStream.Context())
	defer clientCancel()
	proxyDesc := &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: true,
	}
	// 添加本地ip
	local, _ := tiga.GetLocalIP()
	xForwards = append(xForwards, local)
	md, ok := metadata.FromIncomingContext(clientCtx)
	if !ok {
		md = metadata.MD{}
	}
	md.Set("X-Forwarded-For", strings.Join(xForwards, ","))
	clientCtx = metadata.NewOutgoingContext(clientCtx, md)

	clientStream, err := grpc.NewClientStream(clientCtx, proxyDesc, conn, fullMethodName)
	if err != nil {
		return err
	}
	// 转发流量
	// 从客户端到服务端
	s2cErrChan := g.forwardServerToClient(serverStream, clientStream)
	// 从服务端到客户端
	c2sErrChan := g.forwardClientToServer(clientStream, serverStream)

	for i := 0; i < 2; i++ {
		select {
		case s2cErr := <-s2cErrChan:
			if s2cErr == io.EOF {
				// this is the happy case where the sender has encountered io.EOF, and won't be sending anymore./
				// the clientStream>serverStream may continue pumping though.
				err = clientStream.CloseSend()
				if err != nil {
					return status.Errorf(codes.Internal, "failed closing client stream: %v", err)
				}
			} else {
				// however, we may have gotten a receive error (stream disconnected, a read error etc) in which case we need
				// to cancel the clientStream to the backend, let all of its goroutines be freed up by the CancelFunc and
				// exit with an error to the stack
				clientCancel()
				return status.Errorf(codes.Internal, "failed proxying s2c: %v", s2cErr)
			}
		case c2sErr := <-c2sErrChan:
			// This happens when the clientStream has nothing else to offer (io.EOF), returned a gRPC error. In those two
			// cases we may have received Trailers as part of the call. In case of other errors (stream closed) the trailers
			// will be nil.
			serverStream.SetTrailer(clientStream.Trailer())
			// c2sErr will contain RPC error from client code. If not io.EOF return the RPC error as server stream error.
			if c2sErr != io.EOF {
				return c2sErr
			}
			return nil
		}
	}
	return status.Errorf(codes.Internal, "gRPC proxying should never reach this stage.")

}
func (g *GrpcProxy) forwardClientToServer(src grpc.ClientStream, dst grpc.ServerStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &emptypb.Empty{}
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if i == 0 {
				// This is a bit of a hack, but client to server headers are only readable after first client msg is
				// received but must be written to server stream before the first msg is flushed.
				// This is the only place to do it nicely
				// 先转发header
				md, err := src.Header()
				if err != nil {
					ret <- err
					break
				}
				if err := dst.SendHeader(md); err != nil {
					ret <- err
					break
				}
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func (g *GrpcProxy) forwardServerToClient(src grpc.ServerStream, dst grpc.ClientStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &emptypb.Empty{}
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}
