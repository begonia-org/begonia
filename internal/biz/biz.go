package biz

import (
	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/biz/endpoint"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewAuthzUsecase,
	NewUserUsecase,
	NewAccessKeyAuth,
	file.NewFileUsecase,
	endpoint.NewEndpointUsecase,
	NewAppUsecase,
	endpoint.NewWatcher,
	NewDataOperatorUsecase)
