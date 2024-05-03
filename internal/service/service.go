package service

import (
	"context"

	api "github.com/begonia-org/go-sdk/api/file/v1"
	userAPI "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/google/wire"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Service interface {
	Desc() *grpc.ServiceDesc
}

var ProviderSet = wire.NewSet(NewUserService, NewFileService, NewServices, NewEndpointsService, NewAppService, NewSysService)
var ServiceOptionsSet = wire.NewSet(WithFileService, WithUserService)

type ServiceOptions func(*grpc.Server, *runtime.ServeMux, string) error

func NewServices(file *FileService, user *UsersService, ep *EndpointsService, app *AppService, sys *SysService) []Service {
	services := make([]Service, 0)
	services = append(services, file, user, ep, app, sys)
	return services
}
func WithFileService(file *FileService, opts []grpc.DialOption) ServiceOptions {
	return func(server *grpc.Server, mux *runtime.ServeMux, endpoint string) error {
		api.RegisterFileServiceServer(server, file)
		return api.RegisterFileServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
	}
}
func WithUserService(user *UsersService, opts []grpc.DialOption) ServiceOptions {
	return func(server *grpc.Server, mux *runtime.ServeMux, endpoint string) error {
		userAPI.RegisterAuthServiceServer(server, user)
		return userAPI.RegisterAuthServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
	}
}

func GetIdentity(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	identity := md.Get("x-identity")
	if len(identity) > 0 {
		return identity[0]
	}
	return ""
}
