package biz

import (
	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/biz/gateway"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewAuthzUsecase, 
	NewUserUsecase,
	NewAccessKeyAuth,
	file.NewFileUsecase, 
	gateway.NewEndpointUsecase, 
	NewAppUsecase,
	gateway.NewWatcher,
	NewDataOperatorUsecase)
