package transport

import (
	"context"
	"strings"

	loadbalance "github.com/begonia-org/go-loadbalancer"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type HttpForwardGrpcEndpoint interface {
	Request(req GrpcRequest) (proto.Message, runtime.ServerMetadata, error)
	ServerSideStream(req GrpcRequest) (ServerSideStream, error)
	ClientSideStream(req GrpcRequest) (ClientSideStream, error)
	Stream(req GrpcRequest) (StreamClient, error)
}

type httpForwardGrpcEndpointImpl struct {
	httpMatchPattern map[string]string
	pool             loadbalance.Pool
}

func NewEndpoint(pool loadbalance.Pool) HttpForwardGrpcEndpoint {
	return &httpForwardGrpcEndpointImpl{
		httpMatchPattern: make(map[string]string),
		pool:             pool,
	}
}

// request is the request message with method, path, body and query params.
// 发起普通的请求请求
func (e *httpForwardGrpcEndpointImpl) Request(req GrpcRequest) (proto.Message, runtime.ServerMetadata, error) {
	var metadata runtime.ServerMetadata
	cc, err := e.pool.Get(req.GetContext())
	defer e.pool.Release(req.GetContext(), cc)
	if err != nil {
		return nil, runtime.ServerMetadata{}, err
	}
	conn := cc.ConnInstance().(*grpc.ClientConn)
	out := req.GetOut()
	in := req.GetIn()
	ctx := req.GetContext()
	err = conn.Invoke(ctx, req.GetFullMethodName(), in, out, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return out, metadata, err

}

// 发起客户端流式服务
func (e *httpForwardGrpcEndpointImpl) clientSideStream(ctx context.Context, desc *grpc.StreamDesc, out protoreflect.MessageDescriptor, fullName string, opts ...grpc.CallOption) (ClientSideStream, error) {
	cc, err := e.pool.Get(ctx)
	defer e.pool.Release(ctx, cc)
	if err != nil {
		return nil, err
	}
	conn := cc.ConnInstance().(*grpc.ClientConn)
	stream, err := conn.NewStream(ctx, desc, fullName, opts...)
	if err != nil {
		return nil, err

	}
	x := &clientSideStreamClient{
		ClientStream: stream,
		out:          out,
	}
	return x, nil
}

func (e *httpForwardGrpcEndpointImpl) stream(ctx context.Context, desc *grpc.StreamDesc, fullName string, out protoreflect.MessageDescriptor, opts ...grpc.CallOption) (StreamClient, error) {
	cc, err := e.pool.Get(ctx)
	defer e.pool.Release(ctx, cc)
	if err != nil {
		return nil, err
	}
	conn := cc.ConnInstance().(*grpc.ClientConn)
	stream, err := conn.NewStream(ctx, desc, fullName, opts...)
	if err != nil {
		return nil, err
	}
	x := &streamClient{stream, out}
	return x, nil
}
func (e *httpForwardGrpcEndpointImpl) createStreamDesc(fullName string, server bool, client bool) *grpc.StreamDesc {
	// fullname := req.GetFullMethodName()
	method := fullName[strings.LastIndex(fullName, "/")+1:]
	return &grpc.StreamDesc{
		StreamName:    method,
		ServerStreams: server,
		ClientStreams: client,
	}
}

// 请求服务端流式服务
func (e *httpForwardGrpcEndpointImpl) ServerSideStream(req GrpcRequest) (ServerSideStream, error) {
	desc := e.createStreamDesc(req.GetFullMethodName(), true, false)
	cc, err := e.pool.Get(req.GetContext())
	defer e.pool.Release(req.GetContext(), cc)
	if err != nil {
		return nil, err
	}
	conn := cc.ConnInstance().(*grpc.ClientConn)
	stream, err := conn.NewStream(req.GetContext(), desc, req.GetFullMethodName(), req.GetCallOptions()...)
	if err != nil {
		return nil, err
	}
	x := &serverSideStreamClient{
		ClientStream: stream,
		out:          req.GetOutType(),
	}

	if err := x.ClientStream.SendMsg(req.GetIn()); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// 请求客户端流式服务
func (e *httpForwardGrpcEndpointImpl) ClientSideStream(req GrpcRequest) (ClientSideStream, error) {

	desc := e.createStreamDesc(req.GetFullMethodName(), false, true)
	return e.clientSideStream(req.GetContext(), desc, req.GetOutType(), req.GetFullMethodName(), req.GetCallOptions()...)

}

// 双向流式服务
func (e *httpForwardGrpcEndpointImpl) Stream(req GrpcRequest) (StreamClient, error) {
	desc := e.createStreamDesc(req.GetFullMethodName(), true, true)

	return e.stream(req.GetContext(), desc, req.GetFullMethodName(), req.GetOutType(), req.GetCallOptions()...)

}
