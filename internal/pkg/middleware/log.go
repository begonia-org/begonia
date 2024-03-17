package middleware

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type LoggerMiddleware struct {
	*logrus.Logger
}

func NewLoggerMiddleware(log *logrus.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{Logger: log}
}
func (log *LoggerMiddleware) logger(ctx context.Context, fullMethod string, err error, elapsed time.Duration) {
	md, _ := metadata.FromIncomingContext(ctx)
	reqId := md.Get("x-request-id")[0]
	remoteAddr := ""
	if len(md.Get("x-forwarded-for")) > 0 {
		remoteAddr = md.Get("x-forwarded-for")[0]
	}
	if len(md.Get("remote_addr")) > 0 {
		remoteAddr = md.Get("remote_addr")[0]
	}
	method := "Unkonwn"
	uri := "Unkonwn"
	xuid := "Unkonwn"
	if len(md.Get("method")) > 0 {
		method = md.Get("method")[0]
	}
	if len(md.Get("uri")) > 0 {
		uri = md.Get("uri")[0]
	}
	if len(md.Get("x-uid")) > 0 {
		xuid = md.Get("x-uid")[0]
	}
	code := status.Code(err)
	logger := log.WithFields(logrus.Fields{
		"x-request-id": reqId,
		"uri":          uri,
		"method":       method,
		"remote_addr":  remoteAddr,
		"status":       code,
		"x-uid":        xuid,
		"name":         fullMethod,
		"elapsed":      elapsed.String(),
	})
	if err != nil {
		logger.Info(err)
	} else {
		logger.Info("success")
	}

}
func (log *LoggerMiddleware) LoggerUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	reqId := uuid.New().String()
	if !ok {
		md = metadata.New(map[string]string{"x-request-id": reqId})
	} else if _, exists := md["x-request-id"]; !exists {
		// 如果metadata存在但没有x-request-id，那么添加它
		md = metadata.Join(md, metadata.New(map[string]string{"x-request-id": reqId}))
	}
	ctx = metadata.NewIncomingContext(ctx, md)
	now := time.Now()
	// log.Info("start request")
	rsp, err := handler(ctx, req)
	elapsed := time.Since(now)
	log.logger(ctx, info.FullMethod, err, elapsed)
	return rsp, err
}
func (log *LoggerMiddleware) LoggerStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(ss.Context())
	reqId := uuid.New().String()
	ctx := ss.Context()
	var err error
	if !ok {
		md = metadata.New(map[string]string{"x-request-id": reqId})
	} else if _, exists := md["x-request-id"]; !exists {
		// 如果metadata存在但没有x-request-id，那么添加它
		md = metadata.Join(md, metadata.New(map[string]string{"x-request-id": reqId}))
	}
	ctx = metadata.NewIncomingContext(ctx, md)
	now := time.Now()
	err = handler(srv, ss)
	elapsed := time.Since(now)

	log.logger(ctx, info.FullMethod, err, elapsed)
	return err
}
