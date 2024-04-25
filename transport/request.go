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
	SetMarshaler(m runtime.Marshaler) error
	SetReq(r *http.Request) error
	SetPathParams(pathParams map[string]string) error
	SetCallOptions(options ...grpc.CallOption)

	GetIn() proto.Message
	GetOut() proto.Message
	GetOutType() protoreflect.MessageDescriptor
	GetInType() protoreflect.MessageDescriptor
	GetFullMethodName() string
	GetCallOptions() []grpc.CallOption
	FullMethodNameMatch(uri string) string
}
type GrpcRequestOptions func(req GrpcRequest) error

func WithGatewayMarshaler(m runtime.Marshaler) GrpcRequestOptions {
	return func(req GrpcRequest) error {
		return req.SetMarshaler(m)
	}
}
func WithGatewayReq(r *http.Request) GrpcRequestOptions {
	return func(req GrpcRequest) error {
		return req.SetReq(r)
	}
}
func WithGatewayPathParams(pathParams map[string]string) GrpcRequestOptions {
	return func(req GrpcRequest) error {
		return req.SetPathParams(pathParams)
	}
}
func WithGatewayCallOptions(options ...grpc.CallOption) GrpcRequestOptions {
	return func(req GrpcRequest) error {
		req.SetCallOptions(options...)
		return nil
	}

}

type GrpcRequestImpl struct {
	Ctx             context.Context
	Marshaler       runtime.Marshaler
	Req             *http.Request
	PathParams      map[string]string
	In              proto.Message
	Out             proto.Message
	InType          protoreflect.MessageDescriptor
	OutType         protoreflect.MessageDescriptor
	FullMethodName  string
	HttpPathPattern map[string]string
	CallOptions     []grpc.CallOption
}

func (g *GrpcRequestImpl) GetContext() context.Context {
	return g.Ctx
}

func (g *GrpcRequestImpl) GetMarshaler() runtime.Marshaler {
	return g.Marshaler
}

func (g *GrpcRequestImpl) GetReq() *http.Request {
	return g.Req
}

func (g *GrpcRequestImpl) GetPathParams() map[string]string {
	return g.PathParams
}
func (g *GrpcRequestImpl) GetProtoReq() proto.Message {
	return g.In
}
func (g *GrpcRequestImpl) GetOut() proto.Message {
	return g.Out
}
func (g *GrpcRequestImpl) GetIn() proto.Message {
	return g.In
}
func (g *GrpcRequestImpl) GetFullMethodName() string {
	return g.FullMethodName
}

func (g *GrpcRequestImpl) SetMarshaler(m runtime.Marshaler) error {
	g.Marshaler = m
	return nil
}
func (g *GrpcRequestImpl) SetReq(r *http.Request) error {
	g.Req = r
	g.FullMethodName = g.FullMethodNameMatch(r.URL.Path)
	return nil
}

func (g *GrpcRequestImpl) SetPathParams(pathParams map[string]string) error {
	g.PathParams = pathParams
	return nil
}
func (g *GrpcRequestImpl) FullMethodNameMatch(uri string) string {
	return g.HttpPathPattern[uri]
}

func (g *GrpcRequestImpl) SetCallOptions(options ...grpc.CallOption) {
	g.CallOptions = options
}
func (g *GrpcRequestImpl) GetCallOptions() []grpc.CallOption {
	return g.CallOptions
}

func (g *GrpcRequestImpl) GetOutType() protoreflect.MessageDescriptor {
	return g.OutType
}
func (g *GrpcRequestImpl) GetInType() protoreflect.MessageDescriptor {
	return g.InType
}
