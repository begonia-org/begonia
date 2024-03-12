//go:build wireinject
// +build wireinject

package server

import (
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/service"
	dp "github.com/begonia-org/dynamic-proto"
	golayeredbloom "github.com/begonia-org/go-layered-bloom"
	"github.com/sirupsen/logrus"

	"github.com/google/wire"
	"github.com/spark-lence/tiga"
)

func New(config *tiga.Configuration, log *logrus.Logger, endpoint string, name golayeredbloom.ConsumerName) *dp.GatewayServer {

	panic(wire.Build(biz.ProviderSet, pkg.ProviderSet, data.ProviderSet, service.ProviderSet, ProviderSet))

}
