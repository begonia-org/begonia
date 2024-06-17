package middleware

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg/routers"
	gosdk "github.com/begonia-org/go-sdk"
	_ "github.com/begonia-org/go-sdk/api/app/v1"
	_ "github.com/begonia-org/go-sdk/api/endpoint/v1"
	_ "github.com/begonia-org/go-sdk/api/file/v1"
	_ "github.com/begonia-org/go-sdk/api/iam/v1"
	_ "github.com/begonia-org/go-sdk/api/plugin/v1"
	_ "github.com/begonia-org/go-sdk/api/sys/v1"
	_ "github.com/begonia-org/go-sdk/api/user/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"

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
	if protocol, ok := md["grpcgateway-accept"]; ok {
		if !strings.EqualFold(protocol[0], "application/json") {
			return s.ServerStream.SendMsg(m)
		}
		routersList := routers.Get()
		router := routersList.GetRouteByGrpcMethod(s.FullMethod)
		// 对内置服务的http响应进行格式化
		if routersList.IsLocalSrv(s.FullMethod) || router.UseJsonResponse {
			log.Printf("转换fullMethod:%v", s.FullMethod)
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
						msg := codesMap[rspCode]
						if errDetail.ToClientMessage != "" {
							msg = errDetail.ToClientMessage
						}
						return &common.HttpResponse{
							Code:    rspCode,
							Message: msg,
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
		}, err

	}
	anyData := &structpb.Struct{}
	if rsp != nil {
		data := rsp.(protoreflect.ProtoMessage)
		anyData, err = toStructMessage(data)
		if err != nil {
			return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "internal_error")
		}
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
		router := routersList.GetRouteByGrpcMethod(info.FullMethod)
		// 对内置服务的http响应进行格式化
		if routersList.IsLocalSrv(info.FullMethod) || router.UseJsonResponse {
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
