package gateway

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	gosdk "github.com/begonia-org/go-sdk"
	_ "github.com/begonia-org/go-sdk/api"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	XRequestID  = "x-request-id"
	XUID        = "x-uid"
	XAccessKey  = "x-access-key"
	XHttpMethod = "x-http-method"
	XRemoteAddr = "x-http-forwarded-for"
	XProtocol   = "x-http-protocol"
	XHttpURI    = "x-http-uri"
	XIdentity   = "x-identity"
)

func preflightHandler(w http.ResponseWriter, _ *http.Request) {
	// headers := []string{"Content-Type", "Accept", "Authorization", "X-Token", "x-date", "x-access-key"}
	w.Header().Set("Access-Control-Allow-Headers", "*")
	// methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.WriteHeader(http.StatusNoContent)
}

type CorsHandler struct {
	Cors []string
}

func (cors *CorsHandler) Handle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if clientOrigin := r.Header.Get("Origin"); clientOrigin != "" {
			var isAllowed bool

			for _, origin := range cors.Cors {
				if origin == "*" || strings.HasSuffix(clientOrigin, origin) {
					isAllowed = true
					break
				}
			}
			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", clientOrigin)
				if r.Method == "OPTIONS" {
					preflightHandler(w, r)
					return
				}
			} else {
				Log.Errorf(r.Context(), "origin:%s not allowed", clientOrigin)
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func IncomingHeadersToMetadata(ctx context.Context, req *http.Request) metadata.MD {
	// 创建一个新的 metadata.MD 实例

	md := metadata.MD{}
	invalidHeaders := []string{
		"Connection", "Keep-Alive", "Proxy-Connection",
		"Transfer-Encoding", "Upgrade", "TE",
	}
	headers := req.Header.Clone()

	for _, h := range invalidHeaders {
		headers.Del(h)

	}
	for k, v := range headers {
		if strings.HasPrefix(strings.ToLower(k), "sec-") {
			continue
		}
		if strings.ToLower(k) == "pragma" {
			continue
		}

		md.Set(strings.ToLower(k), v...)
	}
	// 设置一些默认的元数据
	reqID := uuid.New().String()
	md.Set(XRequestID, reqID)
	md.Set("uri", req.RequestURI)
	md.Set(XHttpURI, req.RequestURI)
	md.Set(XHttpMethod, req.Method)
	md.Set("remote_addr", req.RemoteAddr)
	md.Set(XRemoteAddr, req.RemoteAddr)
	md.Set("protocol", req.Proto)
	md.Set(XProtocol, req.Proto)
	md.Set(gosdk.GetMetadataKey(XRequestID), reqID)
	xuid := md.Get(XUID)
	accessKey := md.Get(XAccessKey)
	author := ""
	if len(xuid) > 0 {
		author = xuid[0]
	}
	if author == "" && len(accessKey) > 0 {
		author = accessKey[0]
	}
	if author == "" {
		return md
	}
	md.Set(XIdentity, author)

	return md
}

type LoggerMiddleware struct {
	log      logger.Logger
	priority int
	name     string
}

func NewLoggerMiddleware(log logger.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{log: log, name: "logger"}
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
	code := status.Code(err)
	httpCode := runtime.HTTPStatusFromCode(code)
	logger := log.log.WithFields(logrus.Fields{
		"status":  httpCode,
		"elapsed": elapsed.String(),
		"name":    fullMethod,
		"module":  "request",
	})
	if err != nil {
		logger.Error(ctx, err)
	} else {
		logger.Info(ctx, "success")
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

	ctx := ss.Context()

	err = handler(srv, ss)
	elapsed := time.Since(now)

	log.logger(ctx, info.FullMethod, err, elapsed)
	return err
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

func clientMessageFromCode(code codes.Code) string {
	switch code {
	case codes.ResourceExhausted:
		return "The requested resource size exceeds the server limit."
	default:
		return "Unknown error"

	}
}

func HandleErrorWithLogger(logger logger.Logger) runtime.ErrorHandlerFunc {
	codes := getClientMessageMap()
	return func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {

		statusCode := http.StatusOK

		log := logger.WithFields(logrus.Fields{

			"status": statusCode,
		},
		)
		code := http.StatusOK
		data := &common.HttpResponse{}
		data.Code = int32(common.Code_INTERNAL_ERROR)
		data.Message = "internal error"
		if st, ok := status.FromError(err); ok {
			msg := st.Message()
			details := st.Details()
			data.Message = clientMessageFromCode(st.Code())

			for _, detail := range details {
				if anyType, ok := detail.(*anypb.Any); ok {
					var errDetail common.Errors
					if err := anyType.UnmarshalTo(&errDetail); err == nil {
						rspCode := float64(errDetail.Code)
						log = log.WithFields(logrus.Fields{
							"status": int(rspCode),
							"file":   errDetail.File,
							"line":   errDetail.Line,
							"fn":     errDetail.Fn,
						})

						msg := codes[int32(errDetail.Code)]
						if errDetail.ToClientMessage != "" {
							msg = errDetail.ToClientMessage
						}

						data.Code = errDetail.Code
						data.Message = msg
						data.Data = &structpb.Struct{}
						break
					}
				} else {
					log.Errorf(ctx, "error type:%T, error:%v", err, err)

				}

			}
			code = runtime.HTTPStatusFromCode(st.Code())

			log.WithField("status", code).Errorf(ctx, msg)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			bData, _ := protojson.Marshal(data)
			_, _ = w.Write(bData)
			return

		} else {
			if err != nil {
				log.Errorf(ctx, "error type:%T, error:%v", err, err)
			}
		}
		w.WriteHeader(code)
	}
}
func writeHttpHeaders(w http.ResponseWriter, key string, value []string) {
	if httpKey := gosdk.GetHttpHeaderKey(key); httpKey != "" {
		for _, v := range value {
			w.Header().Del(key)
			if v != "" {
				if strings.EqualFold(httpKey, "Content-Type") {
					if v == "application/grpc" {
						continue
					}
					w.Header().Set(httpKey, v)
				} else {
					w.Header().Add(httpKey, v)

				}

				headers := w.Header().Values("Access-Control-Expose-Headers")
				for _, h := range headers {
					if strings.EqualFold(h, httpKey) {
						return
					}
				}
				headers = append(headers, http.CanonicalHeaderKey(httpKey))
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(headers, ","))

			}
		}

	}

}
func HttpResponseBodyModify(ctx context.Context, w http.ResponseWriter, msg proto.Message) error {
	httpCode := http.StatusOK
	for key, value := range w.Header() {
		if strings.HasPrefix(key, "Grpc-Metadata-") {
			w.Header().Del(key)

		}
		writeHttpHeaders(w, key, value)
		if strings.HasSuffix(http.CanonicalHeaderKey(key), "X-Http-Code") {
			codeStr := value[0]
			code, err := strconv.ParseInt(codeStr, 10, 32)
			if err != nil {
				Log.Error(ctx, err)
				return status.Error(codes.Internal, err.Error())
			}
			httpCode = int(code)

		}

	}

	out, ok := metadata.FromIncomingContext(ctx)
	if ok {
		for k, v := range out {
			writeHttpHeaders(w, k, v)
		}
	}
	if httpCode != http.StatusOK {
		w.WriteHeader(httpCode)
	}
	return nil
}
