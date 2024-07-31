package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	lb "github.com/begonia-org/go-loadbalancer"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/plugin/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RPCPluginCaller interface{}

type pluginImpl struct {
	priority int
	name     string
	timeout  time.Duration
	lb       lb.LoadBalance
}

func (p *pluginImpl) SetPriority(priority int) {
	p.priority = priority
}
func (p *pluginImpl) Priority() int {
	return p.priority
}
func (p *pluginImpl) Name() string {
	return p.name

}

func (p *pluginImpl) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(make(map[string]string))
	}
	rsp, header, err := p.Apply(ctx, req, info.FullMethod)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("call plugin error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "call_plugin")
	}
	for k, v := range header {
		md[k] = append(md[k], v...)

	}
	newRequest := rsp.NewRequest
	if newRequest != nil {
		err = newRequest.UnmarshalTo(req.(proto.Message))
		if err != nil {
			return nil, gosdk.NewError(fmt.Errorf("unmarshal to request error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "unmarshal_to_request")
		}
	}

	ctx = metadata.NewIncomingContext(ctx, md)
	return handler(ctx, req)
}
func (p *pluginImpl) getEndpoint(ctx context.Context) (lb.Endpoint, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, gosdk.NewError(fmt.Errorf("get metadata from context error"), int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_metadata")
	}
	xforwardeds := md.Get("X-Forwarded-For")
	clientIP := ""
	if p, ok := peer.FromContext(ctx); ok {
		clientIP = p.Addr.String()
		clientIP = strings.Split(clientIP, ":")[0]
	}
	if len(xforwardeds) > 0 {
		clientIP = xforwardeds[0]
	}
	endpoint, err := p.lb.Select(clientIP)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("select endpoint error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "select_endpoint")
	}
	return endpoint, nil

}
func (p *pluginImpl) Apply(ctx context.Context, in interface{}, fullMethodName string) (*api.PluginResponse, metadata.MD, error) {

	endpoint, err := p.getEndpoint(ctx)
	if err != nil {
		return nil, nil, err
	}
	cn, err := endpoint.Get(ctx)
	if err != nil {
		return nil, nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_connection")
	}
	defer endpoint.AfterTransform(ctx, cn.((goloadbalancer.Connection)))
	conn := cn.(goloadbalancer.Connection).ConnInstance().(*grpc.ClientConn)

	plugin := api.NewPluginServiceClient(conn)
	anyReq, err := anypb.New(in.(proto.Message))
	if err != nil {
		return nil, nil, gosdk.NewError(fmt.Errorf("new any to plugin error: %w", err), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "new_any")

	}
	var header, trailer metadata.MD
	rsp, err := plugin.Apply(ctx, &api.PluginRequest{
		Request:        anyReq,
		FullMethodName: fullMethodName,
	}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return nil, nil, gosdk.NewError(fmt.Errorf("call plugin error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "call_plugin")
	}
	return rsp, header, nil
}
func (p *pluginImpl) Metadata(ctx context.Context, in *emptypb.Empty) (metadata.MD, error) {
	endpoint, err := p.getEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	cn, err := endpoint.Get(ctx)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_connection")
	}
	defer endpoint.AfterTransform(ctx, cn.((goloadbalancer.Connection)))
	conn := cn.(goloadbalancer.Connection).ConnInstance().(*grpc.ClientConn)
	plugin := api.NewPluginServiceClient(conn)
	var header, trailer metadata.MD
	_, err = plugin.Metadata(ctx, in, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("call plugin metadata error:%w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "metadata")
	}
	return header, nil

}
func (p *pluginImpl) Info(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*api.PluginInfo, error) {
	endpoint, err := p.getEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	cn, err := endpoint.Get(ctx)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_connection")
	}
	defer endpoint.AfterTransform(ctx, cn.((goloadbalancer.Connection)))
	conn := cn.(goloadbalancer.Connection).ConnInstance().(*grpc.ClientConn)
	plugin := api.NewPluginServiceClient(conn)
	return plugin.Info(ctx, in, opts...)
}
func (p *pluginImpl) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, err := p.Metadata(ss.Context(), &emptypb.Empty{})
	if err != nil {
		return err

	}
	in, _ := metadata.FromIncomingContext(ss.Context())

	ctx := metadata.NewIncomingContext(ss.Context(), metadata.Join(in, md))
	grpcStream := NewGrpcPluginStream(ss, info.FullMethod, ctx, p)
	if grpcStream != nil {
		defer grpcStream.Release()

	}
	return handler(srv, grpcStream)

}
func (p *pluginImpl) StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

	return streamer(ctx, desc, cc, method, opts...)

}
func NewPluginImpl(lb lb.LoadBalance, name string, timeout time.Duration) *pluginImpl {
	return &pluginImpl{
		lb:      lb,
		name:    name,
		timeout: timeout,
	}
}
