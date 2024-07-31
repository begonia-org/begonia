package gateway

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	gosdk "github.com/begonia-org/go-sdk"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type Exception struct {
	log      logger.Logger
	priority int
	name     string
}

func (e *Exception) setHeader(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	reqId := ""
	if !ok || len(md.Get(XRequestID)) == 0 {
		reqId = uuid.New().String()
		if !ok {
			md = metadata.New(make(map[string]string))
		}
		md.Set(XRequestID, reqId)
		ctx = metadata.NewIncomingContext(ctx, md)

	} else {
		reqId = md.Get(XRequestID)[0]
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(XRequestID, reqId))

	_ = grpc.SetHeader(ctx, metadata.Pairs(XRequestID, reqId))
	return ctx

}
func (e *Exception) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = e.handlePanic(p)
		}

	}()
	ctx = e.setHeader(ctx)
	resp, err = handler(ctx, req)
	if err == nil {
		return resp, err
	}
	return nil, err
}
func (e *Exception) handlePanic(p interface{}) error {
	const maxFrames = 15
	var pcs [maxFrames]uintptr
	n := runtime.Callers(2, pcs[:]) // skip first 3 frames

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("panic: %v\nStack trace:\n", p))
	frames := runtime.CallersFrames(pcs[:n])
	for i := 0; i < maxFrames; i++ {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	err := fmt.Errorf("%s", sb.String())
	err = gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "panic")
	return err
	// _ = ss.SendMsg(err)
}
func (e *Exception) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = e.handlePanic(p)
		}

	}()
	e.setHeader(ss.Context())
	return handler(srv, ss)
}
func (e *Exception) wrapHandlerWithPanicRecovery(handler grpc.StreamHandler) grpc.StreamHandler {
	return func(srv any, stream grpc.ServerStream) (err error) {
		// reqId := ""
		defer func() {
			if p := recover(); p != nil {
				err = e.handlePanic(p)
			}
		}()

		return handler(srv, stream)
	}
}
func (e *Exception) StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	reqId := ""
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok || len(md.Get(XRequestID)) == 0 {
		reqID := uuid.New().String()
		reqId = reqID
		if !ok {
			md = metadata.New(make(map[string]string))
		}
		md.Set(XRequestID, reqId)
		ctx = metadata.NewOutgoingContext(ctx, md)

	} else {
		reqId = md.Get(XRequestID)[0]
	}

	_ = grpc.SetHeader(ctx, metadata.Pairs(XRequestID, reqId))
	desc.Handler = e.wrapHandlerWithPanicRecovery(desc.Handler)
	ss, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err

	}
	return ss, nil
}
func NewException(log logger.Logger) *Exception {
	return &Exception{log: log, name: "exception"}
}

func (e *Exception) Priority() int {
	return e.priority
}
func (e *Exception) SetPriority(priority int) {
	e.priority = priority
}

func (e *Exception) Name() string {
	return e.name
}
