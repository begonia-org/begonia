//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/daemon"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/middleware"
	"github.com/begonia-org/begonia/internal/migrate"
	"github.com/begonia-org/begonia/internal/pkg"

	"github.com/begonia-org/begonia/internal/server"
	"github.com/begonia-org/begonia/internal/service"
	app "github.com/begonia-org/go-sdk/api/app/v1"
	ep "github.com/begonia-org/go-sdk/api/endpoint/v1"
	file "github.com/begonia-org/go-sdk/api/file/v1"
	sys "github.com/begonia-org/go-sdk/api/sys/v1"
	user "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"

	"github.com/google/wire"

	"github.com/spark-lence/tiga"
)

func InitOperatorApp(config *tiga.Configuration) *migrate.InitOperator {

	panic(wire.Build(data.ProviderSet, pkg.ProviderSet, migrate.ProviderSet))

}

func New(config *tiga.Configuration, log logger.Logger, endpoint string) GatewayWorker {

	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, service.ProviderSet, daemon.ProviderSet, server.ProviderSet, middleware.ProviderSet, NewGatewayWorkerImpl))

}

func NewAuthzSvr(config *tiga.Configuration, log logger.Logger) user.AuthServiceServer {
	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, service.ProviderSet))
}
func NewAPPSvr(config *tiga.Configuration, log logger.Logger) app.AppsServiceServer {
	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, service.ProviderSet))
}
func NewEndpointSvr(config *tiga.Configuration, log logger.Logger) ep.EndpointServiceServer {
	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, service.ProviderSet))
}
func NewFileSvr(config *tiga.Configuration, log logger.Logger) file.FileServiceServer {
	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, service.ProviderSet))
}
func NewSysSvr(config *tiga.Configuration, log logger.Logger) sys.SystemServiceServer {
	panic(wire.Build( service.ProviderSet))
}
