package middleware

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	sysRuntime "runtime"
	"strings"

	_ "github.com/begonia-org/begonia/api/v1"
	common "github.com/begonia-org/begonia/common/api/v1"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/routers"
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
	"google.golang.org/protobuf/types/dynamicpb"
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
	return s.SendMsg(m)
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
func isValidContentType(ct string) bool {
	mimeType, _, err := mime.ParseMediaType(ct)
	return err == nil && mimeType != ""
}
func convertDynamicMessageToHttpBody(dynamicMessage *dynamicpb.Message) (*httpbody.HttpBody, error) {
	// 序列化dynamicpb.Message为字节流
	serialized, err := proto.Marshal(dynamicMessage)
	if err != nil {
		return nil, err
	}

	// 反序列化字节流回原始的HttpBody
	var httpBody httpbody.HttpBody
	if err := proto.Unmarshal(serialized, &httpBody); err != nil {
		return nil, err
	}
	// 这里做检查是因为httpbody的字段定义和Any类型的字段定义存在冲突
	// 防止anypb.Any被序列化为httpbody.HttpBody
	// if !isValidContentType(httpBody.ContentType) {
	// 	return nil, fmt.Errorf("invalid content type: %s", httpBody.ContentType)
	// }
	if len(httpBody.Data) == 0 {
		return nil, nil
	}
	return &httpBody, nil
}

func HttpResponseBodyModify(ctx context.Context, w http.ResponseWriter, msg proto.Message) error {
	// 如果消息是dynamicpb.Message类型，那么我们需要将其转换为httpbody.HttpBody类型
	if dynamicMessage, ok := msg.(*dynamicpb.Message); ok {
		if httpBody, err := convertDynamicMessageToHttpBody(dynamicMessage); err == nil && httpBody != nil {
			w.Header().Set("Content-Type", httpBody.ContentType)
			_, err := w.Write(httpBody.Data)
			return err
		}
	}
	// w.Header().Add("Content-Type", "charset=utf-8")
	return nil
}
func HandleErrorWithLogger(logger *logrus.Logger) runtime.ErrorHandlerFunc {
	// codes := getClientMessageMap()
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
			// rspCode := float64(common.Code_INTERNAL_ERROR)
			msg := st.Message()
			details := st.Details()
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
						break
					}
				}

			}
			code = runtime.HTTPStatusFromCode(st.Code())

			log.Errorf(msg)

			w.WriteHeader(code)

			// _, _ = w.Write(data)
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

func UnaryStreamServerErrInterceptor(logger *logrus.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
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
				_ = handler(srv, ss)
			}
		}()
		err = handler(srv, ss)
		if err == nil {
			return err
		}
		return err
	}
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
	// anyData, _ := anypb.New(data)
	// anyData.TypeUrl = ""
	return &common.HttpResponse{
		Code:    int32(common.Code_OK),
		Message: "success",
		Data:    anyData,
	}, err
}
func HttpUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}
	if protocol, ok := md["grpcgateway-content-type"]; ok {
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

func HttpStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	stream := &HttpStream{ServerStream: ss, FullMethod: info.FullMethod}
	return handler(srv, stream)

}
