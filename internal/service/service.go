package service

import (
	"context"

	"github.com/google/wire"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	api "github.com/wetrycode/begonia/api/v1"
	common "github.com/wetrycode/begonia/common/api/v1"
	"google.golang.org/grpc"
)

var ProviderSet = wire.NewSet(NewUserService, NewFileService)
var ServiceOptionsSet = wire.NewSet(WithFileService, WithUserService)

type ServiceOptions func(*grpc.Server, *runtime.ServeMux, string) error

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
