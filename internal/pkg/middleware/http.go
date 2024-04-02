package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	gosdk "github.com/begonia-org/go-sdk"
	_ "github.com/begonia-org/go-sdk/api/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/api/httpbody"
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

// func HttpResponseModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {

//		w.Header().Set("Content-Type", "application/json")
//		return nil
//	}
type HttpStream struct {
	grpc.ServerStream
	FullMethod string
}
type Http struct {
	priority int
	name     string
}

func (s *HttpStream) SendMsg(m interface{}) error {
	ctx := s.ServerStream.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return s.ServerStream.SendMsg(m)
	}
	if protocol, ok := md["grpcgateway-content-type"]; ok {
		if !strings.EqualFold(protocol[0], "application/json") {
			return s.ServerStream.SendMsg(m)
		}
		routersList := routers.Get()
		// 对内置服务的http响应进行格式化
		if routersList.IsLocalSrv(s.FullMethod) {

			rsp, _ := grpcToHttpResponse(m, nil)
			return s.ServerStream.SendMsg(rsp)
		}
	}
	return s.ServerStream.SendMsg(m)
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

//	func isValidContentType(ct string) bool {
//		mimeType, _, err := mime.ParseMediaType(ct)
//		return err == nil && mimeType != ""
//	}

func writeHttpHeaders(w http.ResponseWriter, key string, value []string) {
	if httpKey := gosdk.GetHttpHeaderKey(key); httpKey != "" {
		for _, v := range value {
			w.Header().Del(key)
			if v != "" {
				if strings.EqualFold(httpKey, "Content-Type") {
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
				return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "internal_error")
			}
			httpCode = int(code)

		}

	}

	out, ok := metadata.FromOutgoingContext(ctx)
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
		data := &common.HttpResponse{}
		data.Code = int32(common.Code_INTERNAL_ERROR)
		data.Message = "internal error"
		if st, ok := status.FromError(err); ok {
			// rspCode := float64(common.Code_INTERNAL_ERROR)
			msg := st.Message()
			details := st.Details()
			data.Message = clientMessageFromCode(st.Code())

			// code = runtime.HTTPStatusFromCode(st.Code())
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
						data.Code = errDetail.Code
						data.Message = codes[int32(errDetail.Code)]
						data.Data = &structpb.Struct{}
						break
					}
				} else {
					log.Errorf("error type:%T, error:%v", err, err)

				}

			}
			code = runtime.HTTPStatusFromCode(st.Code())

			log.Errorf(msg)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			bData, _ := protojson.Marshal(data)
			_, _ = w.Write(bData)
			return

		} else {
			if err != nil {
				log.Errorf("error type:%T, error:%v", err, err)
			}
		}
		w.WriteHeader(code)
	}
}

func OutgoingMetaInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if reqIds, ok := md["x-request-id"]; ok {
			ctx = metadata.AppendToOutgoingContext(ctx, gosdk.GetMetadataKey("x-request-id"), reqIds[0])
		}
	}
	resp, err = handler(ctx, req)
	if err == nil {
		return resp, err
	}
	return nil, err
}

func toStructMessage(msg protoreflect.ProtoMessage) (*structpb.Struct, error) {
	jsonBytes, err := protojson.Marshal(msg)

	if err != nil {
		return nil, fmt.Errorf("Failed to serialize message to JSON: %w", err)
	}

	// 将 JSON 字符串解析为 structpb.Struct
	structMsg := &structpb.Struct{}
	if err := protojson.Unmarshal(jsonBytes, structMsg); err != nil {
		return nil, fmt.Errorf("Failed to parse JSON into structpb.Struct: %w", err)
	}
	return structMsg, nil
}
func grpcToHttpResponse(rsp interface{}, err error) (*common.HttpResponse, error) {
	if err != nil {
		if st, ok := status.FromError(err); ok {
			details := st.Details()
			for _, detail := range details {
				if anyType, ok := detail.(*anypb.Any); ok {
					var errDetail common.Errors
					var stErr = anyType.UnmarshalTo(&errDetail)

					if stErr == nil {
						rspCode := int32(errDetail.Code)
						codesMap := getClientMessageMap()
						return &common.HttpResponse{
							Code:    rspCode,
							Message: codesMap[int32(rspCode)],
							Data:    nil,
						}, err
					}
				}
			}
			code := int32(common.Code_INTERNAL_ERROR)
			if st.Code() == codes.NotFound {
				code = int32(common.Code_NOT_FOUND)
			}
			if st.Code() == codes.Unimplemented {
				code = int32(common.Code_NOT_FOUND)
			}

			return &common.HttpResponse{
				Code:    code,
				Message: st.Message(),
				Data:    nil,
			}, err

		}
		return &common.HttpResponse{
			Code:    int32(common.Code_INTERNAL_ERROR),
			Data:    nil,
			Message: "internal error",
		}, nil

	}
	data := rsp.(protoreflect.ProtoMessage)
	anyData, err := toStructMessage(data)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "internal_error")
	}
	return &common.HttpResponse{
		Code:    int32(common.Code_OK),
		Message: "success",
		Data:    anyData,
	}, err
}
func (h *Http) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}
	if protocol, ok := md["grpcgateway-accept"]; ok {
		if !strings.EqualFold(protocol[0], "application/json") {
			return handler(ctx, req)
		}
		routersList := routers.Get()
		// 对内置服务的http响应进行格式化
		if routersList.IsLocalSrv(info.FullMethod) {
			rsp, err := handler(ctx, req)
			if _, ok := rsp.(*httpbody.HttpBody); ok {
				return rsp, err
			}
			return grpcToHttpResponse(rsp, err)
		}
	}

	return handler(ctx, req)
}

func (h *Http) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	stream := &HttpStream{ServerStream: ss, FullMethod: info.FullMethod}
	return handler(srv, stream)

}

func NewHttp() *Http {
	return &Http{name: "http"}
}

func (h *Http) Priority() int {
	return h.priority
}
func (h *Http) SetPriority(priority int) {
	h.priority = priority
}
func (h *Http) Name() string {
	return h.name
}
