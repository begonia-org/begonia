package service

import (
	"context"

	gosdk "github.com/begonia-org/go-sdk"
	app "github.com/begonia-org/go-sdk/api/app/v1"
	ep "github.com/begonia-org/go-sdk/api/endpoint/v1"
	file "github.com/begonia-org/go-sdk/api/file/v1"
	sys "github.com/begonia-org/go-sdk/api/sys/v1"
	user "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/google/wire"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Service interface {
	Desc() *grpc.ServiceDesc
}

var ProviderSet = wire.NewSet(NewAuthzService, NewUserService,
	NewFileService,
	NewServices,
	NewEndpointsService,
	NewAppService,
	NewSysService)

type ServiceOptions func(*grpc.Server, *runtime.ServeMux, string) error

func NewServices(file file.FileServiceServer,
	authz user.AuthServiceServer,
	ep ep.EndpointServiceServer,
	app app.AppsServiceServer,
	sys sys.SystemServiceServer,
	users user.UserServiceServer,

) []Service {
	services := make([]Service, 0)
	services = append(services, file.(Service), authz.(Service), ep.(Service), app.(Service), sys.(Service), users.(Service))
	return services
}

func GetIdentity(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	identity := md.Get(gosdk.HeaderXIdentity)
	if len(identity) > 0 {
		return identity[0]
	}
	return ""
}
