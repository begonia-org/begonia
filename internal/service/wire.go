//go:build wireinject
// +build wireinject

package service

import (
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg"


	"github.com/begonia-org/go-sdk/logger"
	app "github.com/begonia-org/go-sdk/api/app/v1"
	ep "github.com/begonia-org/go-sdk/api/endpoint/v1"
	file "github.com/begonia-org/go-sdk/api/file/v1"
	sys "github.com/begonia-org/go-sdk/api/sys/v1"
	user "github.com/begonia-org/go-sdk/api/user/v1"

	"github.com/google/wire"

	"github.com/spark-lence/tiga"
)



func NewAuthzSvrForTest(config *tiga.Configuration, log logger.Logger)user.AuthServiceServer {
	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, NewAuthzService))
}
func NewAPPSvrForTest(config *tiga.Configuration, log logger.Logger)app.AppsServiceServer {
	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, NewAppService))
}
func NewEndpointSvrForTest(config *tiga.Configuration, log logger.Logger)ep.EndpointServiceServer {
	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, NewEndpointsService))
}
func NewFileSvrForTest(config *tiga.Configuration, log logger.Logger)file.FileServiceServer {
	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, NewFileService))
}
func NewSysSvrForTest(config *tiga.Configuration, log logger.Logger)sys.SystemServiceServer {
	panic(wire.Build(NewSysService))
}
func NewUserSvrForTest(config *tiga.Configuration, log logger.Logger)user.UserServiceServer {
	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, NewUserService))
}