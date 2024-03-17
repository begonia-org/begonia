package middleware

import (
	"context"
	"fmt"
	"net/http"
	sysRuntime "runtime"

	_ "github.com/begonia-org/begonia/api/v1"
	common "github.com/begonia-org/begonia/common/api/v1"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
)

func HttpResponseModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	// statusCode := http.StatusOK
	// defer func() {
	// 	logger := logger.LoggerFromContext(ctx)
	// 	logger.WithField("status", statusCode).Info("响应结束")
	// }()
	// md, ok := metadata.FromOutgoingContext(ctx)

	// if !ok {
	// 	serverMD, ok := runtime.ServerMetadataFromContext(ctx)
	// 	if !ok {
	// 		return nil
	// 	}
	// 	md = serverMD.HeaderMD
	// }
	// requestID := md.Get("x-request-id")
	// if len(requestID) > 0 {
	// 	w.Header().Set("x-request-id", requestID[0])
	// }
	// w.Header().Del("Grpc-Metadata-Content-Type")
	// w.Header().Del("Grpc-Metadata-X-Request-Id")
	w.Header().Set("Content-Type", "application/json")
	return nil
}

func getClientMessageMap() map[int32]string {
	codes := make(map[int32]string)
	protoregistry.GlobalTypes.RangeEnums(func(desc protoreflect.EnumType) bool {
		values := desc.Descriptor().Values()
		for i := 0; i < values.Len(); i++ {
			v := values.Get(i)
			opts := v.Options()
			if msg := proto.GetExtension(opts, common.E_Msg); msg != nil {
				codes[int32(v.Number())] = msg.(string)
			}
		}
		return true
	})
	return codes

}
func HandleErrorWithLogger(logger *logrus.Logger) runtime.ErrorHandlerFunc {
	codes := getClientMessageMap()
	return func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {
		md, ok := runtime.ServerMetadataFromContext(ctx)
		reqId := ""
		if ok {
			reqIds := md.HeaderMD.Get("x-request-id")
			if len(reqIds) > 0 {
				reqId = reqIds[0]
			}
		}
		if reqId == "" {
			reqId = uuid.New().String()
		}
		uri := req.RequestURI
		method := req.Method
		remoteAddr := req.RemoteAddr
		statusCode := http.StatusOK
		// file, line, fn, _ := errors.GetOneLineSource(err)
		log := logger.WithFields(logrus.Fields{
			"x-request-id": reqId,
			"uri":          uri,
			"method":       method,
			"remote_addr":  remoteAddr,
			"status":       statusCode,
		},
		)
		code := http.StatusOK

		if st, ok := status.FromError(err); ok {
			rspCode := float64(common.Code_INTERNAL_ERROR)
			msg := st.Message()
			details := st.Details()
			for _, detail := range details {
				if anyType, ok := detail.(*anypb.Any); ok {
					var errDetail common.Errors
					if err := anyType.UnmarshalTo(&errDetail); err == nil {
						rspCode = float64(errDetail.Code)
						log = log.WithFields(logrus.Fields{
							"status": int(rspCode),
							"file":   errDetail.File,
							"line":   errDetail.Line,
							"fn":     errDetail.Fn,
						})
						break
					}
				}
			}

			rsp := &common.APIResponse{
				Code:    rspCode,
				Message: codes[int32(rspCode)],
				Data:    nil,
			}
			log.Errorf("request fail:%s", msg)

			data, _ := marshaler.Marshal(rsp)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)

			_, _ = w.Write(data)
			return

		}
		w.WriteHeader(code)
	}
}

func UnaryServerErrInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if p := recover(); p != nil {
				buf := make([]byte, 1024)
				n := sysRuntime.Stack(buf, false) // false 表示不需要所有goroutine的调用栈
				stackTrace := string(buf[:n])

				// 封装错误信息，包括 panic 的信息和调用栈
				// 注意：这里使用 fmt.Errorf 创建了一个基本的 error 实例。
				// 如果你使用的是自定义的错误类型或函数，需要相应地调整
				err = fmt.Errorf("panic: %v\nStack trace: %s", p, stackTrace)
				err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "panic")
			}
		}()
		resp, err = handler(ctx, req)
		if err == nil {
			return resp, err
		}
		return nil, err
	}
}
