package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	loadbalance "github.com/begonia-org/go-loadbalancer"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
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

type IOType struct {
	In  protoreflect.MessageDescriptor
	Out protoreflect.MessageDescriptor
}
type GrpcProxyMiddleware func(srv interface{}, serverStream grpc.ServerStream) error
type GrpcProxy struct {
	lb *GrpcLoadBalancer
	// middlewares []grpc.StreamServerInterceptor
	// chainStreamInts []grpc.StreamServerInterceptor
	streamInt         grpc.StreamClientInterceptor
	chainClientStream []grpc.StreamClientInterceptor
	ioType            map[string]*IOType
	log               logger.Logger
}

func NewGrpcProxy(lb *GrpcLoadBalancer, log logger.Logger, middlewares ...grpc.StreamClientInterceptor) *GrpcProxy {
	g := &GrpcProxy{
		lb: lb,
		// middlewares: middlewares,
		chainClientStream: middlewares,
		ioType:            make(map[string]*IOType),
		log:               log,
	}
	g.chainStreamClientInterceptors()
	return g
}
func (g *GrpcProxy) Register(lb loadbalance.LoadBalance, pd ProtobufDescription) {
	g.lb.Register(lb, pd)
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

// chainStreamServerInterceptors chains all stream server interceptors into one.
func (g *GrpcProxy) chainStreamClientInterceptors() {
	// Prepend opts.streamInt to the chaining interceptors if it exists, since streamInt will
	// be executed before any other chained interceptors.
	interceptors := g.chainClientStream
	if g.streamInt != nil {
		interceptors = append([]grpc.StreamClientInterceptor{g.streamInt}, g.chainClientStream...)
	}

	var chainedInt grpc.StreamClientInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = g.chainStreamInterceptors(interceptors)
	}

	g.streamInt = chainedInt
}

func (g *GrpcProxy) chainStreamInterceptors(interceptors []grpc.StreamClientInterceptor) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return interceptors[0](ctx, desc, cc, method, g.getChainStreamHandler(interceptors, 0, streamer))
	}
}

func (g *GrpcProxy) getChainStreamHandler(interceptors []grpc.StreamClientInterceptor, curr int, finalHandler grpc.Streamer) grpc.Streamer {
	if curr == len(interceptors)-1 {
		return finalHandler
	}
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return interceptors[curr+1](ctx, desc, cc, method, g.getChainStreamHandler(interceptors, curr+1, finalHandler))
	}
}

