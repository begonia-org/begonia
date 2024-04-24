//go:build wireinject
// +build wireinject

package server

import (
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	"github.com/begonia-org/begonia/internal/service"

	"github.com/google/wire"
	"github.com/spark-lence/tiga"
)

func New(config *tiga.Configuration, log logger.Logger, endpoint string) *transport.GatewayServer {

	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, service.ProviderSet, ProviderSet))

}
