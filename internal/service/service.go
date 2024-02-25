package service

import (
	"context"

	api "github.com/begonia-org/begonia/api/v1"
	common "github.com/begonia-org/begonia/common/api/v1"
	"github.com/google/wire"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Service interface {
	Desc() *grpc.ServiceDesc
}

var ProviderSet = wire.NewSet(NewUserService, NewFileService, NewServices, NewEndpointsService)
var ServiceOptionsSet = wire.NewSet(WithFileService, WithUserService)

type ServiceOptions func(*grpc.Server, *runtime.ServeMux, string) error

func NewServices(file *FileService, user *UsersService, ep *EndpointsService) []Service {
	services := make([]Service, 0)
	services = append(services, file, user, ep)
	return services
}
func WithFileService(file *FileService, opts []grpc.DialOption) ServiceOptions {
	return func(server *grpc.Server, mux *runtime.ServeMux, endpoint string) error {
		common.RegisterFileServiceServer(server, file)
		return common.RegisterFileServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
	}
}
func WithUserService(user *UsersService, opts []grpc.DialOption) ServiceOptions {
	return func(server *grpc.Server, mux *runtime.ServeMux, endpoint string) error {
		api.RegisterAuthServiceServer(server, user)
		return api.RegisterAuthServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
	}
}