func (g *GrpcProxy) Do(srv interface{}, serverStream grpc.ServerStream) error {

	return g.Handler(srv, serverStream)
}
func (g *GrpcProxy) Handler(srv interface{}, serverStream grpc.ServerStream) error {
	defer func() {
		if p := recover(); p != nil {
			s := debug.Stack()
			g.log.Errorf(serverStream.Context(), "panic recover! p: %v stack:%s", p, s)
		}
	}()
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
	md, ok := metadata.FromIncomingContext(serverStream.Context())
	if !ok {
		md = metadata.MD{}
	}

	md.Set("X-Forwarded-For", strings.Join(xForwards, ","))
	clientCtx = metadata.NewOutgoingContext(clientCtx, md)
	var clientStream grpc.ClientStream = nil

	if len(g.chainClientStream) > 0 && g.streamInt != nil {
		clientStream, err = g.streamInt(clientCtx, proxyDesc, conn, fullMethodName, g.getChainStreamHandler(g.chainClientStream, 0, func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return grpc.NewClientStream(ctx, desc, cc, method, opts...)
		}))
		if err != nil {
			return fmt.Errorf("failed creating client stream from stream client int: %v", err)
		}
	} else {
		clientStream, err = grpc.NewClientStream(clientCtx, proxyDesc, conn, fullMethodName)
		if err != nil {
			return err
		}
	}

	// proxyStream := &proxyClientStream{ClientStream: clientStream, ctx: clientCtx}
	// 转发流量
	// 从客户端到服务端
	s2cErrChan := g.forwardServerToClient(serverStream, clientStream)

	// 从服务端到客户端
	c2sErrChan := g.forwardClientToServer(clientStream, serverStream)
	for i := 0; i < 2; i++ {
		select {
		case s2cErr := <-s2cErrChan:
			if errors.Is(s2cErr, io.EOF) {
				// this is the happy case where the sender has encountered io.EOF, and won't be sending anymore./
				// the clientStream>serverStream may continue pumping though.
				// log.Printf("s2cErr:%v", s2cErr)
				err = clientStream.CloseSend()
				if err != nil {
					return fmt.Errorf("failed closing client stream: %w", err)
				}
			} else {
				// however, we may have gotten a receive error (stream disconnected, a read error etc) in which case we need
				// to cancel the clientStream to the backend, let all of its goroutines be freed up by the CancelFunc and
				// exit with an error to the stack
				clientCancel()
				return fmt.Errorf("failed proxying s2c: %w", s2cErr)
			}
		case c2sErr := <-c2sErrChan:
			// This happens when the clientStream has nothing else to offer (io.EOF), returned a gRPC error. In those two
			// cases we may have received Trailers as part of the call. In case of other errors (stream closed) the trailers
			// will be nil.
			serverStream.SetTrailer(clientStream.Trailer())
			// c2sErr will contain RPC error from client code. If not io.EOF return the RPC error as server stream error.
			if !errors.Is(c2sErr, io.EOF) {
				// log.Printf("c2sErr:%v", c2sErr)
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
		defer func() {
			if p := recover(); p != nil {
				s := debug.Stack()
				g.log.Errorf(dst.Context(), "panic recover! p: %v stack:%s", p, s)
				ret <- fmt.Errorf("panic recover! p: %v stack:%s", p, s)
			}
		}()
		method, _ := grpc.Method(dst.Context())
		io := g.ioType[method]
		f := dynamicpb.NewMessage(io.Out)
		// f := &emptypb.Empty{}

		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- fmt.Errorf("fail forward client to server:%w", err) // this can be io.EOF which is happy case
				break
			}
			if i == 0 {
				// This is a bit of a hack, but client to server headers are only readable after first client msg is
				// received but must be written to server stream before the first msg is flushed.
				// This is the only place to do it nicely
				// 先转发header
				// inMD, _ := metadata.FromIncomingContext(src.Context())
				// outMD, _ := metadata.FromOutgoingContext(src.Context())

				// md := metadata.Join(inMD, outMD)
				// m := make(map[string][]string)

				md, err := src.Header()
				if err != nil {
					ret <- fmt.Errorf("failed reading header from client stream: %w", err)
					break
				}

				if err := dst.SendHeader(md); err != nil {
					ret <- fmt.Errorf("failed sending header to server stream: %w", err)
					break
				}
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- fmt.Errorf("failed sending msg to server stream: %w", err)
				break
			}
		}
	}()
	return ret
}

func (g *GrpcProxy) forwardServerToClient(src grpc.ServerStream, dst grpc.ClientStream) chan error {
	ret := make(chan error, 1)
	go func() {
		method, _ := grpc.Method(dst.Context())
		io := g.ioType[method]
		f := dynamicpb.NewMessage(io.In)

		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- fmt.Errorf("recv msg error:%w from client forward", err) // this can be io.EOF which is happy case
				break
			}

			if err := dst.SendMsg(f); err != nil {
				ret <- fmt.Errorf("failed sending msg to client stream: %w", err)
				break
			}
		}
	}()
	return ret
}

func (g *GrpcProxy) buildServiceDesc(pd ProtobufDescription) {
	fd := pd.GetFileDescriptorSet()
	// sds := make([]*grpc.ServiceDesc, 0)
	for _, file := range fd.GetFile() {
		sd := file.GetService()

		for _, service := range sd {

			methods := service.GetMethod()
			for _, method := range methods {
				in := method.GetInputType()
				inDesc := pd.GetMessageTypeByFullName(strings.TrimPrefix(in, "."))
				outDesc := pd.GetMessageTypeByFullName(strings.TrimPrefix(method.GetOutputType(), "."))
				srv := fmt.Sprintf("/%s.%s/%s", file.GetPackage(), service.GetName(), method.GetName())
				g.ioType[srv] = &IOType{
					In:  inDesc,
					Out: outDesc,
				}

			}

		}
	}
}
