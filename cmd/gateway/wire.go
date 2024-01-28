//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
	"github.com/spark-lence/tiga"
	"github.com/wetrycode/begonia/internal/biz"
	"github.com/wetrycode/begonia/internal/data"
	"github.com/wetrycode/begonia/internal/pkg"
	"github.com/wetrycode/begonia/internal/server"
	"github.com/wetrycode/begonia/internal/service"
)

func initApp(config *tiga.Configuration, log *logrus.Logger, endpoint string) *server.GatewayServer {

	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, service.ProviderSet, server.ProviderSet))

}
