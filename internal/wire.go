//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/daemon"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/migrate"
	"github.com/begonia-org/begonia/internal/server"
	"github.com/begonia-org/begonia/internal/service"

	"github.com/begonia-org/begonia/internal/pkg/logger"
	"github.com/google/wire"

	"github.com/spark-lence/tiga"
)

func InitOperatorApp(config *tiga.Configuration) *migrate.InitOperator {

	panic(wire.Build(data.ProviderSet, pkg.ProviderSet))

}

func New(config *tiga.Configuration, log logger.Logger, endpoint string) GatewayWorker {

	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, service.ProviderSet, daemon.ProviderSet, server.ProviderSet, NewGatewayWorkerImpl))

}
