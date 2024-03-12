//go:build wireinject
// +build wireinject

package main

import (
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/migrate"

	"github.com/google/wire"
	"github.com/spark-lence/tiga"
)


func initOperatorApp(config *tiga.Configuration) *migrate.InitOperator {

	panic(wire.Build(data.ProviderSet, pkg.ProviderSet))

}
