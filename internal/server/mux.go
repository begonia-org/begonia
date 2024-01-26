package server

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"github.com/wetrycode/begonia/internal/pkg/logger"
	middleware "github.com/wetrycode/begonia/internal/pkg/middlerware"
	"github.com/wetrycode/begonia/internal/service"
	"google.golang.org/protobuf/encoding/protojson"
)

func NewGatewayMux(user *service.UsersService, config *config.Config) *runtime.ServeMux {
	mux := runtime.NewServeMux(runtime.WithMarshalerOption("application/json", &middleware.ResponseJSONMarshaler{JSONPb: runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true, // 设置为 true 以确保默认值（例如 0 或空字符串）被序列化
			UseEnumNumbers:  true, // 设置为 true 以确保枚举值被序列化为数字而不是字符串
			UseProtoNames:   true, // 设置为 true 以确保 proto 消息的原始名称（而不是 Go 字段名称）被序列化

		},
	}}),
		runtime.WithMarshalerOption("text/event-stream", &middleware.EventSourceMarshaler{ResponseJSONMarshaler: middleware.ResponseJSONMarshaler{
			JSONPb: runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					EmitUnpopulated: true, // 设置为 true 以确保默认值（例如 0 或空字符串）被序列化
					UseEnumNumbers:  true, // 设置为 true 以确保枚举值被序列化为数字而不是字符串
					UseProtoNames:   true, // 设置为 true 以确保 proto 消息的原始名称（而不是 Go 字段名称）被序列化

				},
			}}}),
		runtime.WithErrorHandler(middleware.HandleErrorWithLogger(logger.Logger)),
		runtime.WithMetadata(middleware.IncomingHeadersToMetadata),
		runtime.WithForwardResponseOption(middleware.HttpResponseModifier),
	)
	ctx := context.Background()
	err := initGrpcSvr(ctx, mux, user, config)
	if err != nil {
		panic(err)
	}
	return mux
}
