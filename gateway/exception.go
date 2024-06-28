package gateway

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	gosdk "github.com/begonia-org/go-sdk"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/begonia-org/go-sdk/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Exception struct {
	log      logger.Logger
	priority int
	name     string
}

func (e *Exception) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if p := recover(); p != nil {
			// buf := make([]byte, 1024)
			// n := sysRuntime.Stack(buf, false) // false 表示不需要所有goroutine的调用栈
			// stackTrace := string(buf[:n])
			// err = fmt.Errorf("panic: %v\nStack trace: %s", p, stackTrace)
			// err = gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "panic")
			err = e.handlePanic(p)
		}
	}()
	resp, err = handler(ctx, req)
	if err == nil {
		return resp, err
	}
	return nil, err
}
func (e *Exception) handlePanic(p interface{}) error {
	const maxFrames = 10
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
			// buf := make([]byte, 512)
			// n := sysRuntime.Stack(buf, false) // false 表示不需要所有goroutine的调用栈
			// stackTrace := string(buf[:n])

			// err = fmt.Errorf("panic: %v\nStack trace: %s", p, stackTrace)
			// err = gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "panic")
			// _ = ss.SendMsg(err)
			err = e.handlePanic(p)
		}
	}()
	return handler(srv, ss)
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
