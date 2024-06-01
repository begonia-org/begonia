//go:build wireinject
// +build wireinject

package data

import (
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/biz/endpoint"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/google/wire"
	"github.com/spark-lence/tiga"
)

func NewAppRepo(cfg *tiga.Configuration, log logger.Logger) biz.AppRepo {
	panic(wire.Build(ProviderSet, config.NewConfig))
	// return &appRepoImpl{data: data, curd: curd, local: local, cfg: cfg}
}

func NewEndpointRepo(cfg *tiga.Configuration, log logger.Logger) endpoint.EndpointRepo {
	panic(wire.Build(ProviderSet, config.NewConfig))
}
func NewAuthzRepo(cfg *tiga.Configuration, log logger.Logger) biz.AuthzRepo {
	panic(wire.Build(ProviderSet, config.NewConfig))
}
func NewUserRepo(cfg *tiga.Configuration, log logger.Logger) biz.UserRepo {
	panic(wire.Build(ProviderSet, config.NewConfig))
}

func NewLayered(cfg *tiga.Configuration, log logger.Logger) *LayeredCache {
	panic(wire.Build(ProviderSet, config.NewConfig))
}

func NewOperator(cfg *tiga.Configuration, log logger.Logger) biz.DataOperatorRepo {
	panic(wire.Build(ProviderSet, config.NewConfig))
}
func NewDataRepo(cfg *tiga.Configuration, log logger.Logger) *Data {
	panic(wire.Build(ProviderSet))
}
func NewLocker(cfg *tiga.Configuration, log logger.Logger, key string, ttl time.Duration, retry int) biz.DataLock {
	panic(wire.Build(ProviderSet))
}
