package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/pkg/errors"
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

type rpcPluginCallerImpl struct {
	plugins gosdk.Plugins
}

func NewRPCPluginCaller() RPCPluginCaller {
	return &rpcPluginCallerImpl{
		plugins: make(gosdk.Plugins, 0),
	}
}

type pluginImpl struct {
	priority int
	name     string
	timeout  time.Duration
	lb       lb.LoadBalance
	// api.PluginServiceClient
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
	endpoint, err := p.getEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	cn, err := endpoint.Get(ctx)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_connection")
	}
	defer endpoint.AfterTransform(ctx, cn.((goloadbalancer.Connection)))
	conn := cn.(goloadbalancer.Connection).ConnInstance().(*grpc.ClientConn)
	plugin := api.NewPluginServiceClient(conn)
	anyReq, err := anypb.New(req.(proto.Message))
	if err != nil {
		return nil, errors.New(fmt.Errorf("new any to plugin error: %w", err), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "new_any")

	}
	rsp, err := plugin.Call(ctx, &api.PluginRequest{
		Request:        anyReq,
		FullMethodName: info.FullMethod,
	})
	if err != nil {
		return nil, errors.New(fmt.Errorf("call plugin error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "call_plugin")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	for k, v := range rsp.Metadata {
		md.Append(k, v)
	}

	newRequest := rsp.NewRequest
	if newRequest != nil {
		err = newRequest.UnmarshalTo(req.(proto.Message))
		if err != nil {
			return nil, errors.New(fmt.Errorf("unmarshal to request error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "unmarshal_to_request")
		}
	}

	ctx = metadata.NewIncomingContext(ctx, md)
	return handler(ctx, req)
}
func (p *pluginImpl) getEndpoint(ctx context.Context) (lb.Endpoint, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New(fmt.Errorf("get metadata from context error"), int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_metadata")
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
		return nil, errors.New(fmt.Errorf("select endpoint error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "select_endpoint")
	}
	return endpoint, nil

}
func (p *pluginImpl) Call(ctx context.Context, in *api.PluginRequest, opts ...grpc.CallOption) (*api.PluginResponse, error) {

	endpoint, err := p.getEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	cn, err := endpoint.Get(ctx)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_connection")
	}
	defer endpoint.AfterTransform(ctx, cn.((goloadbalancer.Connection)))
	conn := cn.(goloadbalancer.Connection).ConnInstance().(*grpc.ClientConn)

	plugin := api.NewPluginServiceClient(conn)
	return plugin.Call(ctx, in, opts...)
}
func (p *pluginImpl) GetPluginInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*api.PluginInfo, error) {
	endpoint, err := p.getEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	cn, err := endpoint.Get(ctx)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_connection")
	}
	defer endpoint.AfterTransform(ctx, cn.((goloadbalancer.Connection)))
	conn := cn.(goloadbalancer.Connection).ConnInstance().(*grpc.ClientConn)
	plugin := api.NewPluginServiceClient(conn)
	return plugin.GetPluginInfo(ctx, in, opts...)
}
func (p *pluginImpl) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	grpcStream := NewGrpcPluginStream(ss, info.FullMethod, ss.Context(), p)
	return handler(srv, grpcStream)

}
