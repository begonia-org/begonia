package auth

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type StreamValidator interface {
	ValidateStream(ctx context.Context, req interface{}, fullName string, headers Header) (context.Context, error) 

}
type grpcServerStream struct {
	grpc.ServerStream
	fullName string
	validate StreamValidator
	ctx      context.Context
}

var streamPool = &sync.Pool{
	New: func() interface{} {
		return &grpcServerStream{
			// validate: validator,
		}
	},
}
func NewGrpcStream(s grpc.ServerStream, fullName string, ctx context.Context,validator StreamValidator) *grpcServerStream {
	stream := streamPool.Get().(*grpcServerStream)
	stream.ServerStream = s
	stream.fullName = fullName
	stream.ctx = s.Context()
	stream.validate = validator
	return stream
}
func (g *grpcServerStream) Release() {
	g.ctx = nil
	g.fullName = ""
	g.ServerStream = nil
	g.validate = nil
	streamPool.Put(g)
}
func (g *grpcServerStream) Context() context.Context {
	return g.ctx
}
func (s *grpcServerStream) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	in, ok := metadata.FromIncomingContext(s.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata not exists in context")

	}
	out, ok := metadata.FromOutgoingContext(s.Context())
	if !ok {
		out = metadata.MD{}

	}

	header :=NewGrpcStreamHeader(in, s.Context(), out, s.ServerStream)
	_, err := s.validate.ValidateStream(s.Context(), m, s.fullName, header)
	s.ctx = header.ctx
	header.Release()
	return err
}
