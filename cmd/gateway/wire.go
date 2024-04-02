//go:build wireinject
// +build wireinject

package main

import (
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/migrate"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/daemon"
	"github.com/begonia-org/begonia/internal/service"
	"github.com/begonia-org/begonia/internal/server"

	"github.com/google/wire"
	"github.com/sirupsen/logrus"

	"github.com/spark-lence/tiga"
)


func initOperatorApp(config *tiga.Configuration) *migrate.InitOperator {

	panic(wire.Build(data.ProviderSet, pkg.ProviderSet))

}

func New(config *tiga.Configuration, log *logrus.Logger, endpoint string) GatewayWorker {

	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, service.ProviderSet, daemon.ProviderSet,server.ProviderSet,NewGatewayWorkerImpl))

}