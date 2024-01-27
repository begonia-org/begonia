package endpoint

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type EndpointRegister interface {
	RegisterService(serviceName string, ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
	GetAllEndpoints() map[string]EndpointRegisterFunc
	RegisterAll(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
}
type EndpointRegisterFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
type Endpoints map[string]EndpointRegisterFunc

// var globalMutex sync.RWMutex

var GlobalEndpoints = make(Endpoints)

//	type Endpoint interface{
//		RegisterHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
//		RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
//	}
