package pkg

import (
	"github.com/google/wire"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"github.com/wetrycode/begonia/internal/pkg/crypto"
	"github.com/wetrycode/begonia/internal/pkg/middlerware"
)

var ProviderSet = wire.NewSet(config.NewConfig, middlerware.NewAPIVildator, crypto.NewUsersAuth)
