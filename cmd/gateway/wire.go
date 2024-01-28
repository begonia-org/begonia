//go:build wireinject
// +build wireinject

package main

import (
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/server"
	"github.com/begonia-org/begonia/internal/service"
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
	"github.com/spark-lence/tiga"
)

func initApp(config *tiga.Configuration, log *logrus.Logger, endpoint string) *server.GatewayServer {

	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, service.ProviderSet, server.ProviderSet))

}
