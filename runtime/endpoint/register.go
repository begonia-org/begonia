package endpoint

import (
	"context"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type EndpointRegister interface {
	RegisterService(serviceName string, ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
	GetAllEndpoints() map[string]EndpointRegisterFunc
}
type EndpointRegisterFunc func(serviceName string, ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
type Endpoints map[string]EndpointRegisterFunc

var globalMutex sync.RWMutex

var GlobalEndpoints = make(Endpoints)

func AddEndpoint(name string, registerFunc EndpointRegisterFunc) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	GlobalEndpoints[name] = registerFunc
}

// type Endpoint interface{
// 	RegisterHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
// 	RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
// }
