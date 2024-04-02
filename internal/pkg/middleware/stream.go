package middleware

import (
	"context"
	"fmt"
	"sync"

	"github.com/begonia-org/begonia/internal/pkg/errors"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type grpcPluginStream struct {
	grpc.ServerStream
	fullName string
	plugin   gosdk.RemotePlugin
	ctx      context.Context
}

var streamPool = &sync.Pool{
	New: func() interface{} {
		return &grpcPluginStream{
			// validate: validator,
		}
	},
}

func NewGrpcPluginStream(s grpc.ServerStream, fullName string, ctx context.Context, plugin gosdk.RemotePlugin) *grpcPluginStream {
	stream := streamPool.Get().(*grpcPluginStream)
	stream.ServerStream = s
	stream.fullName = fullName
	stream.ctx = s.Context()
	stream.plugin = plugin
	return stream
}
func (g *grpcPluginStream) Release() {
	g.ctx = nil
	g.fullName = ""
	g.ServerStream = nil
	g.plugin = nil
	streamPool.Put(g)
}
func (g *grpcPluginStream) Context() context.Context {
	return g.ctx
}
func (s *grpcPluginStream) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	anyReq, err := anypb.New(m.(protoreflect.ProtoMessage))
	if err != nil {
		return errors.New(fmt.Errorf("new any error: %w", err), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "new_any")

	}
	rsp, err := s.plugin.Call(s.Context(), &api.PluginRequest{
		FullMethodName: s.fullName,
		Request:        anyReq,
	})
	if err != nil {
		return errors.New(fmt.Errorf("call %s plugin error: %w", s.plugin.Name(), err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "call_plugin")

	}
	md, ok := metadata.FromIncomingContext(s.ctx)
	if !ok {
		md = metadata.New(nil)
	}
	for k, v := range rsp.Metadata {
		md.Append(k, v)
	}
	newRequest := rsp.NewRequest
	if newRequest != nil {
		err = newRequest.UnmarshalTo(m.(proto.Message))
		if err != nil {
			return errors.New(fmt.Errorf("unmarshal to request error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "unmarshal_to_request")
		}
	}
	s.ctx = metadata.NewIncomingContext(s.ctx, md)

	// s.ctx = metadata.NewIncomingContext(s.ctx, metadata.New(rsp.Metadata))
	return err
}
