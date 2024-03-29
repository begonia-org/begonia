package middleware

import (
	"context"
	"time"

	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

type LoggerMiddleware struct {
	*logrus.Logger
	priority int
	name string
}

func NewLoggerMiddleware(log *logrus.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{Logger: log, name: "logger"}
}
func (log *LoggerMiddleware) Priority() int {
	return log.priority
}
func (log *LoggerMiddleware) SetPriority(priority int) {
	log.priority = priority

}
func (log *LoggerMiddleware) Name() string {
	return log.name
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
		if st, ok := status.FromError(err); ok {
			details := st.Details()
			for _, detail := range details {
				if anyType, ok := detail.(*anypb.Any); ok {
					var errDetail common.Errors
					if err := anyType.UnmarshalTo(&errDetail); err == nil {
						rspCode := float64(errDetail.Code)
						logger = logger.WithFields(logrus.Fields{
							"status": int(rspCode),
							"file":   errDetail.File,
							"line":   errDetail.Line,
							"fn":     errDetail.Fn,
						})
						break
					}
				}
			}

		}
		logger.Error(err.Error())
	} else {
		logger.Info("success")
	}

}
func (log *LoggerMiddleware) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (rsp interface{}, err error) {
	now := time.Now()
	defer func() {
		if r := recover(); r != nil {
			elapsed := time.Since(now)
			log.logger(ctx, info.FullMethod, err, elapsed)
		}
	}()
	md, ok := metadata.FromIncomingContext(ctx)
	reqId := uuid.New().String()
	if !ok {
		md = metadata.New(map[string]string{"x-request-id": reqId})
	} else if _, exists := md["x-request-id"]; !exists {
		// 如果metadata存在但没有x-request-id，那么添加它
		md = metadata.Join(md, metadata.New(map[string]string{"x-request-id": reqId}))
	}
	ctx = metadata.NewIncomingContext(ctx, md)

	// log.Info("start request")
	rsp, err = handler(ctx, req)
	elapsed := time.Since(now)
	log.logger(ctx, info.FullMethod, err, elapsed)
	return
}
func (log *LoggerMiddleware) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	now := time.Now()
	defer func() {
		if r := recover(); r != nil {
			elapsed := time.Since(now)
			log.logger(ss.Context(), info.FullMethod, err, elapsed)
		}
	}()
	md, ok := metadata.FromIncomingContext(ss.Context())
	reqId := uuid.New().String()
	ctx := ss.Context()
	if !ok {
		md = metadata.New(map[string]string{"x-request-id": reqId})
	} else if _, exists := md["x-request-id"]; !exists {
		// 如果metadata存在但没有x-request-id，那么添加它
		md = metadata.Join(md, metadata.New(map[string]string{"x-request-id": reqId}))
	}
	ctx = metadata.NewIncomingContext(ctx, md)
	err = handler(srv, ss)
	elapsed := time.Since(now)

	log.logger(ctx, info.FullMethod, err, elapsed)
	return err
}
