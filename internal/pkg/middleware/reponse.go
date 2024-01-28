package middleware

import (
	"context"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	common "github.com/wetrycode/begonia/common/api/v1"
	"github.com/wetrycode/begonia/internal/pkg/logger"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func HttpResponseModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	statusCode := http.StatusOK
	defer func() {
		logger := logger.LoggerFromContext(ctx)
		logger.WithField("status", statusCode).Info("响应结束")
	}()
	md, ok := metadata.FromOutgoingContext(ctx)

	if !ok {
		serverMD, ok := runtime.ServerMetadataFromContext(ctx)
		if !ok {
			return nil
		}
		md = serverMD.HeaderMD
	}
	requestID := md.Get("x-request-id")
	if len(requestID) > 0 {
		w.Header().Set("x-request-id", requestID[0])
	}
	w.Header().Del("Grpc-Metadata-Content-Type")
	w.Header().Del("Grpc-Metadata-X-Request-Id")
	return nil
}

func HandleErrorWithLogger(logger *logrus.Logger) runtime.ErrorHandlerFunc {
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
		file, line, fn, _ := errors.GetOneLineSource(err)
		log := logger.WithFields(logrus.Fields{
			"x-request-id": reqId,
			"uri":          uri,
			"method":       method,
			"remote_addr":  remoteAddr,
			"status":       statusCode,
			"file":         file,
			"line":         line,
			"func":         fn,
		},
		)
		code := http.StatusInternalServerError
		if st, ok := status.FromError(err); ok {
			code = runtime.HTTPStatusFromCode(st.Code())
			log.WithField("status", http.StatusOK).Errorf("请求失败:%s", errors.Cause(err).Error())

			for _, detail := range st.Details() {
				switch t := detail.(type) {
				case *anypb.Any:
					var apiResp common.APIResponse
					if err := anypb.UnmarshalTo(t, &apiResp, proto.UnmarshalOptions{}); err == nil {
						data, err := marshaler.Marshal(&apiResp)
						if err != nil {
							w.WriteHeader(500)
							return
						}
						_, _ = w.Write(data)

						w.WriteHeader(200)
					}
				}
			}

		}
		w.WriteHeader(code)
	}
}
