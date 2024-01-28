package server

import (
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"github.com/wetrycode/begonia/internal/pkg/logger"
	"github.com/wetrycode/begonia/internal/pkg/middleware"
)

var GlobalMutex sync.RWMutex
var GlobalMux *runtime.ServeMux
var muxNewOnce sync.Once

func NewGatewayMux(config *config.Config) *runtime.ServeMux {
	muxNewOnce.Do(func() {
		GlobalMux = runtime.NewServeMux(
			runtime.WithMarshalerOption("application/octet-stream", middleware.NewRawBinaryUnmarshaler()),
			runtime.WithMarshalerOption("application/json", middleware.NewResponseJSONMarshaler()),
			runtime.WithMarshalerOption("text/event-stream", middleware.NewEventSourceMarshaler()),
			runtime.WithErrorHandler(middleware.HandleErrorWithLogger(logger.Logger)),
			runtime.WithMetadata(middleware.IncomingHeadersToMetadata),
		// runtime.WithForwardResponseOption(middleware.HttpResponseModifier),
		)
		// ctx := context.Background()
		// err := initGrpcSvr(ctx, GlobalMux, file, config)
		// if err != nil {
		// 	panic(err)
		// }
	})

	return GlobalMux
}
