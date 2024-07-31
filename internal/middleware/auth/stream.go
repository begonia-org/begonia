package auth

import (
	"context"
	"errors"
	"io"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StreamValidator interface {
	ValidateStream(ctx context.Context, req interface{}, fullName string) (context.Context, error)
}
type grpcServerStream struct {
	grpc.ServerStream
	fullName   string
	validate   StreamValidator
	ctx        context.Context
	firstFrame bool
}

// type grpcClientStream struct {
// 	grpc.ClientStream
// 	fullName   string
// 	ctx        context.Context
// 	validate   StreamValidator
// 	firstFrame bool
// }

var streamPool = &sync.Pool{
	New: func() interface{} {
		return &grpcServerStream{
			// validate: validator,
		}
	},
}

// var clientStreamPool = &sync.Pool{
// 	New: func() interface{} {
// 		return &grpcClientStream{}
// 	},
// }

func NewGrpcStream(s grpc.ServerStream, fullName string, ctx context.Context, validator StreamValidator) *grpcServerStream {
	stream := streamPool.Get().(*grpcServerStream)
	stream.ServerStream = s
	stream.fullName = fullName
	stream.ctx = ctx
	stream.validate = validator
	return stream
}
func (g *grpcServerStream) Release() {
	g.ctx = nil
	g.fullName = ""
	g.ServerStream = nil
	g.validate = nil
	g.firstFrame = false
	streamPool.Put(g)
}
func (g *grpcServerStream) Context() context.Context {
	return g.ctx
}
func (s *grpcServerStream) RecvMsg(m interface{}) error {
	var err error
	if err = s.ServerStream.RecvMsg(m); err != nil && !errors.Is(err, io.EOF) {
		return status.Errorf(codes.Internal, "recv msg err:%s", err.Error())
	}
	if err != nil {
		return err

	}
	if !s.firstFrame {

		ctx, err := s.validate.ValidateStream(s.Context(), m, s.fullName)
		s.ctx = ctx
		s.firstFrame = true
		return err
	}

	return nil
}
