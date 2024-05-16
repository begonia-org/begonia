package transport

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type GrpcRequest interface {
	GetContext() context.Context

	// TODO:with gateway
	GetMarshaler() runtime.Marshaler
	GetReq() *http.Request
	GetPathParams() map[string]string

	GetIn() proto.Message
	GetOut() proto.Message
	GetOutType() protoreflect.MessageDescriptor
	GetInType() protoreflect.MessageDescriptor
	GetFullMethodName() string
	GetCallOptions() []grpc.CallOption
}

type GrpcRequestOptions func(req *GrpcRequestImpl)

func WithGatewayMarshaler(m runtime.Marshaler) GrpcRequestOptions {
	return func(req *GrpcRequestImpl) {
		req.marshaler = m
	}
}
func WithGatewayReq(r *http.Request) GrpcRequestOptions {
	return func(req *GrpcRequestImpl) {
		req.req = r
	}
}
func WithGatewayPathParams(pathParams map[string]string) GrpcRequestOptions {
	return func(req *GrpcRequestImpl) {
		req.pathParams = pathParams
	}
}
func WithGatewayCallOptions(options ...grpc.CallOption) GrpcRequestOptions {
	return func(req *GrpcRequestImpl) {
		req.callOptions = options
	}
}
func WithIn(in proto.Message) GrpcRequestOptions {
	return func(req *GrpcRequestImpl) {
		req.in = in
	}
}
func WithOut(out proto.Message) GrpcRequestOptions {
	return func(req *GrpcRequestImpl) {
		req.out = out
	}

}

type GrpcRequestImpl struct {
	ctx            context.Context
	marshaler      runtime.Marshaler
	req            *http.Request
	pathParams     map[string]string
	in             proto.Message
	out            proto.Message
	inType         protoreflect.MessageDescriptor
	outType        protoreflect.MessageDescriptor
	fullMethodName string
	callOptions    []grpc.CallOption
}

func NewGrpcRequest(ctx context.Context, inType protoreflect.MessageDescriptor, outType protoreflect.MessageDescriptor, fullMethodName string, opts ...GrpcRequestOptions) GrpcRequest {
	req := &GrpcRequestImpl{
		ctx:            ctx,
		inType:         inType,
		outType:        outType,
		fullMethodName: fullMethodName,
		pathParams:     make(map[string]string),
		callOptions:    make([]grpc.CallOption, 0),
	}
	for _, opt := range opts {
		opt(req)

	}
	return req
}
func (g *GrpcRequestImpl) GetContext() context.Context {
	return g.ctx
}

func (g *GrpcRequestImpl) GetMarshaler() runtime.Marshaler {
	return g.marshaler
}

func (g *GrpcRequestImpl) GetReq() *http.Request {
	return g.req
}

func (g *GrpcRequestImpl) GetPathParams() map[string]string {
	return g.pathParams
}

func (g *GrpcRequestImpl) GetOut() proto.Message {
	return g.out
}
func (g *GrpcRequestImpl) GetIn() proto.Message {
	return g.in
}
func (g *GrpcRequestImpl) GetFullMethodName() string {
	return g.fullMethodName
}

func (g *GrpcRequestImpl) GetCallOptions() []grpc.CallOption {
	return g.callOptions
}

func (g *GrpcRequestImpl) GetOutType() protoreflect.MessageDescriptor {
	return g.outType
}
func (g *GrpcRequestImpl) GetInType() protoreflect.MessageDescriptor {
	return g.inType
}
